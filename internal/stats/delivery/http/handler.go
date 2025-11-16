package http

import (
	"context"
	"github.com/dunooo0ooo/avito-test-task/internal/stats/domain"
	"net/http"

	"github.com/dunooo0ooo/avito-test-task/pkg/httpcommon"
)

type StatsService interface {
	GetReviewerStats(ctx context.Context) ([]domain.ReviewerStat, error)
}

type StatsHandler struct {
	svc StatsService
}

func NewStatsHandler(svc StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

func (h *StatsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /stats/reviewers", h.GetReviewerStats)
}

func (h *StatsHandler) GetReviewerStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetReviewerStats(r.Context())
	if err != nil {
		httpcommon.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	resp := ReviewerStatsResponse{
		Stats: make([]ReviewerStatDTO, 0, len(stats)),
	}

	for _, s := range stats {
		resp.Stats = append(resp.Stats, ReviewerStatDTO{
			UserID: s.UserID,
			Count:  s.Count,
		})
	}

	httpcommon.JSONResponse(w, http.StatusOK, resp)
}
