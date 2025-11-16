package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	teamdomain "github.com/dunooo0ooo/avito-test-task/internal/team/domain"
	teammocks "github.com/dunooo0ooo/avito-test-task/internal/team/mocks"
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

func TestTeamHandler_AddTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := teammocks.NewMockTeamService(ctrl)
	h := NewTeamHandler(svc)

	reqBody := AddTeamRequest{
		Name: "backend",
		Members: []TeamMemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	team := &teamdomain.Team{
		TeamName: "backend",
		Members: []teamdomain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		},
	}

	svc.EXPECT().
		CreateTeam(gomock.Any(), "backend", []teamdomain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		}).
		Return(team, nil)

	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddTeam(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusCreated, res.StatusCode)

	var resp AddTeamResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "backend", resp.Team.TeamName)
	require.Len(t, resp.Team.Members, 2)
	assert.Equal(t, "u1", resp.Team.Members[0].UserID)
	assert.Equal(t, "Alice", resp.Team.Members[0].Username)
	assert.True(t, resp.Team.Members[0].IsActive)
}

func TestTeamHandler_AddTeam_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := teammocks.NewMockTeamService(ctrl)
	h := NewTeamHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader([]byte(`{invalid`)))
	w := httptest.NewRecorder()

	h.AddTeam(w, req)

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

func TestTeamHandler_AddTeam_TeamExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := teammocks.NewMockTeamService(ctrl)
	h := NewTeamHandler(svc)

	reqBody := AddTeamRequest{
		Name: "backend",
		Members: []TeamMemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	svc.EXPECT().
		CreateTeam(gomock.Any(), "backend", []teamdomain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		}).
		Return(nil, teamdomain.ErrTeamAlreadyExists)

	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddTeam(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "TEAM_EXISTS", errResp.Error.Code)
	assert.Equal(t, "team_name already exists", errResp.Error.Message)
}

func TestTeamHandler_AddTeam_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := teammocks.NewMockTeamService(ctrl)
	h := NewTeamHandler(svc)

	reqBody := AddTeamRequest{
		Name: "backend",
		Members: []TeamMemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	svc.EXPECT().
		CreateTeam(gomock.Any(), "backend", []teamdomain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		}).
		Return(nil, teamdomain.ErrInternalDatabase)

	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddTeam(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "INTERNAL_ERROR", errResp.Error.Code)
	assert.Equal(t, "internal server error", errResp.Error.Message)
}

func TestTeamHandler_GetTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := teammocks.NewMockTeamService(ctrl)
	h := NewTeamHandler(svc)

	team := &teamdomain.Team{
		TeamName: "backend",
		Members: []teamdomain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	svc.EXPECT().
		GetTeam(gomock.Any(), "backend").
		Return(team, nil)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	w := httptest.NewRecorder()

	h.GetTeam(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp TeamDTO
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))

	assert.Equal(t, "backend", resp.TeamName)
	require.Len(t, resp.Members, 1)
	assert.Equal(t, "u1", resp.Members[0].UserID)
	assert.Equal(t, "Alice", resp.Members[0].Username)
	assert.True(t, resp.Members[0].IsActive)
}

func TestTeamHandler_GetTeam_MissingTeamName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := teammocks.NewMockTeamService(ctrl)
	h := NewTeamHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
	w := httptest.NewRecorder()

	h.GetTeam(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
	assert.Equal(t, "team_name is required", errResp.Error.Message)
}

func TestTeamHandler_GetTeam_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := teammocks.NewMockTeamService(ctrl)
	h := NewTeamHandler(svc)

	svc.EXPECT().
		GetTeam(gomock.Any(), "backend").
		Return(nil, teamdomain.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	w := httptest.NewRecorder()

	h.GetTeam(w, req)

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

func TestTeamHandler_GetTeam_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := teammocks.NewMockTeamService(ctrl)
	h := NewTeamHandler(svc)

	svc.EXPECT().
		GetTeam(gomock.Any(), "backend").
		Return(nil, teamdomain.ErrInternalDatabase)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	w := httptest.NewRecorder()

	h.GetTeam(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errResp))

	assert.Equal(t, "INTERNAL_ERROR", errResp.Error.Code)
	assert.Equal(t, "internal server error", errResp.Error.Message)
}
