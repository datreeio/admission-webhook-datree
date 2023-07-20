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

func New(errorReporter *errorReporter.ErrorReporter) Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	jsonEncoder := zapcore.NewJSONEncoder(config)

	defaultLogLevel := zapcore.DebugLevel

	core := zapcore.NewTee(zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel))

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return Logger{
		zapLogger:     zapLogger,
		errorReporter: errorReporter,
	}
}

func (l *Logger) SetRequestId(requestId string) {
	l.requestId = requestId
}

func (l *Logger) LogDebug(message string) {
	l.zapLogger.Debug(message, zap.String("requestId", l.requestId))
}

func (l *Logger) LogInfo(message string) {
	l.zapLogger.Info(message, zap.String("requestId", l.requestId))
}

func (l *Logger) LogWarn(message string) {
	l.zapLogger.Warn(message, zap.String("requestId", l.requestId))
}

func (l *Logger) LogError(message string) {
	l.zapLogger.Error(message, zap.String("requestId", l.requestId))
}

func (l *Logger) Fatal(message string) {
	l.zapLogger.Fatal(message, zap.String("requestId", l.requestId))
}

func (l *Logger) LogAndReportUnexpectedError(message string) {
	l.LogError(message)
	l.errorReporter.ReportUnexpectedError(errors.New(message))
}

func (l *Logger) LogAdmissionRequest(admissionReview *admission.AdmissionReview, isSkipped bool, direction string) {
	if direction != "incoming" && direction != "outgoing" {
		l.LogError("LogAdmissionRequest: direction must be 'incoming' or 'outgoing'")
	}

	logFields := make(map[string]interface{})
	logFields["requestId"] = l.requestId
	logFields["requestDirection"] = direction
	logFields["isSkipped"] = isSkipped
	logFields["admissionReview"] = admissionReview

	l.zapLogger.Debug("AdmissionRequest", zap.Any("fields", logFields))
}
