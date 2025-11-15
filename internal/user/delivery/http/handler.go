package http

import (
	"context"
	"encoding/json"
	"errors"
	pr "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	"github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"github.com/dunooo0ooo/avito-test-task/pkg/httpcommon"
	"net/http"
)

type UserService interface {
	SetIsActive(ctx context.Context, userID string, active bool) (*domain.User, error)

	GetUserReviews(ctx context.Context, userID string) (*pr.UserReviews, error)
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users/setIsActive", h.SetIsActive)
	mux.HandleFunc("GET /users/getReview", h.GetUserReviews)
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			httpcommon.JSONError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		case errors.Is(err, domain.ErrInternalDatabase):
			httpcommon.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		default:
			httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return
	}

	resp := SetIsActiveResponse{
		User: UserDTO{
			UserID:   user.UserID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}

	httpcommon.JSONResponse(w, http.StatusOK, resp)
}

func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
		return
	}

	reviews, err := h.userService.GetUserReviews(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			httpcommon.JSONError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		case errors.Is(err, domain.ErrInternalDatabase):
			httpcommon.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		default:
			httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return
	}

	resp := GetReviewsResponse{
		UserID:       reviews.UserID,
		PullRequests: reviews.PullRequests,
	}

	httpcommon.JSONResponse(w, http.StatusOK, resp)
}
