package application

import (
	"context"
	"fmt"

	pr "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	"github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"go.uber.org/zap"
)

type Service struct {
	users  domain.UserRepository
	prs    pr.PullRequestRepository
	logger *zap.Logger
}

func NewUserService(users domain.UserRepository, prs pr.PullRequestRepository, logger *zap.Logger) *Service {
	return &Service{
		users:  users,
		prs:    prs,
		logger: logger,
	}
}

func (s *Service) SetIsActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	if err := s.users.UpdateActive(ctx, userID, active); err != nil {
		if s.logger != nil {
			s.logger.Error("failed to update user active flag",
				zap.String("user_id", userID),
				zap.Bool("active", active),
				zap.Error(err),
			)
		}
		return nil, err
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		wrapped := fmt.Errorf("get user after update: %w", err)
		if s.logger != nil {
			s.logger.Error("failed to load user after update",
				zap.String("user_id", userID),
				zap.Error(wrapped),
			)
		}
		return nil, wrapped
	}

	if s.logger != nil {
		s.logger.Info("user active flag updated",
			zap.String("user_id", userID),
			zap.Bool("active", active),
		)
	}

	return u, nil
}

func (s *Service) GetUserReviews(ctx context.Context, userID string) (*pr.UserReviews, error) {
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if s.logger != nil {
			s.logger.Error("user not found when fetching reviews",
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
		return nil, err
	}

	prsList, err := s.prs.ListByReviewer(ctx, userID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to list pull requests for reviewer",
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("fetched user reviews",
			zap.String("user_id", userID),
			zap.Int("pull_requests_count", len(prsList)),
		)
	}

	return &pr.UserReviews{
		UserID:       userID,
		PullRequests: prsList,
	}, nil
}
