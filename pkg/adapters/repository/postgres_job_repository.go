package repository

import (
	"context"
	"database/sql"
	"fmt"

	"file-storage-go/pkg/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createJobQuery = `
		INSERT INTO upload_jobs (id, created_by_user_id, status, created_at, updated_at, file_id, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	getJobQuery = `
		SELECT id, created_by_user_id, status, created_at, updated_at, file_id, error
		FROM upload_jobs
		WHERE id = $1
	`

	updateJobQuery = `
		UPDATE upload_jobs
		SET created_by_user_id = $1, status = $2, updated_at = $3, file_id = $4, error = $5
		WHERE id = $6
	`

	getJobByFileIDQuery = `
		SELECT id, created_by_user_id, status, created_at, updated_at, file_id, error
		FROM upload_jobs
		WHERE file_id = $1
	`

	getJobsByStatusQuery = `
		SELECT id, created_by_user_id, status, created_at, updated_at, file_id, error
		FROM upload_jobs
		WHERE status = $1
	`
)

type PostgresJobRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresJobRepo(connStr string) (*PostgresJobRepo, error) {
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

	return &PostgresJobRepo{
		pool: pool,
	}, nil
}

func (r *PostgresJobRepo) Create(ctx context.Context, job *domain.UploadJob) error {
	fileID := r.stringToNull(job.FileID)
	_, err := r.pool.Exec(ctx, createJobQuery,
		job.ID,
		job.CreatedByUserId,
		job.Status,
		job.CreatedAt,
		job.UpdatedAt,
		fileID,
		job.Error,
	)
	if err != nil {
		return fmt.Errorf("failed to create upload job: %w", err)
	}
	return nil
}

func (r *PostgresJobRepo) Get(ctx context.Context, jobID string) (*domain.UploadJob, error) {
	job := &domain.UploadJob{}
	var fileID sql.NullString
	err := r.pool.QueryRow(ctx, getJobQuery, jobID).Scan(
		&job.ID,
		&job.CreatedByUserId,
		&job.Status,
		&job.CreatedAt,
		&job.UpdatedAt,
		&fileID,
		&job.Error,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get upload job: %w", err)
	}
	job.FileID = r.nullToString(fileID)
	return job, nil
}

func (r *PostgresJobRepo) Update(ctx context.Context, job *domain.UploadJob) error {
	fileID := r.stringToNull(job.FileID)
	result, err := r.pool.Exec(ctx, updateJobQuery,
		job.CreatedByUserId,
		job.Status,
		job.UpdatedAt,
		fileID,
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

func (r *PostgresJobRepo) GetByFileID(ctx context.Context, fileID string) (*domain.UploadJob, error) {
	job := &domain.UploadJob{}
	var dbFileID sql.NullString
	err := r.pool.QueryRow(ctx, getJobByFileIDQuery, fileID).Scan(
		&job.ID,
		&job.CreatedByUserId,
		&job.Status,
		&job.CreatedAt,
		&job.UpdatedAt,
		&dbFileID,
		&job.Error,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get upload job by file ID: %w", err)
	}
	job.FileID = r.nullToString(dbFileID)
	return job, nil
}

func (r *PostgresJobRepo) GetByStatus(ctx context.Context, status domain.JobStatus) ([]*domain.UploadJob, error) {
	rows, err := r.pool.Query(ctx, getJobsByStatusQuery, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by status: %w", err)
	}
	defer rows.Close()

	var jobs []*domain.UploadJob
	for rows.Next() {
		job := &domain.UploadJob{}
		var fileID sql.NullString
		err := rows.Scan(
			&job.ID,
			&job.CreatedByUserId,
			&job.Status,
			&job.CreatedAt,
			&job.UpdatedAt,
			&fileID,
			&job.Error,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		job.FileID = r.nullToString(fileID)
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating jobs: %w", err)
	}

	return jobs, nil
}

func (r *PostgresJobRepo) stringToNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func (r *PostgresJobRepo) nullToString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func (r *PostgresJobRepo) Close() error {
	r.pool.Close()
	return nil
}
