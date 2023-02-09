package storagetestutils

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/minio/minio-go/v7"
)

func ListFiles(
	client *minio.Client,
	bucket string,
	withVersions bool,
) (objectInfos []minio.ObjectInfo, err error) {
	ctx := context.Background()

	objects := client.ListObjects(
		ctx,
		bucket,
		minio.ListObjectsOptions{Recursive: true, WithVersions: withVersions},
	)
	for obj := range objects {
		if obj.Err != nil {
			return nil, err
		}
		objectInfos = append(objectInfos, obj)
	}

	return objectInfos, nil
}

type File interface {
	multipart.File
	Stat() (minio.ObjectInfo, error)
}

func GetFile(
	client *minio.Client,
	bucket string,
	name string,
	versionId string,
) (file File, err error) {
	ctx := context.Background()

	obj, err := client.GetObject(
		ctx,
		bucket,
		name,
		minio.GetObjectOptions{VersionID: versionId},
	)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func FileContenType(file io.ReadSeeker) (string, error) {
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		return "", err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}
	return http.DetectContentType(buf), nil
}

func DeleteBucketWait(
	client *minio.Client,
	timeout, cooldown time.Duration,
	bucket ...string,
) error {
	ctx := context.Background()

	for _, b := range bucket {
		err := client.RemoveBucketWithOptions(
			ctx,
			b,
			minio.RemoveBucketOptions{ForceDelete: true},
		)
		if err != nil {
			return err
		}
	}

	err := testutils.WaitUntil(
		func() (done bool, err error) {
			for _, b := range bucket {
				if exists, err := client.BucketExists(ctx, b); err != nil {
					return false, err
				} else if exists {
					return false, nil
				}
			}
			return true, nil
		},
		timeout,
		cooldown,
	)

	return err
}
