package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DocumentsRepository handles database operations for documents
type DocumentsRepository struct {
	pool *pgxpool.Pool
}

// NewDocumentsRepository creates a new documents repository
func NewDocumentsRepository(pool *pgxpool.Pool) *DocumentsRepository {
	return &DocumentsRepository{pool: pool}
}

// GetStatus retrieves the current status of a document
func (r *DocumentsRepository) GetStatus(ctx context.Context, orgID, docID string) (string, error) {
	var status string
	err := r.pool.QueryRow(ctx,
		`SELECT status FROM documents WHERE id=$1 AND organization_id=$2`,
		docID, orgID,
	).Scan(&status)

	if err != nil {
		return "", err
	}

	return status, nil
}

// MarkProcessing marks a document as processing
func (r *DocumentsRepository) MarkProcessing(ctx context.Context, orgID, docID string) error {
	cmd, err := r.pool.Exec(ctx,
		`UPDATE documents SET status='processing', updated_at=now() WHERE id=$1 AND organization_id=$2`,
		docID, orgID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// MarkIndexed marks a document as indexed
func (r *DocumentsRepository) MarkIndexed(ctx context.Context, orgID, docID string) error {
	cmd, err := r.pool.Exec(ctx,
		`UPDATE documents SET status='indexed', updated_at=now() WHERE id=$1 AND organization_id=$2`,
		docID, orgID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// MarkFailed marks a document as failed with an error message
func (r *DocumentsRepository) MarkFailed(ctx context.Context, orgID, docID string, errMsg string) error {
	cmd, err := r.pool.Exec(ctx,
		`UPDATE documents SET status='failed', error_message=$1, updated_at=now() WHERE id=$2 AND organization_id=$3`,
		errMsg, docID, orgID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
