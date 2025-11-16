package application

import (
	"context"
	"errors"
	"testing"

	prmocks "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStatsService_GetReviewerStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()
	svc := NewStatsService(prRepo, logger)

	ctx := context.Background()

	raw := map[string]string{
		"u1": "3",
		"u2": "10",
	}

	prRepo.EXPECT().
		CountByReviewer(gomock.Any()).
		Return(raw, nil)

	res, err := svc.GetReviewerStats(ctx)
	require.NoError(t, err)
	require.Len(t, res, 2)

	got := make(map[string]int64)
	for _, s := range res {
		got[s.UserID] = s.Count
	}

	assert.Equal(t, int64(3), got["u1"])
	assert.Equal(t, int64(10), got["u2"])
}

func TestStatsService_GetReviewerStats_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()
	svc := NewStatsService(prRepo, logger)

	ctx := context.Background()

	expectedErr := errors.New("db error")

	prRepo.EXPECT().
		CountByReviewer(gomock.Any()).
		Return(nil, expectedErr)

	res, err := svc.GetReviewerStats(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Nil(t, res)
}

func TestStatsService_GetReviewerStats_ParseError_SkipsBadEntry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()
	svc := NewStatsService(prRepo, logger)

	ctx := context.Background()

	raw := map[string]string{
		"u1": "not-a-number",
		"u2": "5",
	}

	prRepo.EXPECT().
		CountByReviewer(gomock.Any()).
		Return(raw, nil)

	res, err := svc.GetReviewerStats(ctx)
	require.NoError(t, err)

	require.Len(t, res, 1)
	assert.Equal(t, "u2", res[0].UserID)
	assert.Equal(t, int64(5), res[0].Count)
}
