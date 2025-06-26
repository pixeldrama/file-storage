package repository

import (
	"context"
	"fmt"

	"file-storage-go/pkg/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createFileInfoQuery = `
		INSERT INTO file_info (id, filename, file_type, linked_resource_type, linked_resource_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	getFileInfoQuery = `
		SELECT id, filename, file_type, linked_resource_type, linked_resource_id, created_at, updated_at
		FROM file_info
		WHERE id = $1
	`

	updateFileInfoQuery = `
		UPDATE file_info
		SET filename = $1, file_type = $2, linked_resource_type = $3, linked_resource_id = $4, updated_at = $5
		WHERE id = $6
	`

	deleteFileInfoQuery = `
		DELETE FROM file_info
		WHERE id = $1
	`
)

type PostgresFileInfoRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresFileInfoRepo(connStr string) (*PostgresFileInfoRepo, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresFileInfoRepo{
		pool: pool,
	}, nil
}

func (r *PostgresFileInfoRepo) Create(ctx context.Context, fileInfo *domain.FileInfo) error {
	_, err := r.pool.Exec(ctx, createFileInfoQuery,
		fileInfo.ID,
		fileInfo.Filename,
		fileInfo.FileType,
		fileInfo.LinkedResourceType,
		fileInfo.LinkedResourceID,
		fileInfo.CreatedAt,
		fileInfo.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create file info: %w", err)
	}
	return nil
}

func (r *PostgresFileInfoRepo) Get(ctx context.Context, fileID string) (*domain.FileInfo, error) {
	fileInfo := &domain.FileInfo{}
	err := r.pool.QueryRow(ctx, getFileInfoQuery, fileID).Scan(
		&fileInfo.ID,
		&fileInfo.Filename,
		&fileInfo.FileType,
		&fileInfo.LinkedResourceType,
		&fileInfo.LinkedResourceID,
		&fileInfo.CreatedAt,
		&fileInfo.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return fileInfo, nil
}

func (r *PostgresFileInfoRepo) Update(ctx context.Context, fileInfo *domain.FileInfo) error {
	result, err := r.pool.Exec(ctx, updateFileInfoQuery,
		fileInfo.Filename,
		fileInfo.FileType,
		fileInfo.LinkedResourceType,
		fileInfo.LinkedResourceID,
		fileInfo.UpdatedAt,
		fileInfo.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update file info: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("file info not found")
	}

	return nil
}

func (r *PostgresFileInfoRepo) Delete(ctx context.Context, fileID string) error {
	result, err := r.pool.Exec(ctx, deleteFileInfoQuery, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file info: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("file info not found")
	}

	return nil
}

func (r *PostgresFileInfoRepo) Close() error {
	r.pool.Close()
	return nil
}
