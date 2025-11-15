package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	userdomain "github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"github.com/dunooo0ooo/avito-test-task/pkg/httpcommon"
	"net/http"
)

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, id string, name string, authorID string) (*domain.PullRequest, error)
	MergePullRequest(ctx context.Context, id string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID string, oldReviewerID string) (*domain.PullRequest, string, error)
}

type PullRequestHandler struct {
	prs PullRequestService
}

func NewPullRequestHandler(prs PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{
		prs: prs,
	}
}

func (h *PullRequestHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /pullRequest/create", h.Create)
	mux.HandleFunc("POST /pullRequest/merge", h.Merge)
	mux.HandleFunc("POST /pullRequest/reassign", h.Reassign)
}

func (h *PullRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	pr, err := h.prs.CreatePullRequest(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPullRequestAlreadyExists):
			httpcommon.JSONError(w, http.StatusConflict, "PR_EXISTS", "pull request already exists")
		case errors.Is(err, userdomain.ErrUserNotFound):
			httpcommon.JSONError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		default:
			httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return
	}

	resp := CreateResponse{
		PullRequestDTO: PullRequestDTO{
			PullRequestID:     pr.PullRequestID,
			PullRequestName:   pr.PullRequestName,
			AuthorID:          pr.AuthorID,
			Status:            PRStatus(pr.Status),
			AssignedReviewers: pr.AssignedReviewers,
		},
	}

	httpcommon.JSONResponse(w, http.StatusCreated, resp)
}

func (h *PullRequestHandler) Merge(w http.ResponseWriter, r *http.Request) {
	var req MergeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	pr, err := h.prs.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		switch {
		case errors.Is(err, userdomain.ErrUserNotFound):
			httpcommon.JSONError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		case errors.Is(err, domain.ErrPullRequestMerged):
			httpcommon.JSONError(w, http.StatusNotFound, "PR_MERGED", "pull request already merged")
		default:
			httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return
	}

	resp := MergeResponse{
		MergedPullRequestDTO: MergedPullRequestDTO{
			PullRequestID:     pr.PullRequestID,
			PullRequestName:   pr.PullRequestName,
			AuthorID:          pr.AuthorID,
			Status:            PRStatus(pr.Status),
			AssignedReviewers: pr.AssignedReviewers,
			MergedAt:          pr.MergedAt,
		},
	}

	httpcommon.JSONResponse(w, http.StatusOK, resp)
}
func (h *PullRequestHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req ReassignRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	pr, id, err := h.prs.ReassignReviewer(r.Context(), req.PullRequestID, req.OldReviewerID)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPullRequestNotFound),
			errors.Is(err, userdomain.ErrUserNotFound):
			httpcommon.JSONError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		case errors.Is(err, domain.ErrPullRequestMerged):
			httpcommon.JSONError(w, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
		case errors.Is(err, domain.ErrReviewerNotAssigned):
			httpcommon.JSONError(w, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
		case errors.Is(err, domain.ErrNoCandidate):
			httpcommon.JSONError(w, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
		default:
			httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return
	}

	resp := ReassignResponse{
		PullRequestDTO: PullRequestDTO{
			PullRequestID:     pr.PullRequestID,
			PullRequestName:   pr.PullRequestName,
			AuthorID:          pr.AuthorID,
			Status:            PRStatus(pr.Status),
			AssignedReviewers: pr.AssignedReviewers,
		},
		ReplacedReviewer: id,
	}

	httpcommon.JSONResponse(w, http.StatusOK, resp)
}
