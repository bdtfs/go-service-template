package notifier

// OperationCreatedIn is the payload sent to the notification service when a new
// operation is created.
type OperationCreatedIn struct {
	ExternalID string `json:"external_id"`
	UserID     int64  `json:"user_id"`
	Amount     int64  `json:"amount"`
}
