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

	manufacturerv1 "github.com/nikitashilov/microblog_grpc/proto/manufacturer/v1"
)

// ManufacturerClientAPI is the minimal client surface used by ManufacturerHandler (testable).
type ManufacturerClientAPI interface {
	CreateManufacturer(ctx context.Context, m *manufacturerv1.Manufacturer) (*manufacturerv1.Manufacturer, error)
	GetManufacturer(ctx context.Context, id string) (*manufacturerv1.Manufacturer, error)
	GetManufacturerByUser(ctx context.Context, userID string) (*manufacturerv1.Manufacturer, error)
	UpdateManufacturer(ctx context.Context, m *manufacturerv1.Manufacturer, actorID string) (*manufacturerv1.Manufacturer, error)
	ListManufacturers(ctx context.Context, req *manufacturerv1.ListManufacturersRequest) (*manufacturerv1.ListManufacturersResponse, error)
	VerifyManufacturer(ctx context.Context, id string, verified bool, actorID string) (*manufacturerv1.Manufacturer, error)
}

type ManufacturerHandler struct {
	manufacturerClient ManufacturerClientAPI
	logger             *logger.Logger
}

func NewManufacturerHandler(manufacturerClient ManufacturerClientAPI, logger *logger.Logger) *ManufacturerHandler {
	return &ManufacturerHandler{manufacturerClient: manufacturerClient, logger: logger}
}

// manufacturerBody carries the mutable fields for create + update. Array fields
// are JSON string arrays; server-owned fields (id, status, verified, stats,
// timestamps) are ignored here.
type manufacturerBody struct {
	Name            string   `json:"name"`
	NameZh          string   `json:"name_zh"`
	City            string   `json:"city"`
	Cluster         string   `json:"cluster"`
	Description     string   `json:"description"`
	Website         string   `json:"website"`
	ServiceTypes    []string `json:"service_types"`
	AssemblyTypes   []string `json:"assembly_types"`
	MinLayers       int32    `json:"min_layers"`
	MaxLayers       int32    `json:"max_layers"`
	Materials       []string `json:"materials"`
	SurfaceFinishes []string `json:"surface_finishes"`
	MinOrderQty     int32    `json:"min_order_qty"`
	MaxOrderQty     int32    `json:"max_order_qty"`
	LeadTimeDays    int32    `json:"lead_time_days"`
	MonthlyCapacity int32    `json:"monthly_capacity"`
	SmallestPackage string   `json:"smallest_package"`
	Certifications  []string `json:"certifications"`
	ContactEmail    string   `json:"contact_email"`
	ContactWechat   string   `json:"contact_wechat"`
}

func (b manufacturerBody) toProto() *manufacturerv1.Manufacturer {
	return &manufacturerv1.Manufacturer{
		Name:            strings.TrimSpace(b.Name),
		NameZh:          b.NameZh,
		City:            b.City,
		Cluster:         b.Cluster,
		Description:     b.Description,
		Website:         b.Website,
		ServiceTypes:    b.ServiceTypes,
		AssemblyTypes:   b.AssemblyTypes,
		MinLayers:       b.MinLayers,
		MaxLayers:       b.MaxLayers,
		Materials:       b.Materials,
		SurfaceFinishes: b.SurfaceFinishes,
		MinOrderQty:     b.MinOrderQty,
		MaxOrderQty:     b.MaxOrderQty,
		LeadTimeDays:    b.LeadTimeDays,
		MonthlyCapacity: b.MonthlyCapacity,
		SmallestPackage: b.SmallestPackage,
		Certifications:  b.Certifications,
		ContactEmail:    b.ContactEmail,
		ContactWechat:   b.ContactWechat,
	}
}

// CreateManufacturer handles POST /api/v1/manufacturers.
func (h *ManufacturerHandler) CreateManufacturer(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var body manufacturerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_NAME", "name is required")
		return
	}

	m := body.toProto()
	m.UserId = userID.(string)

	created, err := h.manufacturerClient.CreateManufacturer(c.Request.Context(), m)
	if err != nil {
		h.handleManufacturerError(c, err, "CREATE_MANUFACTURER_FAILED", "Failed to create manufacturer")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Manufacturer created successfully", manufacturerToMap(created))
}

// GetManufacturer handles GET /api/v1/manufacturers/:id.
func (h *ManufacturerHandler) GetManufacturer(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	m, err := h.manufacturerClient.GetManufacturer(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleManufacturerError(c, err, "GET_MANUFACTURER_FAILED", "Failed to retrieve manufacturer")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Manufacturer retrieved successfully", manufacturerToMap(m))
}

// GetMyManufacturer handles GET /api/v1/manufacturer-profile.
func (h *ManufacturerHandler) GetMyManufacturer(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	m, err := h.manufacturerClient.GetManufacturerByUser(c.Request.Context(), userID.(string))
	if err != nil {
		h.handleManufacturerError(c, err, "GET_MANUFACTURER_PROFILE_FAILED", "Failed to retrieve manufacturer profile")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Manufacturer profile retrieved successfully", manufacturerToMap(m))
}

// ListManufacturers handles GET /api/v1/manufacturers.
func (h *ManufacturerHandler) ListManufacturers(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	limit := parseQueryInt(c, "limit", 20, 1, 100)
	offset := parseQueryInt(c, "offset", 0, 0, 1<<30)

	resp, err := h.manufacturerClient.ListManufacturers(c.Request.Context(), &manufacturerv1.ListManufacturersRequest{
		Limit:        limit,
		Offset:       offset,
		Cluster:      c.Query("cluster"),
		ServiceType:  c.Query("service_type"),
		AssemblyType: c.Query("assembly_type"),
		Material:     c.Query("material"),
		VerifiedOnly: c.Query("verified_only") == "true",
		MinLayersGte: parseQueryInt(c, "min_layers_gte", 0, 0, 1<<20),
	})
	if err != nil {
		h.handleManufacturerError(c, err, "LIST_MANUFACTURERS_FAILED", "Failed to list manufacturers")
		return
	}

	manufacturers := make([]map[string]interface{}, 0, len(resp.GetManufacturers()))
	for _, m := range resp.GetManufacturers() {
		manufacturers = append(manufacturers, manufacturerToMap(m))
	}
	utils.SuccessResponse(c, http.StatusOK, "Manufacturers retrieved successfully", map[string]interface{}{
		"manufacturers": manufacturers,
		"total":         resp.GetTotal(),
	})
}

// UpdateManufacturer handles PUT /api/v1/manufacturers/:id.
func (h *ManufacturerHandler) UpdateManufacturer(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var body manufacturerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_NAME", "name is required")
		return
	}

	m := body.toProto()
	m.Id = c.Param("id")

	updated, err := h.manufacturerClient.UpdateManufacturer(c.Request.Context(), m, userID.(string))
	if err != nil {
		h.handleManufacturerError(c, err, "UPDATE_MANUFACTURER_FAILED", "Failed to update manufacturer")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Manufacturer updated successfully", manufacturerToMap(updated))
}

type verifyManufacturerBody struct {
	Verified bool `json:"verified"`
}

// VerifyManufacturer handles POST /api/v1/manufacturers/:id/verify.
func (h *ManufacturerHandler) VerifyManufacturer(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var body verifyManufacturerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}

	m, err := h.manufacturerClient.VerifyManufacturer(c.Request.Context(), c.Param("id"), body.Verified, userID.(string))
	if err != nil {
		h.handleManufacturerError(c, err, "VERIFY_MANUFACTURER_FAILED", "Failed to verify manufacturer")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Manufacturer verification updated successfully", manufacturerToMap(m))
}

func (h *ManufacturerHandler) handleManufacturerError(c *gin.Context, err error, fallbackCode, fallbackMessage string) {
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
			utils.ErrorResponse(c, http.StatusServiceUnavailable, "CATALOG_UNAVAILABLE", "Catalog service temporarily unavailable")
			return
		}
	}
	h.logger.Error(fallbackMessage + ": " + err.Error())
	utils.ErrorResponse(c, http.StatusInternalServerError, fallbackCode, fallbackMessage)
}

func manufacturerToMap(m *manufacturerv1.Manufacturer) map[string]interface{} {
	if m == nil {
		return nil
	}
	return map[string]interface{}{
		"id":               m.GetId(),
		"user_id":          m.GetUserId(),
		"name":             m.GetName(),
		"name_zh":          m.GetNameZh(),
		"city":             m.GetCity(),
		"country":          m.GetCountry(),
		"cluster":          m.GetCluster(),
		"description":      m.GetDescription(),
		"website":          m.GetWebsite(),
		"service_types":    m.GetServiceTypes(),
		"assembly_types":   m.GetAssemblyTypes(),
		"min_layers":       m.GetMinLayers(),
		"max_layers":       m.GetMaxLayers(),
		"materials":        m.GetMaterials(),
		"surface_finishes": m.GetSurfaceFinishes(),
		"min_order_qty":    m.GetMinOrderQty(),
		"max_order_qty":    m.GetMaxOrderQty(),
		"lead_time_days":   m.GetLeadTimeDays(),
		"monthly_capacity": m.GetMonthlyCapacity(),
		"smallest_package": m.GetSmallestPackage(),
		"certifications":   m.GetCertifications(),
		"verified":         m.GetVerified(),
		"verified_at":      timestampString(m.GetVerifiedAt()),
		"rating":           m.GetRating(),
		"order_count":      m.GetOrderCount(),
		"on_time_rate":     m.GetOnTimeRate(),
		"contact_email":    m.GetContactEmail(),
		"contact_wechat":   m.GetContactWechat(),
		"status":           m.GetStatus(),
		"created_at":       timestampString(m.GetCreatedAt()),
		"updated_at":       timestampString(m.GetUpdatedAt()),
	}
}
