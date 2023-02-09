package main

import (
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newZapLogger(file *os.File) *zap.Logger {
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	cores := []zapcore.Core{
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderConfig),
			zapcore.Lock(os.Stdout),
			zapcore.DebugLevel,
		),
	}

	if file != nil {
		fileEncoderConfig := zap.NewProductionEncoderConfig()
		fileEncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		// add file encoder to cores
		cores = append(
			cores,
			zapcore.NewCore(
				zapcore.NewJSONEncoder(fileEncoderConfig),
				zapcore.Lock(file),
				zapcore.InfoLevel,
			),
		)
	}

	return zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

// esCustomLogger implements the elastictransport.Logger interface.
type esCustomLogger struct {
	*zap.Logger
}

// LogRoundTrip prints the information about request and response.
func (l *esCustomLogger) LogRoundTrip(
	req *http.Request,
	res *http.Response,
	err error,
	start time.Time,
	dur time.Duration,
) error {
	var (
		lvl  zapcore.Level
		nReq int64
		nRes int64
	)

	// Set log level.
	//
	switch {
	case err != nil:
		lvl = zapcore.ErrorLevel
	case res != nil && res.StatusCode > 0 && res.StatusCode < 300:
		lvl = zapcore.InfoLevel
	case res != nil && res.StatusCode > 299 && res.StatusCode < 500:
		lvl = zapcore.WarnLevel
	case res != nil && res.StatusCode > 499:
		lvl = zapcore.ErrorLevel
	default:
		lvl = zapcore.ErrorLevel
	}

	// Count number of bytes in request and response.
	//
	if req != nil && req.Body != nil && req.Body != http.NoBody {
		nReq, _ = io.Copy(io.Discard, req.Body)
	}
	if res != nil && res.Body != nil && res.Body != http.NoBody {
		nRes, _ = io.Copy(io.Discard, res.Body)
	}

	// Log event.
	//
	if ce := l.Check(lvl, req.URL.String()); ce != nil {
		ce.Write(
			zap.String("method", req.Method),
			zap.Int("status_code", res.StatusCode),
			zap.Duration("duration", dur),
			zap.Int64("req_bytes", nReq),
			zap.Int64("res_bytes", nRes),
		)
	}

	return nil
}

// RequestBodyEnabled makes the client pass request body to logger
func (l *esCustomLogger) RequestBodyEnabled() bool { return true }

// RequestBodyEnabled makes the client pass response body to logger
func (l *esCustomLogger) ResponseBodyEnabled() bool { return true }
