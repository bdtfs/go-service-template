package notifier

type operationCreatedRequest struct {
	ExternalID string `json:"external_id"`
	UserID     int64  `json:"user_id"`
	Amount     int64  `json:"amount"`
}
