package domain

import (
	"fmt"
	"time"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	ID       string   `json:"pull_request_id"`
	Name     string   `json:"pull_request_name"`
	AuthorID string   `json:"author_id"`
	Status   PRStatus `json:"status"`
}

type CreatePRRequest struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type MergePRRequest struct {
	ID string `json:"pull_request_id"`
}

type ReassignRequest struct {
	ID        string `json:"pull_request_id"`
	OldUserID string `json:"old_user_id"`
}

type ReassignResponse struct {
	PR         *PullRequest `json:"pr"`
	ReplacedBy string       `json:"replaced_by"`
}

type UserReviewsResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

func (s PRStatus) IsValid() bool {
	switch s {
	case PRStatusOpen, PRStatusMerged:
		return true
	default:
		return false
	}
}

func (pr *PullRequest) ValidateStatus() error {
	if !pr.Status.IsValid() {
		return fmt.Errorf("invalid PR status: %s", pr.Status)
	}
	return nil
}

func (pr *PullRequest) IsPRMerged() bool {
	return pr.Status == PRStatusMerged
}
