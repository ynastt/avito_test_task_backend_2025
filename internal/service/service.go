package service

import (
	pr "ynastt/avito_test_task_backend_2025/internal/service/pullrequest"
	"ynastt/avito_test_task_backend_2025/internal/service/team"
	"ynastt/avito_test_task_backend_2025/internal/service/user"
)

type Services struct {
	TeamService        *team.TeamService
	UserService        *user.UserService
	PullRequestService *pr.PullRequestService
	StatsService       *StatsService
}
