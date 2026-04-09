package entity

import "time"

type DocumentStatus string

const (
	DocumentStatusPendingUpload DocumentStatus = "pending_upload"
	DocumentStatusUploaded      DocumentStatus = "uploaded"
	DocumentStatusProcessing    DocumentStatus = "processing"
	DocumentStatusIndexed       DocumentStatus = "indexed"
	DocumentStatusFailed        DocumentStatus = "failed"
)

type Document struct {
	ID          string
	OrgID       string
	FileName    string
	ContentType string
	Status      DocumentStatus
	ObjectKey   string
	SizeBytes   int64
	CreatedAt   time.Time
}
