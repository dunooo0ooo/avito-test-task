package http

import (
	"github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	"time"
)

type PullRequestDTO struct {
	PullRequestID     string          `json:"pull_request_id"`
	PullRequestName   string          `json:"pull_request_name"`
	AuthorID          string          `json:"author_id"`
	Status            domain.PRStatus `json:"status"`
	AssignedReviewers []string        `json:"assigned_reviewers"`
}

type CreateResponse struct {
	PullRequestDTO PullRequestDTO `json:"pr"`
}

type MergedPullRequestDTO struct {
	PullRequestID     string          `json:"pull_request_id"`
	PullRequestName   string          `json:"pull_request_name"`
	AuthorID          string          `json:"author_id"`
	Status            domain.PRStatus `json:"status"`
	AssignedReviewers []string        `json:"assigned_reviewers"`
	MergedAt          *time.Time      `json:"mergedAt"`
}

type MergeResponse struct {
	MergedPullRequestDTO MergedPullRequestDTO `json:"pr"`
}

type ReassignResponse struct {
	PullRequestDTO   PullRequestDTO `json:"pr"`
	ReplacedReviewer string         `json:"replaced_by"`
}
