package service

import (
	pr "github.com/ynastt/avito_test_task_backend_2025/internal/service/pullrequest"
	"github.com/ynastt/avito_test_task_backend_2025/internal/service/team"
	"github.com/ynastt/avito_test_task_backend_2025/internal/service/user"
)

type Services struct {
	TeamService        *team.TeamService
	UserService        *user.UserService
	PullRequestService *pr.PullRequestService
	StatsService       *StatsService
}
