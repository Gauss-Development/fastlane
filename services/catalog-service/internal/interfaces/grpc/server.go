package grpc

import (
	"context"
	"fmt"
	"time"

	appErrors "catalog-service/internal/application/errors"
	"catalog-service/internal/application/services"
	"catalog-service/internal/domain/entities"
	"catalog-service/internal/domain/repositories"
	"catalog-service/pkg/logger"

	manufacturerv1 "github.com/nikitashilov/microblog_grpc/proto/manufacturer/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ManufacturerServer implements manufacturerv1.ManufacturerServiceServer.
type ManufacturerServer struct {
	manufacturerv1.UnimplementedManufacturerServiceServer
	svc    *services.ManufacturerService
	logger *logger.Logger
}

func NewManufacturerServer(svc *services.ManufacturerService, log *logger.Logger) *ManufacturerServer {
	return &ManufacturerServer{svc: svc, logger: log}
}

func (s *ManufacturerServer) CreateManufacturer(ctx context.Context, req *manufacturerv1.CreateManufacturerRequest) (*manufacturerv1.Manufacturer, error) {
	p := req.GetManufacturer()
	if p == nil {
		return nil, status.Error(codes.InvalidArgument, "manufacturer is required")
	}
	m, err := s.svc.CreateManufacturer(ctx, fromProto(p))
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProto(m), nil
}

func (s *ManufacturerServer) GetManufacturer(ctx context.Context, req *manufacturerv1.GetManufacturerRequest) (*manufacturerv1.Manufacturer, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	m, err := s.svc.GetManufacturer(ctx, req.GetId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProto(m), nil
}

func (s *ManufacturerServer) GetManufacturerByUser(ctx context.Context, req *manufacturerv1.GetManufacturerByUserRequest) (*manufacturerv1.Manufacturer, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	m, err := s.svc.GetManufacturerByUser(ctx, req.GetUserId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProto(m), nil
}

func (s *ManufacturerServer) UpdateManufacturer(ctx context.Context, req *manufacturerv1.UpdateManufacturerRequest) (*manufacturerv1.Manufacturer, error) {
	if req.GetManufacturer() == nil {
		return nil, status.Error(codes.InvalidArgument, "manufacturer is required")
	}
	if req.GetActorId() == "" {
		return nil, status.Error(codes.InvalidArgument, "actor_id is required")
	}
	m, err := s.svc.UpdateManufacturer(ctx, fromProto(req.GetManufacturer()), req.GetActorId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProto(m), nil
}

func (s *ManufacturerServer) ListManufacturers(ctx context.Context, req *manufacturerv1.ListManufacturersRequest) (*manufacturerv1.ListManufacturersResponse, error) {
	ms, total, err := s.svc.ListManufacturers(ctx, repositories.ListFilter{
		Limit:        req.GetLimit(),
		Offset:       req.GetOffset(),
		Cluster:      req.GetCluster(),
		ServiceType:  req.GetServiceType(),
		AssemblyType: req.GetAssemblyType(),
		Material:     req.GetMaterial(),
		VerifiedOnly: req.GetVerifiedOnly(),
		MinLayersGte: req.GetMinLayersGte(),
	})
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	out := make([]*manufacturerv1.Manufacturer, 0, len(ms))
	for _, m := range ms {
		out = append(out, toProto(m))
	}
	return &manufacturerv1.ListManufacturersResponse{Manufacturers: out, Total: total}, nil
}

func (s *ManufacturerServer) VerifyManufacturer(ctx context.Context, req *manufacturerv1.VerifyManufacturerRequest) (*manufacturerv1.Manufacturer, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	m, err := s.svc.VerifyManufacturer(ctx, req.GetId(), req.GetVerified())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProto(m), nil
}

func (s *ManufacturerServer) HealthCheck(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *ManufacturerServer) toGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if ce, ok := err.(*appErrors.CatalogError); ok {
		switch ce.Code {
		case appErrors.ErrManufacturerNotFound.Code:
			return status.Error(codes.NotFound, ce.Message)
		case appErrors.ErrManufacturerExists.Code:
			return status.Error(codes.AlreadyExists, ce.Message)
		case appErrors.ErrUnauthorizedAccess.Code:
			return status.Error(codes.PermissionDenied, ce.Message)
		case appErrors.ErrInvalidRequest.Code:
			return status.Error(codes.InvalidArgument, ce.Message)
		default:
			return status.Error(codes.Internal, ce.Message)
		}
	}
	s.logger.Error(fmt.Sprintf("catalog grpc: unexpected error: %v", err))
	return status.Error(codes.Internal, "internal server error")
}

func toProto(m *entities.Manufacturer) *manufacturerv1.Manufacturer {
	if m == nil {
		return nil
	}
	p := &manufacturerv1.Manufacturer{
		Id:              m.ID,
		UserId:          m.UserID,
		Name:            m.Name,
		NameZh:          m.NameZh,
		City:            m.City,
		Country:         m.Country,
		Cluster:         m.Cluster,
		Description:     m.Description,
		Website:         m.Website,
		ServiceTypes:    m.ServiceTypes,
		AssemblyTypes:   m.AssemblyTypes,
		MinLayers:       m.MinLayers,
		MaxLayers:       m.MaxLayers,
		Materials:       m.Materials,
		SurfaceFinishes: m.SurfaceFinishes,
		MinOrderQty:     m.MinOrderQty,
		MaxOrderQty:     m.MaxOrderQty,
		LeadTimeDays:    m.LeadTimeDays,
		MonthlyCapacity: m.MonthlyCapacity,
		SmallestPackage: m.SmallestPackage,
		Certifications:  m.Certifications,
		Verified:        m.Verified,
		Rating:          m.Rating,
		OrderCount:      m.OrderCount,
		OnTimeRate:      m.OnTimeRate,
		ContactEmail:    m.ContactEmail,
		ContactWechat:   m.ContactWechat,
		Status:          m.Status,
		CreatedAt:       toTimestamp(m.CreatedAt),
		UpdatedAt:       toTimestamp(m.UpdatedAt),
	}
	if m.VerifiedAt != nil {
		p.VerifiedAt = toTimestamp(*m.VerifiedAt)
	}
	return p
}

func fromProto(p *manufacturerv1.Manufacturer) *entities.Manufacturer {
	m := &entities.Manufacturer{
		ID:              p.GetId(),
		UserID:          p.GetUserId(),
		Name:            p.GetName(),
		NameZh:          p.GetNameZh(),
		City:            p.GetCity(),
		Country:         p.GetCountry(),
		Cluster:         p.GetCluster(),
		Description:     p.GetDescription(),
		Website:         p.GetWebsite(),
		ServiceTypes:    p.GetServiceTypes(),
		AssemblyTypes:   p.GetAssemblyTypes(),
		MinLayers:       p.GetMinLayers(),
		MaxLayers:       p.GetMaxLayers(),
		Materials:       p.GetMaterials(),
		SurfaceFinishes: p.GetSurfaceFinishes(),
		MinOrderQty:     p.GetMinOrderQty(),
		MaxOrderQty:     p.GetMaxOrderQty(),
		LeadTimeDays:    p.GetLeadTimeDays(),
		MonthlyCapacity: p.GetMonthlyCapacity(),
		SmallestPackage: p.GetSmallestPackage(),
		Certifications:  p.GetCertifications(),
		Verified:        p.GetVerified(),
		Rating:          p.GetRating(),
		OrderCount:      p.GetOrderCount(),
		OnTimeRate:      p.GetOnTimeRate(),
		ContactEmail:    p.GetContactEmail(),
		ContactWechat:   p.GetContactWechat(),
		Status:          p.GetStatus(),
	}
	if p.GetVerifiedAt() != nil {
		t := p.GetVerifiedAt().AsTime()
		m.VerifiedAt = &t
	}
	return m
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
