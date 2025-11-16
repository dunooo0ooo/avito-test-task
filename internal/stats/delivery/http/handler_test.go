package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	statsdomain "github.com/dunooo0ooo/avito-test-task/internal/stats/domain"
	statsmocks "github.com/dunooo0ooo/avito-test-task/internal/stats/mocks"
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

func TestStatsHandler_GetReviewerStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := statsmocks.NewMockStatsService(ctrl)
	h := NewStatsHandler(svc)
	stats := []statsdomain.ReviewerStat{
		{UserID: "u1", Count: 3},
		{UserID: "u2", Count: 5},
	}

	svc.EXPECT().
		GetReviewerStats(gomock.Any()).
		Return(stats, nil)

	req := httptest.NewRequest(http.MethodGet, "/stats/reviewers", nil).
		WithContext(context.Background())
	w := httptest.NewRecorder()

	h.GetReviewerStats(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp ReviewerStatsResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))

	require.Len(t, resp.Stats, 2)

	got := make(map[string]int64)
	for _, s := range resp.Stats {
		got[s.UserID] = s.Count
	}

	assert.Equal(t, int64(3), got["u1"])
	assert.Equal(t, int64(5), got["u2"])
}

func TestStatsHandler_GetReviewerStats_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := statsmocks.NewMockStatsService(ctrl)
	h := NewStatsHandler(svc)

	svc.EXPECT().
		GetReviewerStats(gomock.Any()).
		Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/stats/reviewers", nil).
		WithContext(context.Background())
	w := httptest.NewRecorder()

	h.GetReviewerStats(w, req)

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
