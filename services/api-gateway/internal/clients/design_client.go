package clients

import (
	"context"
	"fmt"
	"time"

	"api-gateway/internal/config"
	"api-gateway/pkg/logger"

	designv1 "github.com/nikitashilov/microblog_grpc/proto/design/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/emptypb"
)

const defaultDesignTimeout = 15 * time.Second

// DesignClient talks to the DesignService hosted by design-service (projects,
// design files, and per-manufacturer NDAs).
type DesignClient struct {
	conn   *grpc.ClientConn
	client designv1.DesignServiceClient
	logger *logger.Logger
}

func NewDesignClient(addr string, tlsCfg config.GRPCTLSConfig, logger *logger.Logger) (*DesignClient, error) {
	creds, err := buildClientTransportCredentials(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("build design client transport credentials: %w", err)
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                keepaliveTime,
			Timeout:             keepaliveTimeout,
			PermitWithoutStream: keepalivePermitWithoutStream,
		}),
		grpc.WithUnaryInterceptor(unaryClientLoggingInterceptor(logger)),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to design gRPC service: %w", err)
	}
	return &DesignClient{
		conn:   conn,
		client: designv1.NewDesignServiceClient(conn),
		logger: logger,
	}, nil
}

func (c *DesignClient) CreateProject(ctx context.Context, project *designv1.Project) (*designv1.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.CreateProject(ctx, &designv1.CreateProjectRequest{Project: project})
}

func (c *DesignClient) GetProject(ctx context.Context, id, actorID string) (*designv1.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.GetProject(ctx, &designv1.GetProjectRequest{Id: id, ActorId: actorID})
}

func (c *DesignClient) ListProjects(ctx context.Context, ownerID, status string, limit, offset int32) (*designv1.ListProjectsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.ListProjects(ctx, &designv1.ListProjectsRequest{
		OwnerId: ownerID,
		Status:  status,
		Limit:   limit,
		Offset:  offset,
	})
}

func (c *DesignClient) RequestUploadURL(ctx context.Context, projectID, actorID, kind, filename, contentType string) (*designv1.UploadURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.RequestUploadURL(ctx, &designv1.RequestUploadURLRequest{
		ProjectId:   projectID,
		ActorId:     actorID,
		Kind:        kind,
		Filename:    filename,
		ContentType: contentType,
	})
}

func (c *DesignClient) ConfirmUpload(ctx context.Context, fileID, actorID, contentSha256 string, sizeBytes int64) (*designv1.DesignFile, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.ConfirmUpload(ctx, &designv1.ConfirmUploadRequest{
		FileId:        fileID,
		ActorId:       actorID,
		ContentSha256: contentSha256,
		SizeBytes:     sizeBytes,
	})
}

func (c *DesignClient) ListFiles(ctx context.Context, projectID, actorID string) (*designv1.ListFilesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.ListFiles(ctx, &designv1.ListFilesRequest{ProjectId: projectID, ActorId: actorID})
}

func (c *DesignClient) RequestDownloadURL(ctx context.Context, fileID, actorID string) (*designv1.DownloadURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.RequestDownloadURL(ctx, &designv1.RequestDownloadURLRequest{FileId: fileID, ActorId: actorID})
}

func (c *DesignClient) AcceptNDA(ctx context.Context, projectID, manufacturerID, ndaVersion, acceptedIP string) (*designv1.NDA, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.AcceptNDA(ctx, &designv1.AcceptNDARequest{
		ProjectId:      projectID,
		ManufacturerId: manufacturerID,
		NdaVersion:     ndaVersion,
		AcceptedIp:     acceptedIP,
	})
}

func (c *DesignClient) GetNDAStatus(ctx context.Context, projectID, manufacturerID string) (*designv1.NDA, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.GetNDAStatus(ctx, &designv1.GetNDAStatusRequest{ProjectId: projectID, ManufacturerId: manufacturerID})
}

func (c *DesignClient) InviteManufacturer(ctx context.Context, projectID, manufacturerID, actorID string) (*designv1.NDA, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDesignTimeout)
	defer cancel()
	return c.client.InviteManufacturer(ctx, &designv1.InviteManufacturerRequest{
		ProjectId:      projectID,
		ManufacturerId: manufacturerID,
		ActorId:        actorID,
	})
}

func (c *DesignClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := c.client.HealthCheck(ctx, &emptypb.Empty{})
	return err
}

func (c *DesignClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
