package team

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ynastt/avito_test_task_backend_2025/internal/domain"
	"github.com/ynastt/avito_test_task_backend_2025/internal/repository"
	"github.com/ynastt/avito_test_task_backend_2025/pkg/database"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, teamName string) error
	Exists(ctx context.Context, teamName string) (bool, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}

type UserRepository interface {
	Upsert(ctx context.Context, user domain.TeamMember, teamName string) error
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
}

type TeamService struct {
	teamRepo  TeamRepository
	userRepo  UserRepository
	txManager database.TransactionManagerInterface
	lg        *slog.Logger
}

func NewTeamService(teamRepo TeamRepository,
	userRepo UserRepository,
	txManager database.TransactionManagerInterface,
	lg *slog.Logger) *TeamService {
	return &TeamService{
		teamRepo:  teamRepo,
		userRepo:  userRepo,
		txManager: txManager,
		lg:        lg,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, team domain.Team) (*domain.Team, error) {
	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		exists, err := s.teamRepo.Exists(txCtx, team.TeamName)
		if err != nil {
			return fmt.Errorf("failed to check team existence: %w", err)
		}
		if exists {
			return domain.ErrTeamExists
		}

		if err := s.teamRepo.CreateTeam(txCtx, team.TeamName); err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}

		for _, member := range team.Members {
			if err := s.userRepo.Upsert(txCtx, member, team.TeamName); err != nil {
				return fmt.Errorf("failed to add team member %s: %w", member.UserID, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.lg.Info("team created", slog.String("team_name", team.TeamName), slog.Int("members_count", len(team.Members)))
	return &team, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team by team_name: %w", err)
	}

	return team, nil
}
