package pullrequest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ynastt/avito_test_task_backend_2025/internal/domain"
	"github.com/ynastt/avito_test_task_backend_2025/internal/repository"
	"github.com/ynastt/avito_test_task_backend_2025/internal/service/reviewers"
	"github.com/ynastt/avito_test_task_backend_2025/pkg/database"
)

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr domain.CreatePRRequest) (time.Time, error)
	Exists(ctx context.Context, prID string) (bool, error)
	AssignReviewer(ctx context.Context, reviewerID, prID string) error
	RemoveReviewer(ctx context.Context, reviewerID, prID string) error
	GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetOpenPullRequestsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	MergePullRequest(ctx context.Context, prID string) error
	IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error)
}

type UserRepository interface {
	GetActiveUsersByTeam(ctx context.Context, teamName string, excludeUserIDs []string) ([]domain.User, error)
	GetByID(ctx context.Context, userID string) (*domain.User, error)
}

type PullRequestService struct {
	prRepo    PullRequestRepository
	userRepo  UserRepository
	txManager database.TransactionManagerInterface
	lg        *slog.Logger
}

func NewPullRequestService(prRepo PullRequestRepository,
	userRepo UserRepository,
	txManager database.TransactionManagerInterface,
	lg *slog.Logger) *PullRequestService {
	return &PullRequestService{
		prRepo:    prRepo,
		userRepo:  userRepo,
		txManager: txManager,
		lg:        lg,
	}
}

func (s *PullRequestService) CreatePullRequest(ctx context.Context, prReqInfo domain.CreatePRRequest) (*domain.PullRequest, error) {
	log := s.lg.With(
		slog.String("create PR, pr_id", prReqInfo.ID),
		slog.String("author_id", prReqInfo.AuthorID),
	)

	var pr *domain.PullRequest

	// получим автора PR, проверим существует ли он, если да - определим название команды
	author, err := s.getAuthor(ctx, prReqInfo.AuthorID)
	if err != nil {
		return nil, err
	}
	log.Info("found author team", slog.String("team_name", author.TeamName))

	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		// проверка, существует ли уже PR
		exists, err := s.prRepo.Exists(txCtx, prReqInfo.ID)
		if err != nil {
			return fmt.Errorf("failed to check PR existence: %w", err)
		}
		if exists {
			return domain.ErrPRExists
		}

		// найдем пользователей из той же команды (кроме самого автора),
		// кто с активным статусом, и кого можно рассмотреть в качестве ревьюеров
		candidates, err := s.getPossibleReviewers(txCtx, author.TeamName, []string{prReqInfo.AuthorID})
		if err != nil {
			return err
		}
		log.Info("found reviewer candidates", slog.Int("count", len(candidates)))

		// берем до двух случайных ревьюеров
		reviewers := reviewers.ChooseRandomReviewers(candidates, 2)
		reviewerIDs := make([]string, len(reviewers))
		for i, r := range reviewers {
			reviewerIDs[i] = r.UserID
		}
		log.Info("selected PR reviewers", slog.Any("reviewer_ids", reviewerIDs))

		// создаем PR
		_, err = s.prRepo.CreatePullRequest(txCtx, prReqInfo)
		if err != nil {
			return fmt.Errorf("failed to create PR: %w", err)
		}

		// назначем ревьюеров на PR
		for _, reviewerID := range reviewerIDs {
			if err := s.prRepo.AssignReviewer(txCtx, reviewerID, prReqInfo.ID); err != nil {
				return fmt.Errorf("failed to assign PR reviewer %s: %w", reviewerID, err)
			}
		}

		// получаем созданный PR
		createdPR, err := s.prRepo.GetPullRequestByID(txCtx, prReqInfo.ID)
		if err != nil {
			return fmt.Errorf("failed to get created PR: %w", err)
		}
		pr = createdPR

		return nil
	})

	if err != nil {
		log.Error("failed to create PR", slog.Any("error", err))
		return nil, err
	}
	log.Info("PR created")
	return pr, nil
}

func (s *PullRequestService) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
	log := s.lg.With(
		slog.String("merge PR, pr_id", prID),
	)

	var pr *domain.PullRequest
	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		// проверяем, что PR существует
		exists, err := s.prRepo.Exists(txCtx, prID)
		if err != nil {
			return fmt.Errorf("failed to check PR existence: %w", err)
		}
		if !exists {
			return domain.ErrPRNotFound
		}

		// выполняем merge
		if err := s.prRepo.MergePullRequest(txCtx, prID); err != nil {
			return fmt.Errorf("failed to merge PR: %w", err)
		}

		// получаем PR с обновленным статусом (MERGED)
		mergedPR, err := s.prRepo.GetPullRequestByID(txCtx, prID)
		if err != nil {
			return fmt.Errorf("failed to get merged PR: %w", err)
		}
		pr = mergedPR

		return nil
	})

	if err != nil {
		return nil, err
	}
	log.Info("PR merged")
	return pr, nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID, prevReviewerID string) (*domain.PullRequest, string, error) {
	log := s.lg.With(
		slog.String("reassign PR, pr_id", prID),
		slog.String("old_user_id", prevReviewerID),
	)

	var updPR *domain.PullRequest
	var newReviewerID string

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		// проверяем, что PR существует
		pr, err := s.prRepo.GetPullRequestByID(txCtx, prID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return domain.ErrPRNotFound
			}
			return fmt.Errorf("failed to get PR: %w", err)
		}

		// проверяем, что статус PR - НЕ MERGED
		if pr.IsPRMerged() {
			log.Error("cannot reassign on merged PR")
			return domain.ErrPRMerged
		}

		// проверяем, что старый ревьюер назначен на PR
		isAssigned, err := s.prRepo.IsReviewerAssigned(txCtx, prID, prevReviewerID)
		if err != nil {
			return fmt.Errorf("failed to check reviewer PR assignment: %w", err)
		}
		if !isAssigned {
			log.Error("reviewer not assigned")
			return domain.ErrNotAssigned
		}

		// находим команду старого ревьюера
		prevReviewer, err := s.userRepo.GetByID(ctx, prevReviewerID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return domain.ErrUserNotFound
			}
			return fmt.Errorf("failed to get old PR reviewer: %w", err)
		}
		log.Info("found old reviewer team", slog.String("team_name", prevReviewer.TeamName))

		// ищем возхможных новых ревьюеров из этой команды (активные, не автор, не текущие ревьюверы)
		excludedUsersIDs := []string{pr.AuthorID}
		excludedUsersIDs = append(excludedUsersIDs, pr.AssignedReviewers...)

		candidates, err := s.getPossibleReviewers(txCtx, prevReviewer.TeamName, excludedUsersIDs)
		if err != nil {
			return err
		}
		log.Info("found reviewer candidates", slog.Int("count", len(candidates)))

		// проверка на ошибку NO_CANDIDATE
		if len(candidates) == 0 {
			return domain.ErrNoCandidate
		}

		//выбираем 1 случайного ревьюера и обновляем
		reviewer, err := reviewers.ChooseRandomReviewer(candidates)
		if err != nil {
			return domain.ErrNoCandidate
		}

		reviewerID := reviewer.UserID
		log.Info("selected PR reviewer", slog.String("reviewer_id", reviewerID))

		if err := s.prRepo.RemoveReviewer(txCtx, prevReviewerID, prID); err != nil {
			return fmt.Errorf("failed to remove old reviewer: %w", err)
		}

		if err := s.prRepo.AssignReviewer(txCtx, reviewerID, prID); err != nil {
			return fmt.Errorf("failed to assign new reviewer: %w", err)
		}

		// получаем PR с обновленным ревьюером
		pr, err = s.prRepo.GetPullRequestByID(txCtx, prID)
		if err != nil {
			return fmt.Errorf("failed to get updated PR: %w", err)
		}
		updPR = pr
		newReviewerID = reviewerID

		return nil
	})

	if err != nil {
		return nil, "", err
	}
	log.Info("PR reassigned")
	return updPR, newReviewerID, nil
}

func (s *PullRequestService) getAuthor(ctx context.Context, authorID string) (*domain.User, error) {
	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get author of PR: %w", err)
	}

	return author, nil
}

func (s *PullRequestService) getPossibleReviewers(ctx context.Context, teamName string, excludeUserIDs []string) ([]domain.User, error) {
	candidates, err := s.userRepo.GetActiveUsersByTeam(ctx, teamName, excludeUserIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get active team members: %w", err)
	}

	return candidates, nil
}
