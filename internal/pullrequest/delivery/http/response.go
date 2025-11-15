package http

import (
	"time"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequestDTO struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            PRStatus `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type CreateResponse struct {
	PullRequestDTO PullRequestDTO `json:"pr"`
}

type MergedPullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	MergedAt          *time.Time `json:"mergedAt"`
}

type MergeResponse struct {
	MergedPullRequestDTO MergedPullRequestDTO `json:"pr"`
}

type ReassignResponse struct {
	PullRequestDTO   PullRequestDTO `json:"pr"`
	ReplacedReviewer string         `json:"replaced_by"`
}
