package domain

import "time"

type PRReassignStatus string

const (
	// дективированный пользователь заменен на нового ревьюера
	ReviewerReplaced PRReassignStatus = "REPLACED"

	// нет замены для деактивированного пользователя
	ReviewerRemoved PRReassignStatus = "REMOVED_NO_REPLACEMENT"
)

type User struct {
	UserID    string     `json:"user_id"`
	Username  string     `json:"username"`
	TeamName  string     `json:"team_name"`
	IsActive  bool       `json:"is_active"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type SetActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// для метода массовой деактивации
type BulkDeactivateRequest struct {
	UserIDs []string `json:"user_ids"`
}

type BulkDeactivateResponse struct {
	DeactivatedUserIDs []string  `json:"deactivated_user_ids"`
	PRsInfo            []PRsInfo `json:"pull_requests_info"`
	Errors             []string  `json:"errors,omitempty"`
}

type PRsInfo struct {
	PRID           string `json:"pr_id"`
	OldReviewerID  string `json:"old_reviewer_id"`
	NewReviewerID  string `json:"new_reviewer_id,omitempty"`
	ReassignStatus string `json:"reassign_status"`
}
