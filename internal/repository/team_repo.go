package repository

import (
	"context"
	"fmt"

	"github.com/ynastt/avito_test_task_backend_2025/internal/domain"
	"github.com/ynastt/avito_test_task_backend_2025/pkg/database"
)

type TeamRepository struct {
	db *database.DB
}

func NewTeamRepository(db *database.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, teamName string) error {
	conn := r.db.Conn(ctx)

	_, err := conn.ExecContext(ctx, "INSERT INTO teams (team_name) VALUES ($1)", teamName)
	if err != nil {
		return fmt.Errorf("failed to insert team: %w", err)
	}

	return nil
}

func (r *TeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	conn := r.db.Conn(ctx)

	var exists bool
	err := conn.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}
	return exists, nil
}

func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	exists, err := r.Exists(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	conn := r.db.Conn(ctx)
	rows, err := conn.QueryContext(ctx, `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		`, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team members: %w", err)
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}
