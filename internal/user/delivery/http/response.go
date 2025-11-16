package http

import "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/delivery/http"

type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type PullRequestShortDTO struct {
	PullRequestID   string        `json:"pull_request_id"`
	PullRequestName string        `json:"pull_request_name"`
	AuthorID        string        `json:"author_id"`
	Status          http.PRStatus `json:"status"`
}

type SetIsActiveResponse struct {
	User UserDTO `json:"user"`
}

type GetReviewsResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}
type DeactivateResponse struct {
	TeamName    string   `json:"team_name"`
	Deactivated []string `json:"deactivated"`
}
