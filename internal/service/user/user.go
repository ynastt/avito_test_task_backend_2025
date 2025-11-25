package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"ynastt/avito_test_task_backend_2025/internal/domain"
	"ynastt/avito_test_task_backend_2025/internal/repository"
	"ynastt/avito_test_task_backend_2025/internal/service/reviewers"
	"ynastt/avito_test_task_backend_2025/pkg/database"
)

type UserRepository interface {
	GetActiveUsersByTeam(ctx context.Context, teamName string, excludeUserIDs []string) ([]domain.User, error)
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
}

type PullRequestRepository interface {
	GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetPullRequestsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	GetOpenPullRequestsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	RemoveReviewer(ctx context.Context, reviewerID, prID string) error
	AssignReviewer(ctx context.Context, reviewerID, prID string) error
}

type UserService struct {
	userRepo  UserRepository
	prRepo    PullRequestRepository
	txManager database.TransactionManagerInterface
	lg        *slog.Logger
}

func NewUserService(userRepo UserRepository,
	prRepo PullRequestRepository,
	txManager database.TransactionManagerInterface,
	lg *slog.Logger) *UserService {
	return &UserService{
		userRepo:  userRepo,
		prRepo:    prRepo,
		txManager: txManager,
		lg:        lg,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	if !isActive {
		user, _, err := s.deactivateUser(ctx, userID)
		return user, err
	}

	user, err := s.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to set user active status: %w", err)
	}

	s.lg.Info("user active status updated", slog.String("user_id", userID), slog.Bool("is_active", isActive))
	return user, nil
}

// метод массовой деактивации
func (s *UserService) BulkDeactivateUsers(ctx context.Context, userIDs []string) (*domain.BulkDeactivateResponse, error) {
	if len(userIDs) == 0 {
		return &domain.BulkDeactivateResponse{
			DeactivatedUserIDs: []string{},
			PRsInfo:            []domain.PRsInfo{},
		}, domain.ErrEmptyUserIDs
	}

	var (
		deactivatedUserIDs []string
		reassignedPRs      []domain.PRsInfo
		errors             []string
	)

	type userResult struct {
		userID string
		user   *domain.User
		prs    []domain.PRsInfo
		err    error
	}

	// обрабатываем пользователей в горутинах
	results := make(chan userResult, len(userIDs))
	var wg sync.WaitGroup

	for _, userID := range userIDs {
		wg.Add(1)
		go func(uid string) {
			defer wg.Done()

			user, prs, err := s.deactivateUser(ctx, uid)
			results <- userResult{
				userID: uid,
				user:   user,
				prs:    prs,
				err:    err,
			}
		}(userID)
	}

	// закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем результаты
	for result := range results {
		if result.err != nil {
			errors = append(errors, fmt.Sprintf("user %s: %v", result.userID, result.err))
			continue
		}

		if result.user != nil {
			deactivatedUserIDs = append(deactivatedUserIDs, result.user.UserID)
		}

		if len(result.prs) > 0 {
			reassignedPRs = append(reassignedPRs, result.prs...)
		}
	}

	// формируем ответ
	response := &domain.BulkDeactivateResponse{
		DeactivatedUserIDs: deactivatedUserIDs,
		PRsInfo:            reassignedPRs,
	}

	if len(errors) > 0 {
		response.Errors = errors
	}

	s.lg.Info("bulk deactivation completed",
		slog.Int("count deactivated users", len(deactivatedUserIDs)),
		slog.Int("errors", len(errors)))

	return response, nil
}

func (s *UserService) deactivateUser(ctx context.Context, userID string) (*domain.User, []domain.PRsInfo, error) {
	var user *domain.User
	var prs []domain.PRsInfo

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		oldUser, err := s.userRepo.GetByID(txCtx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return domain.ErrUserNotFound
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		if !oldUser.IsActive {
			user = oldUser
			return nil
		}

		openPRs, err := s.prRepo.GetOpenPullRequestsByReviewer(txCtx, userID)
		if err != nil {
			return fmt.Errorf("failed to get open PRs for reviewer: %w", err)
		}

		for _, prShort := range openPRs {
			newReviewerID, reassignStatus, err := s.replacePRReviewer(txCtx, prShort.ID, userID, oldUser.TeamName)
			if err != nil {
				return fmt.Errorf("failed to handle PR %s replacement: %w", prShort.ID, err)
			}

			prInfo := domain.PRsInfo{
				PRID:           prShort.ID,
				OldReviewerID:  userID,
				NewReviewerID:  newReviewerID,
				ReassignStatus: reassignStatus,
			}

			prs = append(prs, prInfo)
		}

		user, err = s.userRepo.SetIsActive(txCtx, userID, false)
		if err != nil {
			return fmt.Errorf("failed to deactivate user: %w", err)
		}

		s.lg.Info("user deactivated",
			slog.String("user_id", userID),
			slog.Int("prs_processed", len(openPRs)))

		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, prs, domain.ErrUserNotFound
		}
		return nil, prs, fmt.Errorf("failed to deactivate user: %w", err)
	}

	return user, prs, nil
}

func (s *UserService) replacePRReviewer(
	ctx context.Context,
	prID string,
	OldReviewerID string,
	teamName string,
) (string, string, error) {
	pr, err := s.prRepo.GetPullRequestByID(ctx, prID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get PR %s: %w", prID, err)
	}

	excludeIDs := []string{pr.AuthorID}
	excludeIDs = append(excludeIDs, pr.AssignedReviewers...)

	candidates, err := s.userRepo.GetActiveUsersByTeam(ctx, teamName, excludeIDs)
	if err != nil {
		s.lg.Warn("failed to get replacement candidates, removing reviewer",
			slog.String("pr_id", prID),
			slog.String("user_id", OldReviewerID),
			slog.Any("error", err))
		return "", string(domain.ReviewerRemoved), s.removeReviewer(ctx, OldReviewerID, prID)
	}

	if len(candidates) == 0 {
		s.lg.Info("no replacement candidates found, removing reviewer",
			slog.String("pr_id", prID),
			slog.String("user_id", OldReviewerID))
		return "", string(domain.ReviewerRemoved), s.removeReviewer(ctx, OldReviewerID, prID)
	}

	newReviewer, err := reviewers.ChooseRandomReviewer(candidates)
	if err != nil {
		s.lg.Warn("failed to select reviewer, removing",
			slog.String("pr_id", prID),
			slog.String("user_id", OldReviewerID))
		return "", string(domain.ReviewerRemoved), s.removeReviewer(ctx, OldReviewerID, prID)
	}

	if err := s.removeReviewer(ctx, OldReviewerID, prID); err != nil {
		return "", "", err
	}

	if err := s.assignReviewer(ctx, newReviewer.UserID, prID); err != nil {
		return "", "", err
	}

	s.lg.Info("reviewer reassigned during deactivation",
		slog.String("pr_id", prID),
		slog.String("old_user_id", OldReviewerID),
		slog.String("new_user_id", newReviewer.UserID))
	return newReviewer.UserID, string(domain.ReviewerReplaced), nil
}

func (s *UserService) removeReviewer(ctx context.Context, userID, prID string) error {
	if err := s.prRepo.RemoveReviewer(ctx, userID, prID); err != nil {
		return fmt.Errorf("failed to remove reviewer: %w", err)
	}

	s.lg.Info("removed not active reviewer from PR",
		slog.String("pr_id", prID),
		slog.String("user_id", userID))

	return nil
}

func (s *UserService) assignReviewer(ctx context.Context, userID, prID string) error {
	if err := s.prRepo.AssignReviewer(ctx, userID, prID); err != nil {
		return fmt.Errorf("failed to assign reviewer: %w", err)
	}

	s.lg.Info("assigned reviewer for PR",
		slog.String("pr_id", prID),
		slog.String("user_id", userID))

	return nil
}

func (s *UserService) GetUserReviewerPRs(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	prs, err := s.prRepo.GetPullRequestsByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get review PRs: %w", err)
	}

	s.lg.Info("retrieved review PRs", slog.String("user_id", userID), slog.Int("PR count", len(prs)))
	return prs, nil
}
