package application

import (
	"context"
	prdomain "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/domain"
	"github.com/dunooo0ooo/avito-test-task/internal/stats/domain"
	"go.uber.org/zap"
	"strconv"
)

type StatsService struct {
	prs    prdomain.PullRequestRepository
	logger *zap.Logger
}

func NewStatsService(prs prdomain.PullRequestRepository, logger *zap.Logger) *StatsService {
	return &StatsService{prs: prs, logger: logger}

}

func (s *StatsService) GetReviewerStats(ctx context.Context) ([]domain.ReviewerStat, error) {
	raw, err := s.prs.CountByReviewer(ctx)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to get reviewer stats",
				zap.Error(err),
			)
		}
		return nil, err
	}

	stats := make([]domain.ReviewerStat, 0, len(raw))
	for userID, cntStr := range raw {
		n, parseErr := strconv.ParseInt(cntStr, 10, 64)
		if parseErr != nil {
			if s.logger != nil {
				s.logger.Warn("failed to parse count from repo",
					zap.String("user_id", userID),
					zap.String("raw_count", cntStr),
					zap.Error(parseErr),
				)
			}
			continue
		}

		stats = append(stats, domain.ReviewerStat{
			UserID: userID,
			Count:  n,
		})
	}

	return stats, nil
}
