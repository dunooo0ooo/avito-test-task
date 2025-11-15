package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	teamdomain "github.com/dunooo0ooo/avito-test-task/internal/team/domain"
	"github.com/dunooo0ooo/avito-test-task/pkg/httpcommon"
)

type TeamService interface {
	CreateTeam(ctx context.Context, teamName string, members []teamdomain.TeamMember) (*teamdomain.Team, error)
	GetTeam(ctx context.Context, teamName string) (*teamdomain.Team, error)
}

type TeamHandler struct {
	teams TeamService
}

func NewTeamHandler(teams TeamService) *TeamHandler {
	return &TeamHandler{
		teams: teams,
	}
}

func (h *TeamHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /team/add", h.AddTeam)
	mux.HandleFunc("GET /team/get", h.GetTeam)
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req AddTeamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	members := make([]teamdomain.TeamMember, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, teamdomain.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	team, err := h.teams.CreateTeam(r.Context(), req.Name, members)
	if err != nil {
		switch {
		case errors.Is(err, teamdomain.ErrTeamAlreadyExists):
			httpcommon.JSONError(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
		case errors.Is(err, teamdomain.ErrInternalDatabase):
			httpcommon.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		default:
			httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return
	}

	respMembers := make([]TeamMemberDTO, 0, len(team.Members))
	for _, m := range team.Members {
		respMembers = append(respMembers, TeamMemberDTO{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	resp := AddTeamResponse{
		Team: TeamDTO{
			TeamName: team.TeamName,
			Members:  respMembers,
		},
	}

	httpcommon.JSONResponse(w, http.StatusCreated, resp)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "team_name is required")
		return
	}

	team, err := h.teams.GetTeam(r.Context(), teamName)
	if err != nil {
		switch {
		case errors.Is(err, teamdomain.ErrTeamNotFound):
			httpcommon.JSONError(w, http.StatusNotFound, "NOT_FOUND", "team not found")
		case errors.Is(err, teamdomain.ErrInternalDatabase):
			httpcommon.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		default:
			httpcommon.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return
	}

	respMembers := make([]TeamMemberDTO, 0, len(team.Members))
	for _, m := range team.Members {
		respMembers = append(respMembers, TeamMemberDTO{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	resp := TeamDTO{
		TeamName: team.TeamName,
		Members:  respMembers,
	}

	httpcommon.JSONResponse(w, http.StatusOK, resp)
}
