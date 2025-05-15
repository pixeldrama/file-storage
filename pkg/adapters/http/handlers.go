package http

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/benjamin/file-storage-go/pkg/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateUploadJobRequest struct {
	Filename string `json:"filename" binding:"required"`
}

type Handlers struct {
	fileStorage domain.FileStorage
	jobRepo     domain.UploadJobRepository
}

func NewHandlers(fileStorage domain.FileStorage, jobRepo domain.UploadJobRepository) *Handlers {
	return &Handlers{
		fileStorage: fileStorage,
		jobRepo:     jobRepo,
	}
}

func (h *Handlers) CreateUploadJob(c *gin.Context) {
	var req CreateUploadJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	jobID := uuid.New().String()
	now := time.Now()

	job := &domain.UploadJob{
		ID:        jobID,
		Filename:  req.Filename,
		Status:    "PENDING",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.jobRepo.Create(c.Request.Context(), job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

func (h *Handlers) GetUploadJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	job, err := h.jobRepo.Get(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload job not found"})
		return
	}

	if job.Status == "COMPLETED" && job.FileID != "" {
		c.Header("Location", fmt.Sprintf("/api/files/%s", job.FileID))
	}

	c.JSON(http.StatusOK, job)
}

func (h *Handlers) UploadFile(c *gin.Context) {
	jobID := c.Param("jobId")
	job, err := h.jobRepo.Get(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload job not found"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		job.Status = "FAILED"
		job.Error = "No file provided"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusBadRequest, job)
		return
	}

	src, err := file.Open()
	if err != nil {
		job.Status = "FAILED"
		job.Error = "Failed to open file"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusBadRequest, job)
		return
	}
	defer src.Close()

	job.Status = "UPLOADING"
	job.UpdatedAt = time.Now()
	h.jobRepo.Update(c.Request.Context(), job)

	fileID := uuid.New().String()
	err = h.fileStorage.Upload(c.Request.Context(), fileID, src)
	if err != nil {
		job.Status = "FAILED"
		job.Error = err.Error()
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusInternalServerError, job)
		return
	}

	job.Status = "COMPLETED"
	job.FileID = fileID
	job.UpdatedAt = time.Now()
	h.jobRepo.Update(c.Request.Context(), job)

	c.JSON(http.StatusCreated, job)
}

func (h *Handlers) DownloadFile(c *gin.Context) {
	fileID := c.Param("fileId")

	// Retrieve job to get the original filename
	job, err := h.jobRepo.GetByFileID(c.Request.Context(), fileID)
	if err != nil {
		// This typically means a repository-level error, not just "job not found"
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job details: " + err.Error()})
		return
	}
	// If job is nil, it means no job is associated with this fileID, which could be an issue
	// or the file was not uploaded via the job system (if that's possible).
	// For this flow, we assume a job must exist.
	if job == nil {
		// Log this as it might be an unexpected state. For the client, it's still a file not found.
		// Consider if a different error code is more appropriate if a job *should* always exist.
		c.JSON(http.StatusNotFound, gin.H{"error": "File metadata not found"})
		return
	}

	reader, err := h.fileStorage.Download(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found in storage"})
		return
	}
	defer reader.Close()

	filename := job.Filename
	if filename == "" { // Fallback if filename wasn't stored, though it should be
		filename = fileID
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/octet-stream")
	
	// Instead of using c.Stream which uses chunked encoding, copy the file directly to the response
	c.Writer.WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		// Log error but can't really respond with an error status at this point
		// since headers have already been sent
		fmt.Printf("Error copying file to response: %v\n", err)
	}
}
