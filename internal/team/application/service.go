package application

import (
	"context"
	"github.com/dunooo0ooo/avito-test-task/internal/team/domain"
	userdomain "github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"go.uber.org/zap"
)

type TeamService struct {
	teams  domain.TeamRepository
	users  userdomain.UserRepository
	logger *zap.Logger
}

func NewTeamService(teams domain.TeamRepository, users userdomain.UserRepository, logger *zap.Logger) *TeamService {
	return &TeamService{
		teams:  teams,
		users:  users,
		logger: logger,
	}
}

func (s *TeamService) CreateTeam(
	ctx context.Context,
	teamName string,
	members []domain.TeamMember,
) (*domain.Team, error) {
	t := &domain.Team{
		TeamName: teamName,
	}

	if err := s.teams.Create(ctx, t); err != nil {
		if s.logger != nil {
			s.logger.Error("failed to create team",
				zap.String("team_name", teamName),
				zap.Error(err),
			)
		}
		return nil, err
	}

	users := make([]userdomain.User, 0, len(members))
	for _, m := range members {
		users = append(users, userdomain.User{
			UserID:   m.UserID,
			Username: m.Username,
			TeamName: teamName,
			IsActive: m.IsActive,
		})
	}

	if len(users) > 0 {
		if err := s.users.AddTeamMembers(ctx, teamName, users); err != nil {
			if s.logger != nil {
				s.logger.Error("failed to upsert team members",
					zap.String("team_name", teamName),
					zap.Int("members_count", len(users)),
					zap.Error(err),
				)
			}
			return nil, err
		}
	}

	dbUsers, err := s.users.ListByTeam(ctx, teamName)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to list team members after creation",
				zap.String("team_name", teamName),
				zap.Error(err),
			)
		}
		return nil, err
	}

	teamMembers := make([]domain.TeamMember, 0, len(dbUsers))
	for _, u := range dbUsers {
		teamMembers = append(teamMembers, domain.TeamMember{
			UserID:   u.UserID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	result := &domain.Team{
		TeamName: teamName,
		Members:  teamMembers,
	}

	if s.logger != nil {
		s.logger.Info("team created",
			zap.String("team_name", teamName),
			zap.Int("members_count", len(teamMembers)),
		)
	}

	return result, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teams.GetByName(ctx, teamName)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to get team by name",
				zap.String("team_name", teamName),
				zap.Error(err),
			)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("team loaded",
			zap.String("team_name", teamName),
			zap.Int("members_count", len(team.Members)),
		)
	}

	return team, nil
}
