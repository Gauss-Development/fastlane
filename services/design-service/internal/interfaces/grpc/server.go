package grpc

import (
	"context"
	"fmt"

	appErrors "design-service/internal/application/errors"
	"design-service/internal/application/services"
	"design-service/pkg/logger"

	designv1 "github.com/nikitashilov/microblog_grpc/proto/design/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"design-service/internal/domain/entities"
	"time"
)

// DesignServer implements designv1.DesignServiceServer.
type DesignServer struct {
	designv1.UnimplementedDesignServiceServer
	svc    *services.DesignService
	logger *logger.Logger
}

func NewDesignServer(svc *services.DesignService, logger *logger.Logger) *DesignServer {
	return &DesignServer{svc: svc, logger: logger}
}

func (s *DesignServer) CreateProject(ctx context.Context, req *designv1.CreateProjectRequest) (*designv1.Project, error) {
	p := req.GetProject()
	if p == nil {
		return nil, status.Error(codes.InvalidArgument, "project is required")
	}
	ent := &entities.Project{
		OwnerID:      p.GetOwnerId(),
		Title:        p.GetTitle(),
		Description:  p.GetDescription(),
		Category:     p.GetCategory(),
		OwnerEmail:   p.GetOwnerEmail(),
		OwnerCompany: p.GetOwnerCompany(),
	}
	created, err := s.svc.CreateProject(ctx, ent)
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoProject(created), nil
}

func (s *DesignServer) GetProject(ctx context.Context, req *designv1.GetProjectRequest) (*designv1.Project, error) {
	if req.GetId() == "" || req.GetActorId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id and actor_id are required")
	}
	p, err := s.svc.GetProject(ctx, req.GetId(), req.GetActorId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoProject(p), nil
}

func (s *DesignServer) ListProjects(ctx context.Context, req *designv1.ListProjectsRequest) (*designv1.ListProjectsResponse, error) {
	if req.GetOwnerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "owner_id is required")
	}
	ps, total, err := s.svc.ListProjects(ctx, req.GetOwnerId(), req.GetStatus(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	out := make([]*designv1.Project, 0, len(ps))
	for _, p := range ps {
		out = append(out, toProtoProject(p))
	}
	return &designv1.ListProjectsResponse{Projects: out, Total: total}, nil
}

func (s *DesignServer) RequestUploadURL(ctx context.Context, req *designv1.RequestUploadURLRequest) (*designv1.UploadURLResponse, error) {
	if req.GetProjectId() == "" || req.GetActorId() == "" || req.GetFilename() == "" || req.GetKind() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id, actor_id, kind and filename are required")
	}
	f, uploadURL, expiresIn, err := s.svc.RequestUploadURL(ctx,
		req.GetProjectId(), req.GetActorId(), req.GetKind(), req.GetFilename(), req.GetContentType())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return &designv1.UploadURLResponse{
		FileId:      f.ID,
		UploadUrl:   uploadURL,
		ObjectKey:   f.ObjectKey,
		ExpiresInS:  int32(expiresIn),
	}, nil
}

func (s *DesignServer) ConfirmUpload(ctx context.Context, req *designv1.ConfirmUploadRequest) (*designv1.DesignFile, error) {
	if req.GetFileId() == "" || req.GetActorId() == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id and actor_id are required")
	}
	f, err := s.svc.ConfirmUpload(ctx, req.GetFileId(), req.GetActorId(), req.GetContentSha256(), req.GetSizeBytes())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoFile(f), nil
}

func (s *DesignServer) ListFiles(ctx context.Context, req *designv1.ListFilesRequest) (*designv1.ListFilesResponse, error) {
	if req.GetProjectId() == "" || req.GetActorId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and actor_id are required")
	}
	fs, err := s.svc.ListFiles(ctx, req.GetProjectId(), req.GetActorId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	out := make([]*designv1.DesignFile, 0, len(fs))
	for _, f := range fs {
		out = append(out, toProtoFile(f))
	}
	return &designv1.ListFilesResponse{Files: out}, nil
}

func (s *DesignServer) RequestDownloadURL(ctx context.Context, req *designv1.RequestDownloadURLRequest) (*designv1.DownloadURLResponse, error) {
	if req.GetFileId() == "" || req.GetActorId() == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id and actor_id are required")
	}
	downloadURL, filename, expiresIn, err := s.svc.RequestDownloadURL(ctx, req.GetFileId(), req.GetActorId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return &designv1.DownloadURLResponse{
		DownloadUrl: downloadURL,
		Filename:    filename,
		ExpiresInS:  int32(expiresIn),
	}, nil
}

func (s *DesignServer) AcceptNDA(ctx context.Context, req *designv1.AcceptNDARequest) (*designv1.NDA, error) {
	if req.GetProjectId() == "" || req.GetManufacturerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and manufacturer_id are required")
	}
	nda, err := s.svc.AcceptNDA(ctx, req.GetProjectId(), req.GetManufacturerId(), req.GetNdaVersion(), req.GetAcceptedIp())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoNDA(nda), nil
}

func (s *DesignServer) InviteManufacturer(ctx context.Context, req *designv1.InviteManufacturerRequest) (*designv1.NDA, error) {
	if req.GetProjectId() == "" || req.GetManufacturerId() == "" || req.GetActorId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id, manufacturer_id and actor_id are required")
	}
	nda, err := s.svc.InviteManufacturer(ctx, req.GetProjectId(), req.GetManufacturerId(), req.GetActorId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoNDA(nda), nil
}

func (s *DesignServer) GetNDAStatus(ctx context.Context, req *designv1.GetNDAStatusRequest) (*designv1.NDA, error) {
	if req.GetProjectId() == "" || req.GetManufacturerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and manufacturer_id are required")
	}
	nda, err := s.svc.GetNDAStatus(ctx, req.GetProjectId(), req.GetManufacturerId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoNDA(nda), nil
}

func (s *DesignServer) HealthCheck(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *DesignServer) toGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if de, ok := err.(*appErrors.DesignError); ok {
		switch de.Code {
		case appErrors.ErrProjectNotFound.Code, appErrors.ErrFileNotFound.Code, appErrors.ErrNDANotFound.Code:
			return status.Error(codes.NotFound, de.Message)
		case appErrors.ErrUnauthorizedAccess.Code:
			return status.Error(codes.PermissionDenied, de.Message)
		case appErrors.ErrInvalidRequest.Code:
			return status.Error(codes.InvalidArgument, de.Message)
		case appErrors.ErrServiceUnavailable.Code:
			return status.Error(codes.Internal, de.Message)
		default:
			return status.Error(codes.Internal, de.Message)
		}
	}
	s.logger.Error(fmt.Sprintf("design grpc: unexpected error: %v", err))
	return status.Error(codes.Internal, "internal server error")
}

func toProtoProject(p *entities.Project) *designv1.Project {
	if p == nil {
		return nil
	}
	return &designv1.Project{
		Id:           p.ID,
		OwnerId:      p.OwnerID,
		Title:        p.Title,
		Description:  p.Description,
		Category:     p.Category,
		Status:       p.Status,
		OwnerEmail:   p.OwnerEmail,
		OwnerCompany: p.OwnerCompany,
		CreatedAt:    toTimestamp(p.CreatedAt),
		UpdatedAt:    toTimestamp(p.UpdatedAt),
	}
}

func toProtoFile(f *entities.DesignFile) *designv1.DesignFile {
	if f == nil {
		return nil
	}
	return &designv1.DesignFile{
		Id:            f.ID,
		ProjectId:     f.ProjectID,
		Kind:          f.Kind,
		Filename:      f.Filename,
		Version:       f.Version,
		ContentSha256: f.ContentSHA256,
		ObjectKey:     f.ObjectKey,
		SizeBytes:     f.SizeBytes,
		ContentType:   f.ContentType,
		UploadedBy:    f.UploadedBy,
		Status:        f.Status,
		CreatedAt:     toTimestamp(f.CreatedAt),
	}
}

func toProtoNDA(n *entities.NDA) *designv1.NDA {
	if n == nil {
		return nil
	}
	out := &designv1.NDA{
		Id:             n.ID,
		ProjectId:      n.ProjectID,
		ManufacturerId: n.ManufacturerID,
		Status:         n.Status,
		NdaVersion:     n.NDAVersion,
		AcceptedIp:     n.AcceptedIP,
		CreatedAt:      toTimestamp(n.CreatedAt),
	}
	if n.AcceptedAt != nil {
		out.AcceptedAt = toTimestamp(*n.AcceptedAt)
	}
	return out
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
