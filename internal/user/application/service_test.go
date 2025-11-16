package application

import (
	"context"
	"errors"
	"testing"

	prdomain "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	prmocks "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/mocks"
	userdomain "github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	usermocks "github.com/dunooo0ooo/avito-test-task/internal/user/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_SetIsActive_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)

	ctx := context.Background()

	userRepo.EXPECT().
		UpdateActive(gomock.Any(), "u1", true).
		Return(nil)

	expectedUser := &userdomain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	userRepo.EXPECT().
		GetByID(gomock.Any(), "u1").
		Return(expectedUser, nil)

	u, err := svc.SetIsActive(ctx, "u1", true)
	require.NoError(t, err)
	require.NotNil(t, u)

	assert.Equal(t, expectedUser.UserID, u.UserID)
	assert.Equal(t, expectedUser.Username, u.Username)
	assert.Equal(t, expectedUser.TeamName, u.TeamName)
	assert.Equal(t, expectedUser.IsActive, u.IsActive)
}

func TestService_SetIsActive_UpdateFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)

	ctx := context.Background()

	userRepo.EXPECT().
		UpdateActive(gomock.Any(), "u1", true).
		Return(userdomain.ErrUserNotFound)

	u, err := svc.SetIsActive(ctx, "u1", true)
	require.Error(t, err)
	assert.True(t, errors.Is(err, userdomain.ErrUserNotFound))
	assert.Nil(t, u)
}

func TestService_SetIsActive_GetAfterUpdateFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)

	ctx := context.Background()

	userRepo.EXPECT().
		UpdateActive(gomock.Any(), "u1", true).
		Return(nil)

	userRepo.EXPECT().
		GetByID(gomock.Any(), "u1").
		Return(nil, userdomain.ErrUserNotFound)

	u, err := svc.SetIsActive(ctx, "u1", true)
	require.Error(t, err)
	assert.True(t, errors.Is(err, userdomain.ErrUserNotFound))
	assert.Nil(t, u)
}

func TestService_GetUserReviews_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)

	ctx := context.Background()

	userRepo.EXPECT().
		GetByID(gomock.Any(), "u2").
		Return(&userdomain.User{UserID: "u2"}, nil)

	expectedPRs := []prdomain.PullRequestShort{
		{
			PullRequestID:   "pr-1",
			PullRequestName: "Add search",
			AuthorID:        "u1",
			Status:          "OPEN",
		},
	}

	prRepo.EXPECT().
		ListByReviewer(gomock.Any(), "u2").
		Return(expectedPRs, nil)

	res, err := svc.GetUserReviews(ctx, "u2")
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "u2", res.UserID)
	assert.Len(t, res.PullRequests, len(expectedPRs))
	assert.Equal(t, expectedPRs[0].PullRequestID, res.PullRequests[0].PullRequestID)
}

func TestService_GetUserReviews_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)

	ctx := context.Background()

	userRepo.EXPECT().
		GetByID(gomock.Any(), "u2").
		Return(nil, userdomain.ErrUserNotFound)

	res, err := svc.GetUserReviews(ctx, "u2")
	require.Error(t, err)
	assert.True(t, errors.Is(err, userdomain.ErrUserNotFound))
	assert.Nil(t, res)
}

func TestService_GetUserReviews_ListByReviewerFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)

	ctx := context.Background()

	userRepo.EXPECT().
		GetByID(gomock.Any(), "u2").
		Return(&userdomain.User{UserID: "u2"}, nil)

	expectedErr := errors.New("db failure")

	prRepo.EXPECT().
		ListByReviewer(gomock.Any(), "u2").
		Return(nil, expectedErr)

	res, err := svc.GetUserReviews(ctx, "u2")
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Nil(t, res)
}

func TestService_DeactivateTeamUsersAndReassign_EmptyUserIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)
	ctx := context.Background()

	err := svc.DeactivateTeamUsersAndReassign(ctx, "backend", []string{})
	require.NoError(t, err)
}

func TestService_DeactivateTeamUsersAndReassign_ListByTeamError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)
	ctx := context.Background()

	expectedErr := errors.New("db error")

	userRepo.EXPECT().
		ListByTeam(gomock.Any(), "backend").
		Return(nil, expectedErr)

	err := svc.DeactivateTeamUsersAndReassign(ctx, "backend", []string{"u2"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
}

func TestService_DeactivateTeamUsersAndReassign_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)
	ctx := context.Background()

	members := []*userdomain.User{
		{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true},
		{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true},
		{UserID: "u3", Username: "Carol", TeamName: "backend", IsActive: true},
	}

	shorts := []prdomain.PullRequestShort{
		{
			PullRequestID:   "pr-1",
			PullRequestName: "Add search",
			AuthorID:        "u0",
			Status:          prdomain.PRStatusOpen,
		},
	}

	fullPR := &prdomain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add search",
		AuthorID:          "u0",
		Status:            prdomain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
	}

	gomock.InOrder(
		userRepo.EXPECT().
			ListByTeam(gomock.Any(), "backend").
			Return(members, nil),

		prRepo.EXPECT().
			ListOpenByReviewers(gomock.Any(), []string{"u2"}).
			Return(shorts, nil),

		prRepo.EXPECT().
			GetByID(gomock.Any(), "pr-1").
			Return(fullPR, nil),

		prRepo.EXPECT().
			SetReviewers(gomock.Any(), "pr-1", gomock.Any()).
			Return(nil),

		userRepo.EXPECT().
			UpdateActive(gomock.Any(), "u2", false).
			Return(nil),
	)

	err := svc.DeactivateTeamUsersAndReassign(ctx, "backend", []string{"u2"})
	require.NoError(t, err)
}

func TestService_DeactivateTeamUsersAndReassign_NoMatchingActiveUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)
	ctx := context.Background()

	members := []*userdomain.User{
		{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true},
		{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: false},
	}

	userRepo.EXPECT().
		ListByTeam(gomock.Any(), "backend").
		Return(members, nil)

	err := svc.DeactivateTeamUsersAndReassign(ctx, "backend", []string{"u3"})
	require.NoError(t, err)
}

func TestService_DeactivateTeamUsersAndReassign_UpdateActiveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := usermocks.NewMockUserRepository(ctrl)
	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	logger := zap.NewNop()

	svc := NewUserService(userRepo, prRepo, logger)
	ctx := context.Background()

	members := []*userdomain.User{
		{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true},
		{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true},
		{UserID: "u3", Username: "Carol", TeamName: "backend", IsActive: true},
	}

	shorts := []prdomain.PullRequestShort{
		{
			PullRequestID:   "pr-1",
			PullRequestName: "Add search",
			AuthorID:        "u0",
			Status:          prdomain.PRStatusOpen,
		},
	}

	fullPR := &prdomain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add search",
		AuthorID:          "u0",
		Status:            prdomain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
	}

	updateErr := errors.New("update active failed")

	gomock.InOrder(
		userRepo.EXPECT().
			ListByTeam(gomock.Any(), "backend").
			Return(members, nil),

		prRepo.EXPECT().
			ListOpenByReviewers(gomock.Any(), []string{"u2"}).
			Return(shorts, nil),

		prRepo.EXPECT().
			GetByID(gomock.Any(), "pr-1").
			Return(fullPR, nil),

		prRepo.EXPECT().
			SetReviewers(gomock.Any(), "pr-1", gomock.Any()).
			Return(nil),

		userRepo.EXPECT().
			UpdateActive(gomock.Any(), "u2", false).
			Return(updateErr),
	)

	err := svc.DeactivateTeamUsersAndReassign(ctx, "backend", []string{"u2"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, updateErr))
}
