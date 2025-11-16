package http

type ReviewerStatDTO struct {
	UserID string `json:"user_id"`
	Count  int64  `json:"count"`
}

type ReviewerStatsResponse struct {
	Stats []ReviewerStatDTO `json:"stats"`
}
