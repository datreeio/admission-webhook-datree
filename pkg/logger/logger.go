package logger

import (
	"errors"
	"os"

	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	admission "k8s.io/api/admission/v1"
)

// most of our logs are in the following places:
// 1. webhook start up
// 2. incoming request
// 3. outgoing request
// 4. errors

// Logger - instructions to get the logs are under /guides/developer-guide.md
type Logger struct {
	zapLogger     *zap.Logger
	requestId     string
	errorReporter *errorReporter.ErrorReporter
}

func New(logLevel zapcore.Level, errorReporter *errorReporter.ErrorReporter) Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	jsonEncoder := zapcore.NewJSONEncoder(config)

	core := zapcore.NewTee(zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), logLevel))

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return Logger{
		zapLogger:     zapLogger,
		errorReporter: errorReporter,
	}
}

func (l *Logger) SetRequestId(requestId string) {
	l.requestId = requestId
}

func (l *Logger) LogDebug(message string, data ...any) {
	l.zapLogger.Debug(message, zap.String("requestId", l.requestId), zap.Any("data", data))
}

func (l *Logger) LogInfo(message string, data ...any) {
	l.zapLogger.Info(message, zap.String("requestId", l.requestId), zap.Any("data", data))
}

func (l *Logger) LogWarn(message string, data ...any) {
	l.zapLogger.Warn(message, zap.String("requestId", l.requestId), zap.Any("data", data))
}

func (l *Logger) LogError(message string, data ...any) {
	l.zapLogger.Error(message, zap.String("requestId", l.requestId), zap.Any("data", data))
}

func (l *Logger) Fatal(message string, data ...any) {
	l.zapLogger.Fatal(message, zap.String("requestId", l.requestId), zap.Any("data", data))
}

func (l *Logger) PanicLevel(message string, data ...any) {
	l.zapLogger.Panic(message, zap.String("requestId", l.requestId), zap.Any("data", data))
}

func (l *Logger) LogAndReportUnexpectedError(message string) {
	l.LogError(message)
	l.errorReporter.ReportUnexpectedError(errors.New(message))
}

type LogDirection string

const (
	Incoming LogDirection = "incoming"
	Outgoing LogDirection = "outgoing"
)

func (l *Logger) LogAdmissionRequest(admissionReview *admission.AdmissionReview, isSkipped bool, direction LogDirection) {
	logFields := make(map[string]interface{})
	logFields["requestId"] = l.requestId
	logFields["requestDirection"] = direction
	logFields["isSkipped"] = isSkipped
	logFields["admissionReview"] = admissionReview

	l.zapLogger.Debug("AdmissionRequest", zap.Any("data", logFields))
}
