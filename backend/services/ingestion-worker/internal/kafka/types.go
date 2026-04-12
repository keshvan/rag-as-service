package kafka

// DocumentUploadedEvent represents a document upload event from the event queue
type DocumentUploadedEvent struct {
	EventID        string `json:"event_id"`
	EventType      string `json:"event_type"` // "DOCUMENT_UPLOADED"
	DocumentID     string `json:"document_id"`
	OrganizationID string `json:"organization_id"`
	ObjectKey      string `json:"object_key"`
	ContentType    string `json:"content_type"`
	SizeBytes      int64  `json:"size_bytes"`
	OccurredAt     string `json:"occurred_at"` // RFC3339
}
