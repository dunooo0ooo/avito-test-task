package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/dunooo0ooo/avito-test-task/internal/team/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Create(ctx context.Context, t *domain.Team) error {
	const query = `
		INSERT INTO teams (team_name)
		VALUES (@name)
	`

	args := pgx.NamedArgs{
		"name": t.TeamName,
	}

	_, err := r.pool.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: %w", domain.ErrTeamAlreadyExists, err)
		}
		return fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return nil
}

func (r *Repository) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	const teamQuery = `
		SELECT team_name
		FROM teams
		WHERE team_name = @name
	`

	if err := r.pool.QueryRow(ctx, teamQuery, pgx.NamedArgs{"name": name}).Scan(new(string)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", domain.ErrTeamNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	const membersQuery = `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = @name
		ORDER BY user_id
	`

	rows, err := r.pool.Query(ctx, membersQuery, pgx.NamedArgs{"name": name})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer rows.Close()

	var members []domain.TeamMember

	for rows.Next() {
		var m domain.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	return &domain.Team{
		TeamName: name,
		Members:  members,
	}, nil
}

func (r *Repository) List(ctx context.Context) ([]*domain.Team, error) {
	const query = `
		SELECT
			t.team_name,
			u.user_id,
			u.username,
			u.is_active
		FROM teams t
		LEFT JOIN users u
			ON u.team_name = t.team_name
		ORDER BY t.team_name, u.user_id
	`

	rows, err := r.pool.Query(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}
	defer rows.Close()

	teamsMap := make(map[string]*domain.Team)

	for rows.Next() {
		var (
			teamName string
			userID   *string
			username *string
			isActive *bool
		)

		if err := rows.Scan(&teamName, &userID, &username, &isActive); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
		}

		team, ok := teamsMap[teamName]
		if !ok {
			team = &domain.Team{
				TeamName: teamName,
				Members:  nil,
			}
			teamsMap[teamName] = team
		}

		if userID != nil {
			team.Members = append(team.Members, domain.TeamMember{
				UserID:   *userID,
				Username: *username,
				IsActive: *isActive,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternalDatabase, err)
	}

	result := make([]*domain.Team, 0, len(teamsMap))
	for _, t := range teamsMap {
		result = append(result, t)
	}

	return result, nil
}
