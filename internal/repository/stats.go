package repository

import (
	"context"

	"github.com/ynastt/avito_test_task_backend_2025/internal/domain"
	"github.com/ynastt/avito_test_task_backend_2025/pkg/database"
)

type StatsRepository interface {
    GetTotalStats(ctx context.Context) (*domain.StatsResponse, error)
    GetUserAssignmentStats(ctx context.Context) ([]domain.UserAssignmentStats, error)
    GetPRAssignmentStats(ctx context.Context) ([]domain.PRAssignmentStats, error)
}

type statsRepository struct {
    db *database.DB
}

func NewStatsRepository(db *database.DB) StatsRepository {
    return &statsRepository{db: db}
}

func (r *statsRepository) GetTotalStats(ctx context.Context) (*domain.StatsResponse, error) {
    conn := r.db.Conn(ctx)
    
    var stats domain.StatsResponse
    
    // Основная статистика
    err := conn.QueryRowContext(ctx, `
        SELECT 
            (SELECT COUNT(*) FROM teams) as total_teams,
            (SELECT COUNT(*) FROM users) as total_users,
            (SELECT COUNT(*) FROM pull_requests) as total_prs,
            (SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN') as open_prs,
            (SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED') as merged_prs,
            (SELECT COUNT(*) FROM users WHERE is_active = true) as active_users,
            (SELECT COUNT(*) FROM users WHERE is_active = false) as inactive_users
    `).Scan(
        &stats.TotalTeams,
        &stats.TotalUsers,
        &stats.TotalPRs,
        &stats.OpenPRs,
        &stats.MergedPRs,
        &stats.ActiveUsers,
        &stats.InactiveUsers,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &stats, nil
}

func (r *statsRepository) GetUserAssignmentStats(ctx context.Context) ([]domain.UserAssignmentStats, error) {
    conn := r.db.Conn(ctx)
    
    rows, err := conn.QueryContext(ctx, `
        SELECT 
            u.user_id,
            u.username,
            u.team_name,
            u.is_active,
            COUNT(pr.pull_request_id) as pr_count
        FROM users u
        LEFT JOIN pull_requests pr ON u.user_id = ANY(pr.assigned_reviewers)
        GROUP BY u.user_id, u.username, u.team_name, u.is_active
        ORDER BY pr_count DESC, u.user_id
    `)
    
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var stats []domain.UserAssignmentStats
    for rows.Next() {
        var s domain.UserAssignmentStats
        err := rows.Scan(&s.UserID, &s.Username, &s.TeamName, &s.IsActive, &s.PRCount)
        if err != nil {
            return nil, err
        }
        stats = append(stats, s)
    }
    
    return stats, rows.Err()
}

func (r *statsRepository) GetPRAssignmentStats(ctx context.Context) ([]domain.PRAssignmentStats, error) {
    conn := r.db.Conn(ctx)
    
    rows, err := conn.QueryContext(ctx, `
        SELECT 
            pull_request_id,
            pull_request_name,
            author_id,
            status,
            array_length(assigned_reviewers, 1) as reviewers_count
        FROM pull_requests
        ORDER BY created_at DESC
    `)
    
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var stats []domain.PRAssignmentStats
    for rows.Next() {
        var s domain.PRAssignmentStats
        var reviewersCount *int
        err := rows.Scan(&s.PRID, &s.PRName, &s.AuthorID, &s.Status, &reviewersCount)
        if err != nil {
            return nil, err
        }
        
        if reviewersCount != nil {
            s.Reviewers = *reviewersCount
        } else {
            s.Reviewers = 0
        }
        
        stats = append(stats, s)
    }
    
    return stats, rows.Err()
}