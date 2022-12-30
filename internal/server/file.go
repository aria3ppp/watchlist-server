package server

import (
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/aria3ppp/watchlist-server/internal/server/response"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// caller is responsible for calling file.Close() when done using it
func (s *Server) getFormFile(
	c echo.Context,
	filename string,
) (file multipart.File, fileHeader *multipart.FileHeader, httpError *echo.HTTPError) {
	var err error

	// check content-type
	ctype := c.Request().Header.Get(echo.HeaderContentType)
	if !strings.HasPrefix(ctype, echo.MIMEMultipartForm) {
		s.logger.Info(
			"server.getFormFile: request content-Type isn't multipart/form-data",
			zap.String("Content-type", ctype),
		)
		httpError = echo.NewHTTPError(
			http.StatusUnsupportedMediaType,
			response.Error(response.StatusUnsupportedMediaType),
		)
		return
	}

	// fetch form file
	fileHeader, err = c.FormFile(filename)
	if err != nil {
		if err == http.ErrMissingFile {
			s.logger.Info(
				"server.getFormFile: missing filename",
				zap.String("filename", filename),
			)
			httpError = echo.NewHTTPError(
				http.StatusBadRequest,
				response.Error(response.StatusMissingFile),
			)
			return
		}

		s.logger.Error(
			"server.getFormFile: failed fetching form file",
			zap.Error(err),
		)
		httpError = echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
		return
	}

	// open file
	file, err = fileHeader.Open()
	if err != nil {
		s.logger.Error(
			"server.getFormFile: failed opening form file",
			zap.Error(err),
		)
		httpError = echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
		return
	}

	return
}

func (s *Server) ensureSupportedFileType(
	file io.ReadSeeker,
	supportedTypes []string,
) (contentType string, httpError *echo.HTTPError) {
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		s.logger.Error(
			"server.ensureSupportedFileType: failed reading file",
			zap.Error(err),
		)
		httpError = echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
		return
	}

	supported := false
	contentType = http.DetectContentType(buff)
	for _, st := range supportedTypes {
		if contentType == st {
			supported = true
			break
		}
	}

	if _, err := file.Seek(0, 0); err != nil {
		s.logger.Error(
			"server.ensureSupportedFileType: failed resetting file seeking point",
			zap.Error(err),
		)
		httpError = echo.NewHTTPError(
			http.StatusInternalServerError,
			response.Error(response.StatusInternalServerError),
		)
		return
	}

	if !supported {
		s.logger.Info(
			"server.ensureSupportedFileType: unsupported media type",
			zap.String("media type", contentType),
		)
		httpError = echo.NewHTTPError(
			http.StatusUnsupportedMediaType,
			response.Error(response.StatusUnsupportedMediaType),
		)
		return
	}

	return
}
