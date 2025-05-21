package repository

import (
	"context"
	"fmt"

	"github.com/benjamin/file-storage-go/pkg/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createJobQuery = `
		INSERT INTO upload_jobs (id, filename, status, created_at, updated_at, file_id, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	getJobQuery = `
		SELECT id, filename, status, created_at, updated_at, file_id, error
		FROM upload_jobs
		WHERE id = $1
	`

	updateJobQuery = `
		UPDATE upload_jobs
		SET filename = $1, status = $2, updated_at = $3, file_id = $4, error = $5
		WHERE id = $6
	`

	getJobByFileIDQuery = `
		SELECT id, filename, status, created_at, updated_at, file_id, error
		FROM upload_jobs
		WHERE file_id = $1
	`
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
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

	return &PostgresRepository{
		pool: pool,
	}, nil
}

func (r *PostgresRepository) Create(ctx context.Context, job *domain.UploadJob) error {
	_, err := r.pool.Exec(ctx, createJobQuery,
		job.ID,
		job.Filename,
		job.Status,
		job.CreatedAt,
		job.UpdatedAt,
		job.FileID,
		job.Error,
	)
	if err != nil {
		return fmt.Errorf("failed to create upload job: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Get(ctx context.Context, jobID string) (*domain.UploadJob, error) {
	job := &domain.UploadJob{}
	err := r.pool.QueryRow(ctx, getJobQuery, jobID).Scan(
		&job.ID,
		&job.Filename,
		&job.Status,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.FileID,
		&job.Error,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get upload job: %w", err)
	}
	return job, nil
}

func (r *PostgresRepository) Update(ctx context.Context, job *domain.UploadJob) error {
	result, err := r.pool.Exec(ctx, updateJobQuery,
		job.Filename,
		job.Status,
		job.UpdatedAt,
		job.FileID,
		job.Error,
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update upload job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("upload job not found")
	}

	return nil
}

func (r *PostgresRepository) GetByFileID(ctx context.Context, fileID string) (*domain.UploadJob, error) {
	job := &domain.UploadJob{}
	err := r.pool.QueryRow(ctx, getJobByFileIDQuery, fileID).Scan(
		&job.ID,
		&job.Filename,
		&job.Status,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.FileID,
		&job.Error,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get upload job by file ID: %w", err)
	}
	return job, nil
}

func (r *PostgresRepository) Close() error {
	r.pool.Close()
	return nil
}
