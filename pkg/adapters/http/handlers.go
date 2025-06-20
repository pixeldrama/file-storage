package http

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"file-storage-go/pkg/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadJobReqest struct {
	Filename           string `json:"filename" binding:"required"`
	FileType           string `json:"fileType" binding:"required"`
	LinkedResourceType string `json:"linkedResourceType" binding:"required"`
	LinkedResourceID   string `json:"linkedResourceID" binding:"required"`
}

type Handlers struct {
	fileStorage       domain.FileStorage
	jobRepo           domain.UploadJobRepository
	fileInfoRepo      domain.FileInfoRepository
	fileAuthorization domain.FileAuthorization
}

func NewHandlers(fileStorage domain.FileStorage, jobRepo domain.UploadJobRepository, fileInfoRepo domain.FileInfoRepository, fileAuthorization domain.FileAuthorization) *Handlers {
	return &Handlers{
		fileStorage:       fileStorage,
		jobRepo:           jobRepo,
		fileInfoRepo:      fileInfoRepo,
		fileAuthorization: fileAuthorization,
	}
}

func (h *Handlers) CreateUploadJob(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user ID found in context"})
		return
	}

	jobID := uuid.New().String()
	now := time.Now()

	job := &domain.UploadJob{
		ID:              jobID,
		CreatedByUserId: userID,
		Status:          domain.JobStatusUploading,
		CreatedAt:       now,
		UpdatedAt:       now,
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

	if err := h.validateUserAccess(c, job); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
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

	if err := h.validateUserAccess(c, job); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	var req UploadJobReqest
	if err := c.ShouldBindJSON(&req); err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "Invalid request payload: " + err.Error()
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusBadRequest, ToAPIJob(job))
		return
	}

	userID := c.GetString("userId")
	authorized, err := h.fileAuthorization.AuthorizeUploadFile(userID, req.FileType, req.LinkedResourceType, req.LinkedResourceID)
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "Authorization check failed: " + err.Error()
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusInternalServerError, ToAPIJob(job))
		return
	}
	if !authorized {
		job.Status = domain.JobStatusFailed
		job.Error = "Upload not authorized"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusForbidden, ToAPIJob(job))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "No file provided"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(c.Request.Context(), job)
		c.JSON(http.StatusBadRequest, ToAPIJob(job))
		return
	}

	src, err := fileHeader.Open()
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

	fileInfo := &domain.FileInfo{
		ID:                 fileID,
		Filename:           req.Filename,
		FileType:           req.FileType,
		LinkedResourceType: req.LinkedResourceType,
		LinkedResourceID:   req.LinkedResourceID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.fileInfoRepo.Create(c.Request.Context(), fileInfo); err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "Failed to create file record: " + err.Error()
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

func (h *Handlers) GetFileInfo(c *gin.Context) {
	fileID := c.Param("fileId")
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user ID found in context"})
		return
	}

	authorized, err := h.fileAuthorization.AuthorizeReadFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization check failed: " + err.Error()})
		return
	}
	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Read access not authorized"})
		return
	}

	fileInfo, err := h.fileInfoRepo.Get(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file info: " + err.Error()})
		return
	}

	if fileInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File info not found"})
		return
	}

	c.JSON(http.StatusOK, fileInfo)
}

func (h *Handlers) DownloadFile(c *gin.Context) {
	fileID := c.Param("fileId")
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user ID found in context"})
		return
	}

	authorized, err := h.fileAuthorization.AuthorizeReadFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization check failed: " + err.Error()})
		return
	}
	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Read access not authorized"})
		return
	}

	job, err := h.jobRepo.GetByFileID(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job details: " + err.Error()})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File metadata not found"})
		return
	}

	fileInfo, err := h.fileInfoRepo.Get(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file details: " + err.Error()})
		return
	}

	if fileInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File details not found"})
		return
	}

	reader, err := h.fileStorage.Download(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found in storage"})
		return
	}
	defer reader.Close()

	filename := fileInfo.Filename
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
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user ID found in context"})
		return
	}

	authorized, err := h.fileAuthorization.AuthorizeDeleteFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization check failed: " + err.Error()})
		return
	}
	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Delete access not authorized"})
		return
	}

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

func (h *Handlers) validateUserAccess(c *gin.Context, job *domain.UploadJob) error {
	userID := c.GetString("userId")

	if userID == "" {
		return fmt.Errorf("no user ID found in context")
	}

	if job.CreatedByUserId != userID {
		return fmt.Errorf("access denied: job belongs to different user")
	}

	return nil
}
