package application

import (
	"context"
	"fmt"

	prdomain "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	"github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"go.uber.org/zap"
)

type Service struct {
	users  domain.UserRepository
	prs    prdomain.PullRequestRepository
	logger *zap.Logger
}

func NewUserService(users domain.UserRepository, prs prdomain.PullRequestRepository, logger *zap.Logger) *Service {
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

func (s *Service) GetUserReviews(ctx context.Context, userID string) (*prdomain.UserReviews, error) {
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

	return &prdomain.UserReviews{
		UserID:       userID,
		PullRequests: prsList,
	}, nil
}

func (s *Service) DeactivateTeamUsersAndReassign(
	ctx context.Context,
	teamName string,
	userIDs []string,
) error {
	if len(userIDs) == 0 {
		return nil
	}

	members, err := s.users.ListByTeam(ctx, teamName)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to list team members for  deactivate",
				zap.String("team_name", teamName),
				zap.Error(err),
			)
		}
		return err
	}

	toDeactivateSet := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		toDeactivateSet[id] = struct{}{}
	}

	var toDeactivate []string
	var stillActive []*domain.User

	for _, u := range members {
		if !u.IsActive {
			continue
		}
		if _, ok := toDeactivateSet[u.UserID]; ok {
			toDeactivate = append(toDeactivate, u.UserID)
		} else {
			stillActive = append(stillActive, u)
		}
	}

	if len(toDeactivate) == 0 {
		return nil
	}

	shorts, err := s.prs.ListOpenByReviewers(ctx, toDeactivate)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to list open PRs for deactivated reviewers",
				zap.String("team_name", teamName),
				zap.Strings("reviewers", toDeactivate),
				zap.Error(err),
			)
		}
		return err
	}

	deactivatedSet := toDeactivateSet

	candidateIDs := make([]string, 0, len(stillActive))
	for _, u := range stillActive {
		candidateIDs = append(candidateIDs, u.UserID)
	}

	for _, sh := range shorts {
		pr, err := s.prs.GetByID(ctx, sh.PullRequestID)
		if err != nil {
			if s.logger != nil {
				s.logger.Error("failed to load full PR for reassignment",
					zap.String("pr_id", sh.PullRequestID),
					zap.Error(err),
				)
			}
			return err
		}

		if pr.Status != prdomain.PRStatusOpen {
			continue
		}

		newReviewers := make([]string, 0, len(pr.AssignedReviewers))

		assignedSet := make(map[string]struct{}, len(pr.AssignedReviewers))

		pickCandidate := func() (string, bool) {
			for _, cid := range candidateIDs {
				if _, already := assignedSet[cid]; already {
					continue
				}
				return cid, true
			}
			return "", false
		}

		for _, rID := range pr.AssignedReviewers {
			if _, isDeactivated := deactivatedSet[rID]; !isDeactivated {
				newReviewers = append(newReviewers, rID)
				assignedSet[rID] = struct{}{}
				continue
			}

			if len(candidateIDs) == 0 {
				continue
			}

			if cid, ok := pickCandidate(); ok {
				newReviewers = append(newReviewers, cid)
				assignedSet[cid] = struct{}{}
			}
		}

		if err := s.prs.SetReviewers(ctx, pr.PullRequestID, newReviewers); err != nil {
			if s.logger != nil {
				s.logger.Error("failed to update reviewers on  deactivate",
					zap.String("pr_id", pr.PullRequestID),
					zap.Strings("new_reviewers", newReviewers),
					zap.Error(err),
				)
			}
			return err
		}
	}

	for _, id := range toDeactivate {
		if err := s.users.UpdateActive(ctx, id, false); err != nil {
			if s.logger != nil {
				s.logger.Error("failed to deactivate user after reassignment",
					zap.String("user_id", id),
					zap.Error(err),
				)
			}
			return err
		}
	}

	if s.logger != nil {
		s.logger.Info(" deactivate & reassignment completed",
			zap.String("team_name", teamName),
			zap.Strings("deactivated", toDeactivate),
		)
	}

	return nil
}
