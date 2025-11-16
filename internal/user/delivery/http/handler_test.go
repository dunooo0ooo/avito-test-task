package http

import (
	"encoding/json"
	"errors"
	prdomain "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	teamdomain "github.com/dunooo0ooo/avito-test-task/internal/team/domain"
	userdomain "github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"github.com/dunooo0ooo/avito-test-task/internal/user/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusNotFound, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
}

func TestUserHandler_SetIsActive_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	svc.EXPECT().
		SetIsActive(gomock.Any(), "u1", true).
		Return(nil, userdomain.ErrInternalDatabase)

	body := `{"user_id":"u1","is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.SetIsActive(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "INTERNAL_ERROR", errResp.Error.Code)
}

func TestUserHandler_SetIsActive_UnexpectedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	svc.EXPECT().
		SetIsActive(gomock.Any(), "u1", true).
		Return(nil, errors.New("some error"))

	body := `{"user_id":"u1","is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.SetIsActive(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
	assert.Equal(t, "some error", errResp.Error.Message)
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
}

func TestUserHandler_GetUserReviews_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	svc.EXPECT().
		GetUserReviews(gomock.Any(), "u2").
		Return(nil, userdomain.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	w := httptest.NewRecorder()

	h.GetUserReviews(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusNotFound, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
}

func TestUserHandler_GetUserReviews_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	svc.EXPECT().
		GetUserReviews(gomock.Any(), "u2").
		Return(nil, userdomain.ErrInternalDatabase)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	w := httptest.NewRecorder()

	h.GetUserReviews(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "INTERNAL_ERROR", errResp.Error.Code)
}

func TestUserHandler_GetUserReviews_UnexpectedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	svc.EXPECT().
		GetUserReviews(gomock.Any(), "u2").
		Return(nil, errors.New("boom"))

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	w := httptest.NewRecorder()

	h.GetUserReviews(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
	assert.Equal(t, "boom", errResp.Error.Message)
}

func TestUserHandler_BulkDeactivate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	body := `{
		"team_name": "backend",
		"user_ids": ["u2", "u3"]
	}`

	svc.EXPECT().
		DeactivateTeamUsersAndReassign(gomock.Any(), "backend", []string{"u2", "u3"}).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/team/deactivateMembers", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.BulkDeactivate(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp DeactivateResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))

	assert.Equal(t, "backend", resp.TeamName)
	assert.ElementsMatch(t, []string{"u2", "u3"}, resp.Deactivated)
}

func TestUserHandler_BulkDeactivate_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/team/deactivateMembers", strings.NewReader("{invalid"))
	w := httptest.NewRecorder()

	h.BulkDeactivate(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
	assert.Equal(t, "invalid request body", errResp.Error.Message)
}

func TestUserHandler_BulkDeactivate_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	body := `{
		"team_name": "",
		"user_ids": ["u2"]
	}`

	req := httptest.NewRequest(http.MethodPost, "/team/deactivateMembers", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.BulkDeactivate(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
	assert.Equal(t, "team_name and user_ids are required", errResp.Error.Message)

	body2 := `{
		"team_name": "backend",
		"user_ids": []
	}`

	req2 := httptest.NewRequest(http.MethodPost, "/team/deactivateMembers", strings.NewReader(body2))
	w2 := httptest.NewRecorder()

	h.BulkDeactivate(w2, req2)

	res2 := w2.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res2.Body)

	require.Equal(t, http.StatusBadRequest, res2.StatusCode)

	var errResp2 errorResponse
	require.NoError(t, json.NewDecoder(res2.Body).Decode(&errResp2))

	assert.Equal(t, "BAD_REQUEST", errResp2.Error.Code)
	assert.Equal(t, "team_name and user_ids are required", errResp2.Error.Message)
}

func TestUserHandler_BulkDeactivate_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockUserService(ctrl)
	h := NewUserHandler(svc)

	body := `{
		"team_name": "unknown",
		"user_ids": ["u2"]
	}`

	svc.EXPECT().
		DeactivateTeamUsersAndReassign(gomock.Any(), "unknown", []string{"u2"}).
		Return(teamdomain.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodPost, "/team/deactivateMembers", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.BulkDeactivate(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusNotFound, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
	assert.Equal(t, "team not found", errResp.Error.Message)
}
