package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/minio/minio-go/v7"
)

//go:generate mockgen -destination mock_storage/mock_service.go . Service

type Service interface {
	PutFile(
		ctx context.Context,
		file io.Reader,
		options *PutOptions,
	) (uri string, err error)
}

type MinIO struct {
	client *minio.Client
}

var _ Service = &MinIO{}

func NewMinIO(client *minio.Client) (*MinIO, error) {
	ctx := context.Background()
	bucket := config.Config.MinIO.Bucket.Image.Name
	// check bucket exists
	if exists, err := client.BucketExists(ctx, bucket); err != nil {
		return nil, err
	} else if !exists {
		// create bucket
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
		// set bucket policy public
		policy := fmt.Sprintf(`{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::%s/*"],"Sid": ""}]}`, bucket)
		if err := client.SetBucketPolicy(ctx, bucket, policy); err != nil {
			return nil, err
		}
		// enable versioning
		if err := client.EnableVersioning(ctx, bucket); err != nil {
			return nil, err
		}
	}
	return &MinIO{client: client}, nil
}

func (m *MinIO) PutFile(
	ctx context.Context,
	file io.Reader,
	options *PutOptions,
) (uri string, err error) {
	// build path
	path := options.BuildPath()
	// put file
	info, err := m.client.PutObject(
		ctx,
		options.Bucket,
		path,
		file,
		options.Size,
		minio.PutObjectOptions{
			ContentType: options.ContentType,
		},
	)
	if err != nil {
		return "", err
	}
	// build file uri string
	uri = fmt.Sprintf(
		"/%s/%s?versionId=%s",
		options.Bucket,
		path,
		info.VersionID,
	)
	return uri, nil
}
