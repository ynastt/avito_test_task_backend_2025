package domain

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrTeamExists   = errors.New("team_name already exists")
	ErrPRExists     = errors.New("PR id already exists")
	ErrPRMerged     = errors.New("cannot reassign on merged PR")
	ErrNotAssigned  = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate  = errors.New("no active replacement candidate in team")
	ErrNotFound     = errors.New("resource not found")
	ErrTeamNotFound = errors.New("team not found")
	ErrUserNotFound = errors.New("user not found")
	ErrPRNotFound   = errors.New("PR not found")
	ErrEmptyUserIDs = errors.New("user_ids cannot be empty")
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
