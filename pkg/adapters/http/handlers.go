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
	jobID := uuid.New().String()
	now := time.Now()

	job := &domain.UploadJob{
		ID:        jobID,
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

	reader, err := h.fileStorage.Download(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	defer reader.Close()

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileID))
	c.Header("Content-Type", "application/octet-stream")
	c.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, reader)
		return err == nil
	})
}
