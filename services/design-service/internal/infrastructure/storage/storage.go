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
	mc         *minio.Client // public endpoint: offline presigning (URL host = browser-reachable)
	internal   *minio.Client // internal endpoint: live ops (StatObject), dialed from the container
	bucket     string
	presignTTL time.Duration
}

func New(endpoint, internalEndpoint, accessKey, secretKey, bucket string, useSSL bool, region string, presignTTL time.Duration) (*Client, error) {
	// Region set => the SDK skips the live GetBucketLocation lookup, so
	// PresignedPutObject/GetObject sign purely offline. The public endpoint is
	// only the host baked into the presigned URL (browser-reachable), never dialed.
	opts := func() *minio.Options {
		return &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: useSSL,
			Region: region,
		}
	}
	mc, err := minio.New(endpoint, opts())
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	// Internal client dials MinIO directly (e.g. minio:9000) for real operations
	// like StatObject — the public endpoint isn't reachable from the container.
	internal, err := minio.New(internalEndpoint, opts())
	if err != nil {
		return nil, fmt.Errorf("init minio internal client: %w", err)
	}
	return &Client{mc: mc, internal: internal, bucket: bucket, presignTTL: presignTTL}, nil
}

// StatObject verifies the object exists in the bucket and returns its size. Uses
// the internal client (a real dial to MinIO), unlike offline presigning.
func (c *Client) StatObject(ctx context.Context, objectKey string) (int64, error) {
	info, err := c.internal.StatObject(ctx, c.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("stat object: %w", err)
	}
	return info.Size, nil
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
