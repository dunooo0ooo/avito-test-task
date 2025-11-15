package application

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	prdomain "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	userdomain "github.com/dunooo0ooo/avito-test-task/internal/user/domain"
	"go.uber.org/zap"
)

type PullRequestService struct {
	prs    prdomain.PullRequestRepository
	users  userdomain.UserRepository
	logger *zap.Logger
}

func NewPullRequestService(
	prs prdomain.PullRequestRepository,
	users userdomain.UserRepository,
	logger *zap.Logger,
) *PullRequestService {
	return &PullRequestService{
		prs:    prs,
		users:  users,
		logger: logger,
	}
}

func (s *PullRequestService) CreatePullRequest(
	ctx context.Context,
	id string,
	name string,
	authorID string,
) (*prdomain.PullRequest, error) {
	author, err := s.users.GetByID(ctx, authorID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("author not found when creating PR",
				zap.String("author_id", authorID),
				zap.Error(err),
			)
		}
		return nil, err
	}

	teamName := author.TeamName

	teamMembers, err := s.users.ListByTeam(ctx, teamName)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to list team members for PR creation",
				zap.String("team_name", teamName),
				zap.Error(err),
			)
		}
		return nil, err
	}

	var candidates []string
	for _, u := range teamMembers {
		if !u.IsActive {
			continue
		}
		if u.UserID == authorID {
			continue
		}
		candidates = append(candidates, u.UserID)
	}

	reviewers := pickRandom(candidates, 2)

	pr := &prdomain.PullRequest{
		PullRequestID:   id,
		PullRequestName: name,
		AuthorID:        authorID,
		Status:          prdomain.PRStatusOpen,
	}

	if err := s.prs.Create(ctx, pr); err != nil {
		if s.logger != nil {
			s.logger.Error("failed to create pull request",
				zap.String("pr_id", id),
				zap.Error(err),
			)
		}
		return nil, err
	}

	if err := s.prs.SetReviewers(ctx, id, reviewers); err != nil {
		if s.logger != nil {
			s.logger.Error("failed to set reviewers after PR creation",
				zap.String("pr_id", id),
				zap.Any("reviewers", reviewers),
				zap.Error(err),
			)
		}
		return nil, err
	}

	created, err := s.prs.GetByID(ctx, id)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to reload PR after creation",
				zap.String("pr_id", id),
				zap.Error(err),
			)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("pull request created",
			zap.String("pr_id", id),
			zap.String("author_id", authorID),
			zap.Strings("reviewers", reviewers),
		)
	}

	return created, nil
}

func (s *PullRequestService) MergePullRequest(ctx context.Context, id string) (*prdomain.PullRequest, error) {
	pr, err := s.prs.GetByID(ctx, id)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to get PR for merge",
				zap.String("pr_id", id),
				zap.Error(err),
			)
		}
		return nil, err
	}

	if pr.Status == prdomain.PRStatusMerged {
		if s.logger != nil {
			s.logger.Info("merge called on already merged PR",
				zap.String("pr_id", id),
			)
		}
		return nil, prdomain.ErrPullRequestMerged
	}

	now := time.Now().UTC()

	if err := s.prs.UpdateStatus(ctx, id, prdomain.PRStatusMerged, &now); err != nil {
		if s.logger != nil {
			s.logger.Error("failed to update PR status to MERGED",
				zap.String("pr_id", id),
				zap.Error(err),
			)
		}
		return nil, err
	}

	updated, err := s.prs.GetByID(ctx, id)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to reload PR after merge",
				zap.String("pr_id", id),
				zap.Error(err),
			)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("pull request merged",
			zap.String("pr_id", id),
		)
	}

	return updated, nil
}

func (s *PullRequestService) ReassignReviewer(
	ctx context.Context,
	prID string,
	oldReviewerID string,
) (*prdomain.PullRequest, string, error) {
	pr, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to get PR for reassign",
				zap.String("pr_id", prID),
				zap.Error(err),
			)
		}
		return nil, "", err
	}

	if pr.Status == prdomain.PRStatusMerged {
		if s.logger != nil {
			s.logger.Warn("attempt to reassign reviewer on merged PR",
				zap.String("pr_id", prID),
			)
		}
		return nil, "", prdomain.ErrPullRequestMerged
	}

	found := false
	for _, rID := range pr.AssignedReviewers {
		if rID == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		if s.logger != nil {
			s.logger.Warn("old reviewer is not assigned to PR",
				zap.String("pr_id", prID),
				zap.String("old_reviewer_id", oldReviewerID),
			)
		}
		return nil, "", prdomain.ErrReviewerNotAssigned
	}

	oldReviewer, err := s.users.GetByID(ctx, oldReviewerID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to load old reviewer",
				zap.String("old_reviewer_id", oldReviewerID),
				zap.Error(err),
			)
		}
		return nil, "", err
	}
	teamName := oldReviewer.TeamName

	teamMembers, err := s.users.ListByTeam(ctx, teamName)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to list team members for reassign",
				zap.String("team_name", teamName),
				zap.Error(err),
			)
		}
		return nil, "", err
	}

	assignedSet := make(map[string]struct{}, len(pr.AssignedReviewers))
	for _, rID := range pr.AssignedReviewers {
		assignedSet[rID] = struct{}{}
	}

	var candidates []string
	for _, u := range teamMembers {
		if !u.IsActive {
			continue
		}
		if u.UserID == oldReviewerID {
			continue
		}
		if _, already := assignedSet[u.UserID]; already {
			continue
		}
		candidates = append(candidates, u.UserID)
	}

	if len(candidates) == 0 {
		if s.logger != nil {
			s.logger.Warn("no candidate for reviewer reassign",
				zap.String("pr_id", prID),
				zap.String("old_reviewer_id", oldReviewerID),
			)
		}
		return nil, "", prdomain.ErrNoCandidate
	}

	newReviewerID := pickRandom(candidates, 1)[0]

	newReviewers := make([]string, len(pr.AssignedReviewers))
	for i, rID := range pr.AssignedReviewers {
		if rID == oldReviewerID {
			newReviewers[i] = newReviewerID
		} else {
			newReviewers[i] = rID
		}
	}

	if err := s.prs.SetReviewers(ctx, prID, newReviewers); err != nil {
		if s.logger != nil {
			s.logger.Error("failed to update reviewers on reassign",
				zap.String("pr_id", prID),
				zap.Strings("new_reviewers", newReviewers),
				zap.Error(err),
			)
		}
		return nil, "", err
	}

	updated, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to reload PR after reassign",
				zap.String("pr_id", prID),
				zap.Error(err),
			)
		}
		return nil, "", err
	}

	if s.logger != nil {
		s.logger.Info("reviewer reassigned",
			zap.String("pr_id", prID),
			zap.String("old_reviewer_id", oldReviewerID),
			zap.String("new_reviewer_id", newReviewerID),
		)
	}

	return updated, newReviewerID, nil
}

func pickRandom(src []string, n int) []string {
	if n <= 0 || len(src) == 0 {
		return nil
	}
	if len(src) <= n {
		out := make([]string, len(src))
		copy(out, src)
		return out
	}

	used := make(map[int]struct{})
	res := make([]string, 0, n)

	for len(res) < n {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(src))))
		i := int(idx.Int64())
		if _, ok := used[i]; ok {
			continue
		}
		used[i] = struct{}{}
		res = append(res, src[i])
	}

	return res
}
