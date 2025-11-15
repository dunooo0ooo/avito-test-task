package http

import "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"

type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveResponse struct {
	User UserDTO `json:"user"`
}

type GetReviewsResponse struct {
	UserID       string                    `json:"user_id"`
	PullRequests []domain.PullRequestShort `json:"pull_requests"`
}
