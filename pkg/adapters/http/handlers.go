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
		Status:    domain.JobStatusUploading,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.jobRepo.Create(c.Request.Context(), job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ToAPIJob(job))
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

	if job.Status == domain.JobStatusCompleted && job.FileID != "" {
		c.Header("Location", fmt.Sprintf("/files/%s", job.FileID))
	}

	c.JSON(http.StatusOK, ToAPIJob(job))
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
		job.Status = domain.JobStatusFailed
		job.Error = "No file provided"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusBadRequest, ToAPIJob(job))
		return
	}

	src, err := file.Open()
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "Failed to open file"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusBadRequest, ToAPIJob(job))
		return
	}
	defer src.Close()

	job.Status = domain.JobStatusUploading
	job.UpdatedAt = time.Now()
	h.jobRepo.Update(c.Request.Context(), job)

	fileID := uuid.New().String()
	err = h.fileStorage.Upload(c.Request.Context(), fileID, src)
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = err.Error()
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusInternalServerError, ToAPIJob(job))
		return
	}

	job.Status = domain.JobStatusVirusCheckPending
	job.FileID = fileID
	job.UpdatedAt = time.Now()
	h.jobRepo.Update(c.Request.Context(), job)

	c.JSON(http.StatusCreated, ToAPIJob(job))
}

func (h *Handlers) DownloadFile(c *gin.Context) {
	fileID := c.Param("fileId")

	job, err := h.jobRepo.GetByFileID(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job details: " + err.Error()})
		return
	}

	if job == nil {
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
	if filename == "" {
		filename = fileID
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/octet-stream")

	c.Writer.WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		fmt.Printf("Error copying file to response: %v\n", err)
	}
}

func (h *Handlers) DeleteFile(c *gin.Context) {
	fileID := c.Param("fileId")

	job, err := h.jobRepo.GetByFileID(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job details: " + err.Error()})
		return
	}

	if job != nil {
		job.Status = domain.JobStatusDeleted
		job.UpdatedAt = time.Now()
		if err := h.jobRepo.Update(c.Request.Context(), job); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job status: " + err.Error()})
			return
		}
	}

	err = h.fileStorage.Delete(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
