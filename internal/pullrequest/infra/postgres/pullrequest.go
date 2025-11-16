package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewPullRequestRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	const query = `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
		VALUES (@id, @name, @auth, @status)
	`

	status := pr.Status
	if status == "" {
		status = domain.PRStatusOpen
	}

	args := pgx.NamedArgs{
		"id":     pr.PullRequestID,
		"name":   pr.PullRequestName,
		"auth":   pr.AuthorID,
		"status": status,
	}

	_, err = r.pool.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: %w", domain.ErrPullRequestAlreadyExists, err)
		}
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	const query = `
		SELECT
			p.pull_request_id,
			p.pull_request_name,
			p.author_id,
			p.status,
			p.created_at,
			p.merged_at,
			COALESCE(
				array_agg(rw.reviewer_id ORDER BY rw.reviewer_id)
					FILTER (WHERE rw.reviewer_id IS NOT NULL),
				'{}'::text[]
			) AS reviewers
		FROM pull_requests p
		LEFT JOIN pr_reviewers rw
			ON rw.pull_request_id = p.pull_request_id
		WHERE p.pull_request_id = @id
		GROUP BY
			p.pull_request_id,
			p.pull_request_name,
			p.author_id,
			p.status,
			p.created_at,
			p.merged_at
	`

	args := pgx.NamedArgs{"id": id}

	var (
		pr        domain.PullRequest
		status    string
		reviewers []string
		createdAt *time.Time
		mergedAt  *time.Time
	)

	err := r.pool.QueryRow(ctx, query, args).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&status,
		&createdAt,
		&mergedAt,
		&reviewers,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", domain.ErrPullRequestNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	pr.Status = domain.PRStatus(status)
	pr.CreatedAt = createdAt
	pr.MergedAt = mergedAt
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status domain.PRStatus, mergedAt *time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	const query = `
		UPDATE pull_requests
		SET status   = @status,
		    merged_at = @merged_at
		WHERE pull_request_id = @id
	`

	args := pgx.NamedArgs{
		"id":        id,
		"status":    status,
		"merged_at": mergedAt,
	}

	cmd, err := r.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("%w: %w", domain.ErrPullRequestNotFound, pgx.ErrNoRows)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return nil
}

func (r *Repository) SetReviewers(ctx context.Context, id string, reviewerIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	const deleteQuery = `
		DELETE FROM pr_reviewers
		WHERE pull_request_id = @id
	`

	if _, err := tx.Exec(ctx, deleteQuery, pgx.NamedArgs{"id": id}); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	if len(reviewerIDs) > 0 {
		const insertQuery = `
			INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
			VALUES (@pr_id, @rev_id)
		`

		for _, rid := range reviewerIDs {
			args := pgx.NamedArgs{
				"pr_id":  id,
				"rev_id": rid,
			}
			if _, err := tx.Exec(ctx, insertQuery, args); err != nil {
				return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return nil
}

func (r *Repository) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	const query = `
		SELECT
			p.pull_request_id,
			p.pull_request_name,
			p.author_id,
			p.status
		FROM pr_reviewers rw
		JOIN pull_requests p
			ON p.pull_request_id = rw.pull_request_id
		WHERE rw.reviewer_id = @rid
		ORDER BY p.created_at DESC
	`

	args := pgx.NamedArgs{"rid": reviewerID}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer rows.Close()

	var res []domain.PullRequestShort

	for rows.Next() {
		var (
			pr     domain.PullRequestShort
			status string
		)

		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&status,
		); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
		}

		pr.Status = domain.PRStatus(status)
		res = append(res, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return res, nil
}

func (r *Repository) ListOpenByReviewers(ctx context.Context, reviewerIDs []string) ([]domain.PullRequestShort, error) {
	if len(reviewerIDs) == 0 {
		return nil, nil
	}

	const query = `
		SELECT DISTINCT
			p.pull_request_id,
			p.pull_request_name,
			p.author_id,
			p.status
		FROM pull_requests p
		JOIN pr_reviewers rw
			ON rw.pull_request_id = p.pull_request_id
		WHERE p.status = 'OPEN'
		  AND rw.reviewer_id = ANY(@reviewer_ids)
	`

	args := pgx.NamedArgs{
		"reviewer_ids": reviewerIDs,
	}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer rows.Close()

	var res []domain.PullRequestShort

	for rows.Next() {
		var (
			pr     domain.PullRequestShort
			status string
		)

		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&status,
		); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
		}

		pr.Status = domain.PRStatus(status)
		res = append(res, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return res, nil
}

func (r *Repository) CountByReviewer(ctx context.Context) (map[string]string, error) {
	const query = `
		SELECT reviewer_id, COUNT(*) AS cnt
		FROM pr_reviewers
		GROUP BY reviewer_id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer rows.Close()

	result := make(map[string]string)

	for rows.Next() {
		var (
			reviewerID string
			count      int64
		)
		if err := rows.Scan(&reviewerID, &count); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
		}
		result[reviewerID] = fmt.Sprintf("%d", count)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return result, nil
}
