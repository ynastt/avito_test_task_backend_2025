package repository

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"ynastt/avito_test_task_backend_2025/internal/domain"
	"ynastt/avito_test_task_backend_2025/pkg/database"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Upsert(ctx context.Context, user domain.TeamMember, teamName string) error {
	conn := r.db.Conn(ctx)

	_, err := conn.ExecContext(ctx, `
        INSERT INTO users (user_id, username, team_name, is_active)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id) DO UPDATE
        SET username = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active,
            updated_at = NOW()
    `, user.UserID, user.Username, teamName, user.IsActive)

	if err != nil {
		return fmt.Errorf("failed to upsert user %s: %w", user.UserID, err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	conn := r.db.Conn(ctx)

	var user domain.User
	err := conn.QueryRowContext(ctx, `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err != nil {
		return nil, HandleNoRowsError(err)
	}

	return &user, nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	conn := r.db.Conn(ctx)

	var user domain.User
	err := conn.QueryRowContext(ctx, `
		UPDATE users
		SET is_active = $1, updated_at = NOW()
		WHERE user_id = $2
		RETURNING user_id, username, team_name, is_active
	`, isActive, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err != nil {
		return nil, HandleNoRowsError(err)
	}

	return &user, nil
}

func (r *UserRepository) GetByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	conn := r.db.Conn(ctx)

	rows, err := conn.QueryContext(ctx, `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
	`, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *UserRepository) GetActiveUsersByTeam(ctx context.Context, teamName string, excludeUserIDs []string) ([]domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = TRUE
	`

	var args []interface{}
	args = append(args, teamName)

	if len(excludeUserIDs) > 0 {
		query += " AND NOT (user_id = ANY($2))"
		args = append(args, pq.Array(excludeUserIDs))
	}

	conn := r.db.Conn(ctx)
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
