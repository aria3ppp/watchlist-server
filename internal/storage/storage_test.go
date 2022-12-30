package storage_test

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/storage"
	"github.com/aria3ppp/watchlist-server/internal/storage/storagetestutils"
	"github.com/stretchr/testify/require"
)

func TestPutFile(t *testing.T) {
	require := require.New(t)

	t.Cleanup(teardown)

	m, err := storage.NewMinIO(client)
	require.NoError(err)

	ctx := context.Background()

	// no file

	objectInfos, err := storagetestutils.ListFiles(
		client,
		config.Config.MinIO.Bucket.Image.Name,
		false,
	)
	require.NoError(err)
	require.Equal(0, len(objectInfos))

	files := []struct {
		filename    string
		contentType string
		cid         int
	}{
		{
			filename:    "image-1.webp",
			contentType: "image/webp",
			cid:         1,
		},
		{
			filename:    "image-2.png",
			contentType: "image/png",
			cid:         2,
		},
		{
			filename:    "image-3.jpeg",
			contentType: "image/jpeg",
			cid:         3,
		},
		{
			filename:    "image-4.jpg",
			contentType: "image/jpeg",
			cid:         4,
		},
		{
			filename:    "image-5.jpg",
			contentType: "image/jpeg",
			cid:         5,
		},
	}

	// add files

	fileInfos := make([]fs.FileInfo, len(files))
	filePaths := make([]string, len(files))
	fileUris := make([]string, len(files))

	for i, f := range files {
		imageFile, err := os.Open(filepath.Join("testdata", f.filename))
		require.NoError(err)
		t.Cleanup(func() {
			imageFile.Close()
		})

		fileInfos[i], err = imageFile.Stat()
		require.NoError(err)

		putOptions := &storage.PutOptions{
			Bucket:      config.Config.MinIO.Bucket.Image.Name,
			Category:    config.Config.MinIO.Category.User,
			CategoryID:  files[i].cid,
			Filename:    config.Config.MinIO.Filename.User,
			ContentType: f.contentType,
			Size:        fileInfos[i].Size(),
		}

		fileUris[i], err = m.PutFile(ctx, imageFile, putOptions)

		require.NoError(err)

		filePaths[i] = putOptions.BuildPath()

		splits := strings.Split(fileUris[i], "?")
		require.Equal(2, len(splits))
		require.Equal(
			fmt.Sprintf(
				"/%s/%s",
				config.Config.MinIO.Bucket.Image.Name,
				filePaths[i],
			),
			splits[0],
		)
		require.NotEmpty(splits[1])
	}

	objectInfos, err = storagetestutils.ListFiles(
		client,
		config.Config.MinIO.Bucket.Image.Name,
		false,
	)
	require.NoError(err)
	require.Equal(len(files), len(objectInfos))

	for i, fi := range objectInfos {
		require.NoError(fi.Err)
		require.Equal(fileInfos[i].Size(), fi.Size)
		// require.Equal(contentType, fi.ContentType)
		require.Equal(filePaths[i], fi.Key)

		// check content type on real file

		file, err := storagetestutils.GetFile(
			client,
			config.Config.MinIO.Bucket.Image.Name,
			filePaths[i],
			"",
		)
		require.NoError(err)
		t.Cleanup(func() {
			file.Close()
		})

		fileContentType, err := storagetestutils.FileContenType(file)
		require.NoError(err)
		require.Equal(files[i].contentType, fileContentType)
	}
}

func TestPutFile_overwrite(t *testing.T) {
	require := require.New(t)

	t.Cleanup(teardown)

	m, err := storage.NewMinIO(client)
	require.NoError(err)

	ctx := context.Background()

	// no file

	objectInfos, err := storagetestutils.ListFiles(
		client,
		config.Config.MinIO.Bucket.Image.Name,
		true,
	)
	require.NoError(err)
	require.Equal(0, len(objectInfos))

	files := []struct {
		filename    string
		contentType string
	}{
		{
			filename:    "image-1.webp",
			contentType: "image/webp",
		},
		{
			filename:    "image-2.png",
			contentType: "image/png",
		},
		{
			filename:    "image-3.jpeg",
			contentType: "image/jpeg",
		},
		{
			filename:    "image-4.jpg",
			contentType: "image/jpeg",
		},
		{
			filename:    "image-5.jpg",
			contentType: "image/jpeg",
		},
	}

	// add files

	fileInfos := make([]fs.FileInfo, len(files))
	filePaths := make([]string, len(files))
	fileUris := make([]string, len(files))

	cid := 1
	for i, f := range files {
		imageFile, err := os.Open(filepath.Join("testdata", f.filename))
		require.NoError(err)
		t.Cleanup(func() {
			imageFile.Close()
		})

		fileInfos[i], err = imageFile.Stat()
		require.NoError(err)

		putOptions := &storage.PutOptions{
			Bucket:      config.Config.MinIO.Bucket.Image.Name,
			Category:    config.Config.MinIO.Category.User,
			CategoryID:  cid,
			Filename:    config.Config.MinIO.Filename.User,
			ContentType: f.contentType,
			Size:        fileInfos[i].Size(),
		}

		fileUris[i], err = m.PutFile(ctx, imageFile, putOptions)

		require.NoError(err)

		filePaths[i] = putOptions.BuildPath()

		splits := strings.Split(fileUris[i], "?")
		require.Equal(2, len(splits))
		require.Equal(
			fmt.Sprintf(
				"/%s/%s",
				config.Config.MinIO.Bucket.Image.Name,
				filePaths[i],
			),
			splits[0],
		)
		require.NotEmpty(splits[1])
	}

	objectInfos, err = storagetestutils.ListFiles(
		client,
		config.Config.MinIO.Bucket.Image.Name,
		false,
	)
	require.NoError(err)
	require.Equal(1, len(objectInfos))

	// with versions
	objectInfos, err = storagetestutils.ListFiles(
		client,
		config.Config.MinIO.Bucket.Image.Name,
		true,
	)
	require.NoError(err)
	require.Equal(len(files), len(objectInfos))

	for i, fi := range fileInfos {
		// order of file versions are inverted in comparison to time added
		invertedIndex := len(fileInfos) - 1 - i
		objectInfo := objectInfos[invertedIndex]

		require.NoError(objectInfo.Err)
		require.Equal(fi.Size(), objectInfo.Size)
		// require.Equal(contentType, objectInfo.ContentType)
		require.Equal(filePaths[i], objectInfo.Key)

		// check content type on real file

		file, err := storagetestutils.GetFile(
			client,
			config.Config.MinIO.Bucket.Image.Name,
			filePaths[i],
			strings.Split(strings.Split(fileUris[i], "?")[1], "=")[1],
		)
		require.NoError(err)
		t.Cleanup(func() {
			file.Close()
		})

		fileContentType, err := storagetestutils.FileContenType(file)
		require.NoError(err)
		require.Equal(files[i].contentType, fileContentType)
	}
}
