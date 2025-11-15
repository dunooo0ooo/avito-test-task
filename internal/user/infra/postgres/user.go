package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) AddTeamMembers(ctx context.Context, teamName string, members []domain.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer tx.Rollback(ctx)

	const query = `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES (@id, @username, @team_name, @is_active)
		ON CONFLICT (user_id) DO UPDATE
		SET
			username   = EXCLUDED.username,
			team_name  = EXCLUDED.team_name,
			is_active  = EXCLUDED.is_active,
			updated_at = NOW()
	`

	for _, m := range members {
		args := pgx.NamedArgs{
			"id":        m.UserID,
			"username":  m.Username,
			"team_name": teamName,
			"is_active": m.IsActive,
		}

		if _, err := r.pool.Exec(ctx, query, args); err != nil {
			return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const query = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = @id
	`

	args := pgx.NamedArgs{"id": id}

	var user domain.User
	err := r.pool.QueryRow(ctx, query, args).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", domain.ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return &user, nil
}

func (r *Repository) ListByTeam(ctx context.Context, teamName string) ([]*domain.User, error) {
	const query = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = @teamName
	`

	args := pgx.NamedArgs{"teamName": teamName}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer rows.Close()

	var users []*domain.User

	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
		); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return users, nil
}

func (r *Repository) UpdateActive(ctx context.Context, id string, active bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE users
		SET is_active = @active,
		    updated_at = NOW()
		WHERE user_id = @id
	`

	args := pgx.NamedArgs{
		"id":     id,
		"active": active,
	}

	cmd, err := r.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("%w: %w", domain.ErrUserNotFound, pgx.ErrNoRows)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return nil
}

func (r *Repository) DeactivateByTeam(ctx context.Context, teamName string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE users
		SET is_active = FALSE,
		    updated_at = NOW()
		WHERE team_name = @teamName
	`

	args := pgx.NamedArgs{"teamName": teamName}

	_, err = r.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return nil
}
