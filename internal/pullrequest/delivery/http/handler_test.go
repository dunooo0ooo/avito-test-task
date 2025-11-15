package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	prdomain "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	prmocks "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/mocks"
	userdomain "github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestPullRequestHandler_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	pr := &prdomain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add search",
		AuthorID:          "u1",
		Status:            prdomain.PRStatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
	}

	svc.EXPECT().
		CreatePullRequest(gomock.Any(), "pr-1", "Add search", "u1").
		Return(pr, nil)

	body := `{"pull_request_id":"pr-1","pull_request_name":"Add search","author_id":"u1"}`

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)

	var resp CreateResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))

	assert.Equal(t, pr.PullRequestID, resp.PullRequestDTO.PullRequestID)
	assert.Equal(t, pr.PullRequestName, resp.PullRequestDTO.PullRequestName)
	assert.Equal(t, pr.AuthorID, resp.PullRequestDTO.AuthorID)
	assert.Equal(t, string(pr.Status), string(resp.PullRequestDTO.Status))
	assert.Equal(t, pr.AssignedReviewers, resp.PullRequestDTO.AssignedReviewers)
}

func TestPullRequestHandler_Create_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()

	h.Create(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
}

func TestPullRequestHandler_Create_PRExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	svc.EXPECT().
		CreatePullRequest(gomock.Any(), "pr-1", "Add search", "u1").
		Return(nil, prdomain.ErrPullRequestAlreadyExists)

	body := `{"pull_request_id":"pr-1","pull_request_name":"Add search","author_id":"u1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusConflict, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "PR_EXISTS", errResp.Error.Code)
}

func TestPullRequestHandler_Create_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	svc.EXPECT().
		CreatePullRequest(gomock.Any(), "pr-1", "Add search", "u1").
		Return(nil, userdomain.ErrUserNotFound)

	body := `{"pull_request_id":"pr-1","pull_request_name":"Add search","author_id":"u1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusNotFound, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
}

func TestPullRequestHandler_Merge_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	pr := &prdomain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add search",
		AuthorID:          "u1",
		Status:            prdomain.PRStatusMerged,
		AssignedReviewers: []string{"u2", "u3"},
	}

	svc.EXPECT().
		MergePullRequest(gomock.Any(), "pr-1").
		Return(pr, nil)

	body := `{"pull_request_id":"pr-1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Merge(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp MergeResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))

	assert.Equal(t, pr.PullRequestID, resp.MergedPullRequestDTO.PullRequestID)
	assert.Equal(t, string(pr.Status), string(resp.MergedPullRequestDTO.Status))
	assert.Equal(t, pr.AssignedReviewers, resp.MergedPullRequestDTO.AssignedReviewers)
}

func TestPullRequestHandler_Merge_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()

	h.Merge(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
}

func TestPullRequestHandler_Merge_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	svc.EXPECT().
		MergePullRequest(gomock.Any(), "pr-1").
		Return(nil, userdomain.ErrUserNotFound)

	body := `{"pull_request_id":"pr-1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Merge(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusNotFound, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
}

func TestPullRequestHandler_Reassign_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	pr := &prdomain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add search",
		AuthorID:          "u1",
		Status:            prdomain.PRStatusOpen,
		AssignedReviewers: []string{"u3", "u4"},
	}

	svc.EXPECT().
		ReassignReviewer(gomock.Any(), "pr-1", "u2").
		Return(pr, "u4", nil)

	body := `{"pull_request_id":"pr-1","old_user_id":"u2"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Reassign(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp ReassignResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))

	assert.Equal(t, pr.PullRequestID, resp.PullRequestDTO.PullRequestID)
	assert.Equal(t, "u4", resp.ReplacedReviewer)
}

func TestPullRequestHandler_Reassign_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()

	h.Reassign(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
}

func TestPullRequestHandler_Reassign_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	svc.EXPECT().
		ReassignReviewer(gomock.Any(), "pr-1", "u2").
		Return(nil, "", prdomain.ErrPullRequestNotFound)

	body := `{"pull_request_id":"pr-1","old_user_id":"u2"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Reassign(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusNotFound, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
}

func TestPullRequestHandler_Reassign_Merged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	svc.EXPECT().
		ReassignReviewer(gomock.Any(), "pr-1", "u2").
		Return(nil, "", prdomain.ErrPullRequestMerged)

	body := `{"pull_request_id":"pr-1","old_user_id":"u2"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Reassign(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusConflict, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "PR_MERGED", errResp.Error.Code)
}

func TestPullRequestHandler_Reassign_NotAssigned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	svc.EXPECT().
		ReassignReviewer(gomock.Any(), "pr-1", "u2").
		Return(nil, "", prdomain.ErrReviewerNotAssigned)

	body := `{"pull_request_id":"pr-1","old_user_id":"u2"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Reassign(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusConflict, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "NOT_ASSIGNED", errResp.Error.Code)
}

func TestPullRequestHandler_Reassign_NoCandidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := prmocks.NewMockPullRequestService(ctrl)
	h := NewPullRequestHandler(svc)

	svc.EXPECT().
		ReassignReviewer(gomock.Any(), "pr-1", "u2").
		Return(nil, "", prdomain.ErrNoCandidate)

	body := `{"pull_request_id":"pr-1","old_user_id":"u2"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Reassign(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusConflict, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "NO_CANDIDATE", errResp.Error.Code)
}
