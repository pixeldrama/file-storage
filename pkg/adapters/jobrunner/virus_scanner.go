package jobrunner

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"file-storage-go/pkg/domain"
)

const (
	defaultWorkerCount = 5
	defaultChannelSize = 100
)

type VirusScannerJobRunner struct {
	jobRepo         domain.UploadJobRepository
	fileStorage     domain.FileStorage
	virusChecker    domain.VirusChecker
	workerCount     int
	stuckJobTimeout time.Duration
	metrics         domain.MetricsCollector
}

func NewVirusScannerJobRunner(
	jobRepo domain.UploadJobRepository,
	fileStorage domain.FileStorage,
	virusChecker domain.VirusChecker,
	stuckJobTimeout time.Duration,
	metrics domain.MetricsCollector,
) *VirusScannerJobRunner {
	return &VirusScannerJobRunner{
		jobRepo:         jobRepo,
		fileStorage:     fileStorage,
		virusChecker:    virusChecker,
		workerCount:     defaultWorkerCount,
		stuckJobTimeout: stuckJobTimeout,
		metrics:         metrics,
	}
}

func (r *VirusScannerJobRunner) Start(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var wg sync.WaitGroup
	jobsChan := make(chan *domain.UploadJob, defaultChannelSize)

	for range make([]struct{}, r.workerCount) {
		wg.Add(1)
		go r.worker(ctx, &wg, jobsChan)
	}

	for {
		select {
		case <-ctx.Done():
			close(jobsChan)
			wg.Wait()
			return
		case <-ticker.C:
			if err := r.queuePendingAndStuckJobs(ctx, jobsChan); err != nil {
				log.Printf("Error queueing pending and stuck jobs: %v", err)
			}
		}
	}
}

func (r *VirusScannerJobRunner) worker(ctx context.Context, wg *sync.WaitGroup, jobsChan <-chan *domain.UploadJob) {
	defer wg.Done()

	for job := range jobsChan {
		if err := r.processJob(ctx, job); err != nil {
			log.Printf("Error processing job %s: %v", job.ID, err)
		}
	}
}

func (r *VirusScannerJobRunner) queuePendingAndStuckJobs(ctx context.Context, jobsChan chan<- *domain.UploadJob) error {
	jobs, err := r.jobRepo.GetByStatus(ctx, domain.JobStatusVirusCheckPending)
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	for _, job := range jobs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case jobsChan <- job:
		}
	}

	stuckJobs, err := r.jobRepo.GetByStatus(ctx, domain.JobStatusVirusChecking)
	if err != nil {
		return fmt.Errorf("failed to get checking jobs: %w", err)
	}

	now := time.Now()
	for _, job := range stuckJobs {
		if now.Sub(job.UpdatedAt) > r.stuckJobTimeout {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case jobsChan <- job:
			}
		}
	}

	return nil
}

func (r *VirusScannerJobRunner) processJob(ctx context.Context, job *domain.UploadJob) error {
	startTime := time.Now()

	job.Status = domain.JobStatusVirusChecking
	job.UpdatedAt = time.Now()
	if err := r.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	reader, err := r.fileStorage.Download(ctx, job.FileID)
	if err != nil {
		r.metrics.RecordVirusCheckDuration("error", time.Since(startTime))
		return r.updateJobWithError(ctx, job, fmt.Errorf("failed to download file: %w", err))
	}
	defer reader.Close()

	isClean, err := r.virusChecker.CheckFile(ctx, reader)
	if err != nil {
		r.metrics.RecordVirusCheckDuration("error", time.Since(startTime))
		return r.updateJobWithError(ctx, job, fmt.Errorf("virus check failed: %w", err))
	}

	if !isClean {
		r.metrics.RecordVirusCheckDuration("virus_detected", time.Since(startTime))
		return r.updateJobWithError(ctx, job, fmt.Errorf("file contains malware"))
	}

	job.Status = domain.JobStatusCompleted
	job.UpdatedAt = time.Now()
	r.metrics.RecordVirusCheckDuration("success", time.Since(startTime))

	if err := r.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

func (r *VirusScannerJobRunner) updateJobWithError(ctx context.Context, job *domain.UploadJob, err error) error {
	job.Status = domain.JobStatusFailed
	job.Error = err.Error()
	job.UpdatedAt = time.Now()

	if updateErr := r.jobRepo.Update(ctx, job); updateErr != nil {
		return fmt.Errorf("failed to update job with error: %w (original error: %v)", updateErr, err)
	}

	return err
}
