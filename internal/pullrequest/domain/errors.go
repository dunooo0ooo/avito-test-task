package domain

import "errors"

var (
	ErrPullRequestNotFound      = errors.New("pull request not found")
	ErrPullRequestAlreadyExists = errors.New("pull request already exists")
	ErrInternalDatabase         = errors.New("pr: internal database error")
)

var (
	ErrReviewerNotAssigned = errors.New("reviewer not assigned")
	ErrPullRequestMerged   = errors.New("pull request merged")
	ErrNoCandidate         = errors.New("no candidate")
)
