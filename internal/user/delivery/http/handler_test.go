package http

import (
	"encoding/json"
	prdomain "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	userdomain "github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"github.com/dunooo0ooo/avito-test-task/internal/user/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestUserHandler_SetIsActive_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	expected := &userdomain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	svc.EXPECT().
		SetIsActive(gomock.Any(), "u1", true).
		Return(expected, nil)

	body := `{"user_id":"u1","is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.SetIsActive(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp SetIsActiveResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))

	assert.Equal(t, expected.UserID, resp.User.UserID)
	assert.Equal(t, expected.Username, resp.User.Username)
	assert.Equal(t, expected.TeamName, resp.User.TeamName)
	assert.Equal(t, expected.IsActive, resp.User.IsActive)
}

func TestUserHandler_SetIsActive_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader("{invalid"))
	w := httptest.NewRecorder()

	h.SetIsActive(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
}

func TestUserHandler_SetIsActive_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	svc.EXPECT().
		SetIsActive(gomock.Any(), "u1", true).
		Return(nil, userdomain.ErrUserNotFound)

	body := `{"user_id":"u1","is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.SetIsActive(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusNotFound, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
}

func TestUserHandler_GetUserReviews_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	expected := &prdomain.UserReviews{
		UserID: "u2",
		PullRequests: []prdomain.PullRequestShort{
			{
				PullRequestID:   "pr-1",
				PullRequestName: "Add search",
				AuthorID:        "u1",
				Status:          "OPEN",
			},
		},
	}

	svc.EXPECT().
		GetUserReviews(gomock.Any(), "u2").
		Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	w := httptest.NewRecorder()

	h.GetUserReviews(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp GetReviewsResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))

	assert.Equal(t, expected.UserID, resp.UserID)
	assert.Len(t, resp.PullRequests, len(expected.PullRequests))
}

func TestUserHandler_GetUserReviews_MissingUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
	w := httptest.NewRecorder()

	h.GetUserReviews(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
}
