package http

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type GetReviewsRequest struct {
	UserID string `json:"user_id"`
}
