package domain

type ReviewerStat struct {
	UserID string `json:"user_id"`
	Count  int64  `json:"count"`
}
