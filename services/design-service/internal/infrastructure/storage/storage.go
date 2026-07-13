package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps MinIO for presigned-URL generation.
// minio.New() is lazy — no dial on construct — so the service boots even if
// MinIO is unreachable. Presigning is offline signing and likewise needs no
// live connection.
type Client struct {
	mc         *minio.Client
	bucket     string
	presignTTL time.Duration
}

func New(endpoint, accessKey, secretKey, bucket string, useSSL bool, presignTTL time.Duration) (*Client, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	return &Client{mc: mc, bucket: bucket, presignTTL: presignTTL}, nil
}

// PresignPut returns a presigned PUT URL for the given object key and content type.
func (c *Client) PresignPut(ctx context.Context, objectKey, contentType string) (string, int, error) {
	params := url.Values{}
	if contentType != "" {
		params.Set("Content-Type", contentType)
	}
	u, err := c.mc.PresignedPutObject(ctx, c.bucket, objectKey, c.presignTTL)
	if err != nil {
		return "", 0, fmt.Errorf("presign put: %w", err)
	}
	return u.String(), int(c.presignTTL.Seconds()), nil
}

// PresignGet returns a presigned GET URL that forces attachment download with the given filename.
func (c *Client) PresignGet(ctx context.Context, objectKey, filename string) (string, int, error) {
	params := url.Values{}
	params.Set("response-content-disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	u, err := c.mc.PresignedGetObject(ctx, c.bucket, objectKey, c.presignTTL, params)
	if err != nil {
		return "", 0, fmt.Errorf("presign get: %w", err)
	}
	return u.String(), int(c.presignTTL.Seconds()), nil
}
