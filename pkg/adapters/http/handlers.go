package http

import (
	"fmt"
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
	ctx := c.Request.Context()
	jobID := c.Param("jobId")
	userID := c.GetString("userId")

	job, err := h.jobRepo.Get(ctx, jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job"})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	if err := h.validateUserAccess(c, job); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	var req UploadJobReqest
	if err := c.ShouldBindJSON(&req); err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "Invalid request format"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(ctx, job)
		c.JSON(http.StatusBadRequest, ToAPIJob(job))
		return
	}

	authorized, err := h.fileAuthorization.CanUploadFile(userID, req.FileType, req.LinkedResourceType, req.LinkedResourceID)
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "Authorization check failed"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(ctx, job)
		c.JSON(http.StatusInternalServerError, ToAPIJob(job))
		return
	}
	if !authorized {
		job.Status = domain.JobStatusFailed
		job.Error = "Upload not authorized"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(ctx, job)
		c.JSON(http.StatusForbidden, ToAPIJob(job))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "No file provided"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(ctx, job)
		c.JSON(http.StatusBadRequest, ToAPIJob(job))
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "Failed to open file"
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(ctx, job)
		c.JSON(http.StatusBadRequest, ToAPIJob(job))
		return
	}
	defer src.Close()

	job.Status = domain.JobStatusUploading
	job.UpdatedAt = time.Now()
	h.jobRepo.Update(ctx, job)

	fileID := uuid.New().String()
	err = h.fileStorage.Upload(ctx, fileID, src)
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = err.Error()
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(ctx, job)
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

	if err := h.fileInfoRepo.Create(ctx, fileInfo); err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "Failed to create file record: " + err.Error()
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(ctx, job)
		c.JSON(http.StatusInternalServerError, ToAPIJob(job))
		return
	}

	if err := h.fileAuthorization.CreateFileAuthorization(fileID, req.FileType, req.LinkedResourceID, req.LinkedResourceType); err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = "Failed to authorize file: " + err.Error()
		job.UpdatedAt = time.Now()
		h.jobRepo.Update(ctx, job)
		c.JSON(http.StatusInternalServerError, ToAPIJob(job))
		return
	}

	job.Status = domain.JobStatusVirusCheckPending
	job.FileID = fileID
	job.UpdatedAt = time.Now()
	h.jobRepo.Update(ctx, job)

	c.JSON(http.StatusCreated, ToAPIJob(job))
}

func (h *Handlers) GetFileInfo(c *gin.Context) {
	ctx := c.Request.Context()
	fileID := c.Param("fileId")
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user ID found in context"})
		return
	}

	authorized, err := h.fileAuthorization.CanReadFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization check failed"})
		return
	}
	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	fileInfo, err := h.fileInfoRepo.Get(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
		return
	}
	if fileInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.JSON(http.StatusOK, fileInfo)
}

func (h *Handlers) DownloadFile(c *gin.Context) {
	ctx := c.Request.Context()
	fileID := c.Param("fileId")
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user ID found in context"})
		return
	}

	authorized, err := h.fileAuthorization.CanReadFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization check failed"})
		return
	}
	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	fileInfo, err := h.fileInfoRepo.Get(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
		return
	}
	if fileInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	reader, err := h.fileStorage.Download(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Filename))
	c.DataFromReader(http.StatusOK, -1, fileInfo.FileType, reader, nil)
}

func (h *Handlers) DeleteFile(c *gin.Context) {
	ctx := c.Request.Context()
	fileID := c.Param("fileId")
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user ID found in context"})
		return
	}

	authorized, err := h.fileAuthorization.CanDeleteFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization check failed"})
		return
	}
	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	job, err := h.jobRepo.GetByFileID(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job details"})
		return
	}

	if job != nil {
		job.Status = domain.JobStatusDeleted
		job.UpdatedAt = time.Now()
		if err := h.jobRepo.Update(ctx, job); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job status"})
			return
		}
	}

	fileInfo, err := h.fileInfoRepo.Get(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch file info for authorization removal"})
		return
	}

	if fileInfo != nil {
		if err := h.fileAuthorization.RemoveFileAuthorization(fileInfo.ID, fileInfo.FileType, fileInfo.LinkedResourceID, fileInfo.LinkedResourceType); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove file authorization"})
			return
		}
	}

	err = h.fileStorage.Delete(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
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
