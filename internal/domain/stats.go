package domain

type StatsResponse struct {
	TotalTeams      int64                 `json:"total_teams"`
	TotalUsers      int64                 `json:"total_users"`
	TotalPRs        int64                 `json:"total_pull_requests"`
	OpenPRs         int64                 `json:"open_pull_requests"`
	MergedPRs       int64                 `json:"merged_pull_requests"`
	ActiveUsers     int64                 `json:"active_users"`
	InactiveUsers   int64                 `json:"inactive_users"`
	UserAssignments []UserAssignmentStats `json:"user_assignments,omitempty"`
	PRAssignments   []PRAssignmentStats   `json:"pr_assignments,omitempty"`
}

type UserAssignmentStats struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	PRCount  int64  `json:"pr_count"`
	IsActive bool   `json:"is_active"`
}

type PRAssignmentStats struct {
	PRID      string `json:"pull_request_id"`
	PRName    string `json:"pull_request_name"`
	AuthorID  string `json:"author_id"`
	Status    string `json:"status"`
	Reviewers int    `json:"reviewers_count"`
}
