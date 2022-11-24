package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TODO: implement a log rotation strategy
func newLogger(file *os.File) *zap.Logger {
	var (
		encoderConfig zapcore.EncoderConfig
		encoder       zapcore.Encoder
		writer        zapcore.WriteSyncer
		logLevel      zapcore.Level
	)

	if file != nil {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoder = zapcore.NewJSONEncoder(encoderConfig)
		writer = zapcore.AddSync(file)
		logLevel = zapcore.InfoLevel
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
		writer = zapcore.AddSync(os.Stdout)
		logLevel = zapcore.DebugLevel
	}

	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewCore(encoder, writer, logLevel)

	return zap.New(
		core,
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
		nReq, _ = io.Copy(ioutil.Discard, req.Body)
	}
	if res != nil && res.Body != nil && res.Body != http.NoBody {
		nRes, _ = io.Copy(ioutil.Discard, res.Body)
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
