package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"api-gateway/pkg/logger"
	"api-gateway/pkg/utils"

	designv1 "github.com/nikitashilov/microblog_grpc/proto/design/v1"
)

// DesignClientAPI is the minimal client surface used by ProjectHandler (testable).
type DesignClientAPI interface {
	CreateProject(ctx context.Context, project *designv1.Project) (*designv1.Project, error)
	GetProject(ctx context.Context, id, actorID string) (*designv1.Project, error)
	ListProjects(ctx context.Context, ownerID, status string, limit, offset int32) (*designv1.ListProjectsResponse, error)
	RequestUploadURL(ctx context.Context, projectID, actorID, kind, filename, contentType string) (*designv1.UploadURLResponse, error)
	ConfirmUpload(ctx context.Context, fileID, actorID, contentSha256 string, sizeBytes int64) (*designv1.DesignFile, error)
	ListFiles(ctx context.Context, projectID, actorID string) (*designv1.ListFilesResponse, error)
	RequestDownloadURL(ctx context.Context, fileID, actorID string) (*designv1.DownloadURLResponse, error)
	AcceptNDA(ctx context.Context, projectID, manufacturerID, ndaVersion, acceptedIP string) (*designv1.NDA, error)
	GetNDAStatus(ctx context.Context, projectID, manufacturerID string) (*designv1.NDA, error)
	InviteManufacturer(ctx context.Context, projectID, manufacturerID, actorID string) (*designv1.NDA, error)
}

type ProjectHandler struct {
	designClient DesignClientAPI
	logger       *logger.Logger
}

func NewProjectHandler(designClient DesignClientAPI, logger *logger.Logger) *ProjectHandler {
	return &ProjectHandler{designClient: designClient, logger: logger}
}

type createProjectBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// CreateProject handles POST /api/v1/projects.
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	userEmail, _ := c.Get("userEmail")
	ownerEmail, _ := userEmail.(string)

	var body createProjectBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_TITLE", "title is required")
		return
	}

	project, err := h.designClient.CreateProject(c.Request.Context(), &designv1.Project{
		OwnerId:     userID.(string),
		OwnerEmail:  ownerEmail,
		Title:       strings.TrimSpace(body.Title),
		Description: body.Description,
		Category:    strings.TrimSpace(body.Category),
	})
	if err != nil {
		h.handleDesignError(c, err, "CREATE_PROJECT_FAILED", "Failed to create project")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Project created successfully", projectToMap(project))
}

// GetProject handles GET /api/v1/projects/:id.
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	project, err := h.designClient.GetProject(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		h.handleDesignError(c, err, "GET_PROJECT_FAILED", "Failed to retrieve project")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Project retrieved successfully", projectToMap(project))
}

// ListProjects handles GET /api/v1/projects.
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	limit := parseQueryInt(c, "limit", 20, 1, 100)
	offset := parseQueryInt(c, "offset", 0, 0, 1<<30)

	resp, err := h.designClient.ListProjects(c.Request.Context(), userID.(string), c.Query("status"), limit, offset)
	if err != nil {
		h.handleDesignError(c, err, "LIST_PROJECTS_FAILED", "Failed to list projects")
		return
	}

	projects := make([]map[string]interface{}, 0, len(resp.GetProjects()))
	for _, p := range resp.GetProjects() {
		projects = append(projects, projectToMap(p))
	}
	utils.SuccessResponse(c, http.StatusOK, "Projects retrieved successfully", map[string]interface{}{
		"projects": projects,
		"total":    resp.GetTotal(),
		"limit":    limit,
		"offset":   offset,
	})
}

type uploadURLBody struct {
	Kind        string `json:"kind"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}

// RequestUploadURL handles POST /api/v1/projects/:id/files/upload-url.
func (h *ProjectHandler) RequestUploadURL(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var body uploadURLBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if strings.TrimSpace(body.Kind) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_KIND", "kind is required")
		return
	}
	if strings.TrimSpace(body.Filename) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_FILENAME", "filename is required")
		return
	}

	resp, err := h.designClient.RequestUploadURL(c.Request.Context(), c.Param("id"), userID.(string),
		strings.TrimSpace(body.Kind), strings.TrimSpace(body.Filename), strings.TrimSpace(body.ContentType))
	if err != nil {
		h.handleDesignError(c, err, "UPLOAD_URL_FAILED", "Failed to create upload URL")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Upload URL created successfully", map[string]interface{}{
		"file":      map[string]interface{}{"id": resp.GetFileId()},
		"uploadUrl": resp.GetUploadUrl(),
		"objectKey": resp.GetObjectKey(),
		"expiresIn": resp.GetExpiresInS(),
	})
}

// ListFiles handles GET /api/v1/projects/:id/files.
func (h *ProjectHandler) ListFiles(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	resp, err := h.designClient.ListFiles(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		h.handleDesignError(c, err, "LIST_FILES_FAILED", "Failed to list files")
		return
	}

	files := make([]map[string]interface{}, 0, len(resp.GetFiles()))
	for _, f := range resp.GetFiles() {
		files = append(files, fileToMap(f))
	}
	utils.SuccessResponse(c, http.StatusOK, "Files retrieved successfully", map[string]interface{}{
		"files": files,
	})
}

type acceptNDABody struct {
	NDAVersion string `json:"ndaVersion"`
}

// AcceptNDA handles POST /api/v1/projects/:id/nda/accept.
func (h *ProjectHandler) AcceptNDA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var body acceptNDABody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if strings.TrimSpace(body.NDAVersion) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_NDA_VERSION", "ndaVersion is required")
		return
	}

	nda, err := h.designClient.AcceptNDA(c.Request.Context(), c.Param("id"), userID.(string),
		strings.TrimSpace(body.NDAVersion), c.ClientIP())
	if err != nil {
		h.handleDesignError(c, err, "ACCEPT_NDA_FAILED", "Failed to accept NDA")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "NDA accepted successfully", ndaToMap(nda))
}

type inviteManufacturerBody struct {
	ManufacturerID string `json:"manufacturerId"`
}

// InviteManufacturer handles POST /api/v1/projects/:id/invite.
func (h *ProjectHandler) InviteManufacturer(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var body inviteManufacturerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if strings.TrimSpace(body.ManufacturerID) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_MANUFACTURER_ID", "manufacturerId is required")
		return
	}

	nda, err := h.designClient.InviteManufacturer(c.Request.Context(), c.Param("id"),
		strings.TrimSpace(body.ManufacturerID), userID.(string))
	if err != nil {
		h.handleDesignError(c, err, "INVITE_MANUFACTURER_FAILED", "Failed to invite manufacturer")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Manufacturer invited successfully", ndaToMap(nda))
}

// GetNDAStatus handles GET /api/v1/projects/:id/nda.
func (h *ProjectHandler) GetNDAStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	nda, err := h.designClient.GetNDAStatus(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		h.handleDesignError(c, err, "GET_NDA_FAILED", "Failed to retrieve NDA status")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "NDA status retrieved successfully", ndaToMap(nda))
}

type confirmUploadBody struct {
	// snake_case to match the rest of the API (upload-url's content_type, the
	// confirm response's content_sha256/size_bytes) and the frontend client.
	ContentSha256 string `json:"content_sha256"`
	SizeBytes     int64  `json:"size_bytes"`
}

// ConfirmUpload handles POST /api/v1/files/:fileId/confirm.
func (h *ProjectHandler) ConfirmUpload(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var body confirmUploadBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if strings.TrimSpace(body.ContentSha256) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_SHA256", "contentSha256 is required")
		return
	}
	if body.SizeBytes <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_SIZE", "sizeBytes must be positive")
		return
	}

	file, err := h.designClient.ConfirmUpload(c.Request.Context(), c.Param("fileId"), userID.(string),
		strings.TrimSpace(body.ContentSha256), body.SizeBytes)
	if err != nil {
		h.handleDesignError(c, err, "CONFIRM_UPLOAD_FAILED", "Failed to confirm upload")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Upload confirmed successfully", fileToMap(file))
}

// RequestDownloadURL handles GET /api/v1/files/:fileId/download-url.
func (h *ProjectHandler) RequestDownloadURL(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	resp, err := h.designClient.RequestDownloadURL(c.Request.Context(), c.Param("fileId"), userID.(string))
	if err != nil {
		h.handleDesignError(c, err, "DOWNLOAD_URL_FAILED", "Failed to create download URL")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Download URL created successfully", map[string]interface{}{
		"downloadUrl": resp.GetDownloadUrl(),
		"filename":    resp.GetFilename(),
		"expiresIn":   resp.GetExpiresInS(),
	})
}

func (h *ProjectHandler) handleDesignError(c *gin.Context, err error, fallbackCode, fallbackMessage string) {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			return
		case codes.NotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", st.Message())
			return
		case codes.AlreadyExists:
			utils.ErrorResponse(c, http.StatusConflict, "ALREADY_EXISTS", st.Message())
			return
		case codes.PermissionDenied:
			utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", st.Message())
			return
		case codes.Unauthenticated:
			utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", st.Message())
			return
		case codes.Unavailable:
			utils.ErrorResponse(c, http.StatusServiceUnavailable, "DESIGN_UNAVAILABLE", "Design service temporarily unavailable")
			return
		}
	}
	h.logger.Error(fallbackMessage + ": " + err.Error())
	utils.ErrorResponse(c, http.StatusInternalServerError, fallbackCode, fallbackMessage)
}

func projectToMap(p *designv1.Project) map[string]interface{} {
	if p == nil {
		return nil
	}
	return map[string]interface{}{
		"id":            p.GetId(),
		"owner_id":      p.GetOwnerId(),
		"owner_email":   p.GetOwnerEmail(),
		"owner_company": p.GetOwnerCompany(),
		"title":         p.GetTitle(),
		"description":   p.GetDescription(),
		"category":      p.GetCategory(),
		"status":        p.GetStatus(),
		"created_at":    timestampString(p.GetCreatedAt()),
		"updated_at":    timestampString(p.GetUpdatedAt()),
	}
}

func fileToMap(f *designv1.DesignFile) map[string]interface{} {
	if f == nil {
		return nil
	}
	return map[string]interface{}{
		"id":             f.GetId(),
		"project_id":     f.GetProjectId(),
		"kind":           f.GetKind(),
		"filename":       f.GetFilename(),
		"version":        f.GetVersion(),
		"content_sha256": f.GetContentSha256(),
		"size_bytes":     f.GetSizeBytes(),
		"content_type":   f.GetContentType(),
		"uploaded_by":    f.GetUploadedBy(),
		"status":         f.GetStatus(),
		"created_at":     timestampString(f.GetCreatedAt()),
	}
}

func ndaToMap(n *designv1.NDA) map[string]interface{} {
	if n == nil {
		return nil
	}
	return map[string]interface{}{
		"id":              n.GetId(),
		"project_id":      n.GetProjectId(),
		"manufacturer_id": n.GetManufacturerId(),
		"status":          n.GetStatus(),
		"nda_version":     n.GetNdaVersion(),
		"accepted_ip":     n.GetAcceptedIp(),
		"accepted_at":     timestampString(n.GetAcceptedAt()),
		"created_at":      timestampString(n.GetCreatedAt()),
	}
}
