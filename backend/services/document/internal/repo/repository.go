package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/entity"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) CreatePending(ctx context.Context, d entity.Document) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO documents (id, organization_id, file_name, content_type, status, object_key, size_bytes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, d.ID, d.OrgID, d.FileName, d.ContentType, d.Status, d.ObjectKey, d.SizeBytes)

	return err
}

func (r *DocumentRepository) MarkUploaded(ctx context.Context, orgID, documentID string, sizeBytes int64) error {
	cmd, err := r.db.Exec(ctx, `
		UPDATE documents
		SET status = 'uploaded', size_bytes = $3, updated_at = now()
		WHERE id = $1 AND organization_id = $2
	`, documentID, orgID, sizeBytes)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *DocumentRepository) ListByOrg(ctx context.Context, orgID string, limit, offset int) ([]entity.Document, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, organization_id, file_name, content_type, status, object_key, size_bytes, created_at
		FROM documents
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.Document, 0)
	for rows.Next() {
		var d entity.Document
		if err := rows.Scan(&d.ID, &d.OrgID, &d.FileName, &d.ContentType, &d.Status, &d.ObjectKey, &d.SizeBytes, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DocumentRepository) GetDocumentByID(ctx context.Context, orgID, documentID string) (*entity.Document, error) {
	var d entity.Document
	err := r.db.QueryRow(ctx, `
		SELECT id, organization_id, file_name, content_type, status, object_key, size_bytes, created_at
		FROM documents
		WHERE id = $1 AND organization_id = $2
	`, documentID, orgID).Scan(&d.ID, &d.OrgID, &d.FileName, &d.ContentType, &d.Status, &d.ObjectKey, &d.SizeBytes, &d.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &d, nil
}
