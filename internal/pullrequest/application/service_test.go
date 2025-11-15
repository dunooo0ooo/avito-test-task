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

func TestCreatePullRequest_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	author := &userdomain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	teamMembers := []*userdomain.User{
		author,
		{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true},
		{UserID: "u3", Username: "Carol", TeamName: "backend", IsActive: true},
	}

	userRepo.EXPECT().
		GetByID(gomock.Any(), "u1").
		Return(author, nil)

	userRepo.EXPECT().
		ListByTeam(gomock.Any(), "backend").
		Return(teamMembers, nil)

	prRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil)

	prRepo.EXPECT().
		SetReviewers(gomock.Any(), "pr-1", gomock.Any()).
		Return(nil)

	expectedPR := &prdomain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add search",
		AuthorID:        "u1",
		Status:          prdomain.PRStatusOpen,
		AssignedReviewers: []string{
			"u2", "u3",
		},
	}

	prRepo.EXPECT().
		GetByID(gomock.Any(), "pr-1").
		Return(expectedPR, nil)

	pr, err := svc.CreatePullRequest(ctx, "pr-1", "Add search", "u1")
	require.NoError(t, err)
	require.NotNil(t, pr)

	assert.Equal(t, expectedPR.PullRequestID, pr.PullRequestID)
	assert.Equal(t, expectedPR.PullRequestName, pr.PullRequestName)
	assert.Equal(t, expectedPR.AuthorID, pr.AuthorID)
	assert.Equal(t, expectedPR.Status, pr.Status)
}

func TestCreatePullRequest_AuthorNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	userRepo.EXPECT().
		GetByID(gomock.Any(), "u1").
		Return(nil, userdomain.ErrUserNotFound)

	pr, err := svc.CreatePullRequest(ctx, "pr-1", "Add search", "u1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, userdomain.ErrUserNotFound))
	assert.Nil(t, pr)
}

func TestCreatePullRequest_ListTeamError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	author := &userdomain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	userRepo.EXPECT().
		GetByID(gomock.Any(), "u1").
		Return(author, nil)

	expectedErr := errors.New("db error")
	userRepo.EXPECT().
		ListByTeam(gomock.Any(), "backend").
		Return(nil, expectedErr)

	pr, err := svc.CreatePullRequest(ctx, "pr-1", "Add search", "u1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Nil(t, pr)
}

func TestMergePullRequest_AlreadyMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	existing := &prdomain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add search",
		AuthorID:        "u1",
		Status:          prdomain.PRStatusMerged,
	}

	prRepo.EXPECT().
		GetByID(gomock.Any(), "pr-1").
		Return(existing, nil)

	pr, err := svc.MergePullRequest(ctx, "pr-1")
	require.NoError(t, err)
	require.NotNil(t, pr)

	assert.Equal(t, prdomain.PRStatusMerged, pr.Status)
}

func TestMergePullRequest_FromOpenToMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	openPR := &prdomain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add search",
		AuthorID:        "u1",
		Status:          prdomain.PRStatusOpen,
	}

	mergedPR := &prdomain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add search",
		AuthorID:        "u1",
		Status:          prdomain.PRStatusMerged,
	}

	gomock.InOrder(
		prRepo.EXPECT().
			GetByID(gomock.Any(), "pr-1").
			Return(openPR, nil),
		prRepo.EXPECT().
			UpdateStatus(gomock.Any(), "pr-1", prdomain.PRStatusMerged, gomock.Any()).
			Return(nil),
		prRepo.EXPECT().
			GetByID(gomock.Any(), "pr-1").
			Return(mergedPR, nil),
	)

	pr, err := svc.MergePullRequest(ctx, "pr-1")
	require.NoError(t, err)
	require.NotNil(t, pr)

	assert.Equal(t, prdomain.PRStatusMerged, pr.Status)
}

func TestMergePullRequest_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	expectedErr := errors.New("db error")

	prRepo.EXPECT().
		GetByID(gomock.Any(), "pr-1").
		Return(nil, expectedErr)

	pr, err := svc.MergePullRequest(ctx, "pr-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Nil(t, pr)
}

func TestReassignReviewer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	pr := &prdomain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add search",
		AuthorID:          "u1",
		Status:            prdomain.PRStatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
	}

	userOld := &userdomain.User{
		UserID:   "u2",
		Username: "Bob",
		TeamName: "backend",
		IsActive: true,
	}

	teamMembers := []*userdomain.User{
		userOld,
		{UserID: "u4", Username: "Dora", TeamName: "backend", IsActive: true},
	}

	updatedPR := &prdomain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add search",
		AuthorID:          "u1",
		Status:            prdomain.PRStatusOpen,
		AssignedReviewers: []string{"u4", "u3"},
	}

	gomock.InOrder(
		prRepo.EXPECT().
			GetByID(gomock.Any(), "pr-1").
			Return(pr, nil),
		userRepo.EXPECT().
			GetByID(gomock.Any(), "u2").
			Return(userOld, nil),
		userRepo.EXPECT().
			ListByTeam(gomock.Any(), "backend").
			Return(teamMembers, nil),
		prRepo.EXPECT().
			SetReviewers(gomock.Any(), "pr-1", gomock.Any()).
			Return(nil),
		prRepo.EXPECT().
			GetByID(gomock.Any(), "pr-1").
			Return(updatedPR, nil),
	)

	resPR, newReviewer, err := svc.ReassignReviewer(ctx, "pr-1", "u2")
	require.NoError(t, err)
	require.NotNil(t, resPR)

	assert.Equal(t, updatedPR.PullRequestID, resPR.PullRequestID)
	assert.Equal(t, updatedPR.Status, resPR.Status)
	assert.Contains(t, []string{"u4"}, newReviewer)
}

func TestReassignReviewer_OnMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	pr := &prdomain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add search",
		AuthorID:        "u1",
		Status:          prdomain.PRStatusMerged,
	}

	prRepo.EXPECT().
		GetByID(gomock.Any(), "pr-1").
		Return(pr, nil)

	resPR, newReviewer, err := svc.ReassignReviewer(ctx, "pr-1", "u2")
	require.Error(t, err)
	assert.True(t, errors.Is(err, prdomain.ErrPullRequestMerged))
	assert.Nil(t, resPR)
	assert.Equal(t, "", newReviewer)
}

func TestReassignReviewer_OldNotAssigned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	pr := &prdomain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add search",
		AuthorID:          "u1",
		Status:            prdomain.PRStatusOpen,
		AssignedReviewers: []string{"u3", "u4"},
	}

	prRepo.EXPECT().
		GetByID(gomock.Any(), "pr-1").
		Return(pr, nil)

	resPR, newReviewer, err := svc.ReassignReviewer(ctx, "pr-1", "u2")
	require.Error(t, err)
	assert.True(t, errors.Is(err, prdomain.ErrReviewerNotAssigned))
	assert.Nil(t, resPR)
	assert.Equal(t, "", newReviewer)
}

func TestReassignReviewer_NoCandidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := prmocks.NewMockPullRequestRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewPullRequestService(prRepo, userRepo, logger)
	ctx := context.Background()

	pr := &prdomain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add search",
		AuthorID:          "u1",
		Status:            prdomain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
	}

	oldRev := &userdomain.User{
		UserID:   "u2",
		Username: "Bob",
		TeamName: "backend",
		IsActive: true,
	}

	teamMembers := []*userdomain.User{
		oldRev,
	}

	gomock.InOrder(
		prRepo.EXPECT().
			GetByID(gomock.Any(), "pr-1").
			Return(pr, nil),
		userRepo.EXPECT().
			GetByID(gomock.Any(), "u2").
			Return(oldRev, nil),
		userRepo.EXPECT().
			ListByTeam(gomock.Any(), "backend").
			Return(teamMembers, nil),
	)

	resPR, newReviewer, err := svc.ReassignReviewer(ctx, "pr-1", "u2")
	require.Error(t, err)
	assert.True(t, errors.Is(err, prdomain.ErrNoCandidate))
	assert.Nil(t, resPR)
	assert.Equal(t, "", newReviewer)
}
