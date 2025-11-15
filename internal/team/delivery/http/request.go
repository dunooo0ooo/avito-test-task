package http

type AddTeamRequest struct {
	Name    string          `json:"team_name"`
	Members []TeamMemberDTO `json:"members"`
}

type GetTeamRequest struct {
	Name string `json:"team_name"`
}
