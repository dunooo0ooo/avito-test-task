package application

import (
	"context"
	"errors"
	"testing"

	teamdomain "github.com/dunooo0ooo/avito-test-task/internal/team/domain"
	teammocks "github.com/dunooo0ooo/avito-test-task/internal/team/mocks"
	userdomain "github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	usermocks "github.com/dunooo0ooo/avito-test-task/internal/user/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTeamService_CreateTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := teammocks.NewMockTeamRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewTeamService(teamRepo, userRepo, logger)
	ctx := context.Background()

	teamName := "backend"
	inputMembers := []teamdomain.TeamMember{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: false},
	}

	dbUsers := []*userdomain.User{
		{UserID: "u1", Username: "Alice", TeamName: teamName, IsActive: true},
		{UserID: "u2", Username: "Bob", TeamName: teamName, IsActive: false},
	}

	gomock.InOrder(
		teamRepo.EXPECT().
			Create(gomock.Any(), &teamdomain.Team{TeamName: teamName}).
			Return(nil),

		userRepo.EXPECT().
			AddTeamMembers(gomock.Any(), teamName, gomock.Any()).
			Return(nil),

		userRepo.EXPECT().
			ListByTeam(gomock.Any(), teamName).
			Return(dbUsers, nil),
	)

	result, err := svc.CreateTeam(ctx, teamName, inputMembers)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, teamName, result.TeamName)
	require.Len(t, result.Members, 2)

	assert.Equal(t, "u1", result.Members[0].UserID)
	assert.Equal(t, "Alice", result.Members[0].Username)
	assert.True(t, result.Members[0].IsActive)

	assert.Equal(t, "u2", result.Members[1].UserID)
	assert.Equal(t, "Bob", result.Members[1].Username)
	assert.False(t, result.Members[1].IsActive)
}

func TestTeamService_CreateTeam_CreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := teammocks.NewMockTeamRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewTeamService(teamRepo, userRepo, logger)
	ctx := context.Background()

	teamName := "backend"
	inputMembers := []teamdomain.TeamMember{
		{UserID: "u1", Username: "Alice", IsActive: true},
	}

	expectedErr := errors.New("db create error")

	teamRepo.EXPECT().
		Create(gomock.Any(), &teamdomain.Team{TeamName: teamName}).
		Return(expectedErr)

	// userRepo не должен вызываться
	result, err := svc.CreateTeam(ctx, teamName, inputMembers)
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Nil(t, result)
}

func TestTeamService_CreateTeam_AddTeamMembersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := teammocks.NewMockTeamRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewTeamService(teamRepo, userRepo, logger)
	ctx := context.Background()

	teamName := "backend"
	inputMembers := []teamdomain.TeamMember{
		{UserID: "u1", Username: "Alice", IsActive: true},
	}

	expectedErr := errors.New("add team members error")

	gomock.InOrder(
		teamRepo.EXPECT().
			Create(gomock.Any(), &teamdomain.Team{TeamName: teamName}).
			Return(nil),

		userRepo.EXPECT().
			AddTeamMembers(gomock.Any(), teamName, gomock.Any()).
			Return(expectedErr),
	)

	result, err := svc.CreateTeam(ctx, teamName, inputMembers)
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Nil(t, result)
}

func TestTeamService_CreateTeam_ListByTeamError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := teammocks.NewMockTeamRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewTeamService(teamRepo, userRepo, logger)
	ctx := context.Background()

	teamName := "backend"
	inputMembers := []teamdomain.TeamMember{
		{UserID: "u1", Username: "Alice", IsActive: true},
	}

	expectedErr := errors.New("list by team error")

	gomock.InOrder(
		teamRepo.EXPECT().
			Create(gomock.Any(), &teamdomain.Team{TeamName: teamName}).
			Return(nil),

		userRepo.EXPECT().
			AddTeamMembers(gomock.Any(), teamName, gomock.Any()).
			Return(nil),

		userRepo.EXPECT().
			ListByTeam(gomock.Any(), teamName).
			Return(nil, expectedErr),
	)

	result, err := svc.CreateTeam(ctx, teamName, inputMembers)
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Nil(t, result)
}

func TestTeamService_GetTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := teammocks.NewMockTeamRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewTeamService(teamRepo, userRepo, logger)
	ctx := context.Background()

	teamName := "backend"
	team := &teamdomain.Team{
		TeamName: teamName,
		Members: []teamdomain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	teamRepo.EXPECT().
		GetByName(gomock.Any(), teamName).
		Return(team, nil)

	result, err := svc.GetTeam(ctx, teamName)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, teamName, result.TeamName)
	require.Len(t, result.Members, 1)
	assert.Equal(t, "u1", result.Members[0].UserID)
}

func TestTeamService_GetTeam_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := teammocks.NewMockTeamRepository(ctrl)
	userRepo := usermocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()

	svc := NewTeamService(teamRepo, userRepo, logger)
	ctx := context.Background()

	teamName := "backend"
	expectedErr := errors.New("team not found")

	teamRepo.EXPECT().
		GetByName(gomock.Any(), teamName).
		Return(nil, expectedErr)

	result, err := svc.GetTeam(ctx, teamName)
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Nil(t, result)
}
