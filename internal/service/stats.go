package service

import (
	"context"
	"log/slog"

	"github.com/ynastt/avito_test_task_backend_2025/internal/domain"
	"github.com/ynastt/avito_test_task_backend_2025/internal/repository"
)

type StatsService struct {
	statsRepo repository.StatsRepository
	logger    *slog.Logger
}

func NewStatsService(statsRepo repository.StatsRepository, logger *slog.Logger) *StatsService {
	return &StatsService{
		statsRepo: statsRepo,
		logger:    logger,
	}
}

func (s *StatsService) GetStats(ctx context.Context, includeDetails bool) (*domain.StatsResponse, error) {
	// Получаем основную статистику
	stats, err := s.statsRepo.GetTotalStats(ctx)
	if err != nil {
		return nil, err
	}

	// Детализированная статистика при необходимости
	if includeDetails {
		userStats, err := s.statsRepo.GetUserAssignmentStats(ctx)
		if err != nil {
			s.logger.Warn("failed to get user assignment stats", "error", err)
		} else {
			stats.UserAssignments = userStats
		}

		prStats, err := s.statsRepo.GetPRAssignmentStats(ctx)
		if err != nil {
			s.logger.Warn("failed to get PR assignment stats", "error", err)
		} else {
			stats.PRAssignments = prStats
		}
	}

	s.logger.Info("stats retrieved",
		"teams", stats.TotalTeams,
		"users", stats.TotalUsers,
		"prs", stats.TotalPRs,
		"open_prs", stats.OpenPRs)

	return stats, nil
}
