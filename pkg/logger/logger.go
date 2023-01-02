package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"go.uber.org/zap"
	admission "k8s.io/api/admission/v1"
)

// most of our logs are in the following places:
// 1. webhook start up
// 2. incoming request
// 3. outgoing request
// 4. errors

// Logger - instructions to get the logs are under /guides/developer-guide.md
type Logger struct {
	zapLogger     *zap.SugaredLogger
	requestId     string
	errorReporter *errorReporter.ErrorReporter
}

func New(requestId string, errorReporter *errorReporter.ErrorReporter) Logger {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync() // flushes buffer, if any
	sugar := zapLogger.Sugar()

	return Logger{zapLogger: sugar, requestId: requestId, errorReporter: errorReporter}
}

func (l *Logger) LogError(message string) {
	l.zapLogger.Errorw(message,
		// Structured context as loosely typed key-value pairs.
		"requestId", l.requestId,
	)
}

func (l *Logger) LogAndReportUnexpectedError(message string) {
	l.LogError(message)
	l.errorReporter.ReportUnexpectedError(errors.New(message))
}

func (l *Logger) LogIncoming(admissionReview *admission.AdmissionReview) {
	l.logInfo(admissionReview, "incoming")
}
func (l *Logger) LogOutgoing(admissionReview *admission.AdmissionReview, isSkipped bool) {
	l.logInfo(outgoingLog{
		AdmissionReview: admissionReview,
		IsSkipped:       isSkipped,
	}, "outgoing")
}

type outgoingLog struct {
	AdmissionReview *admission.AdmissionReview
	IsSkipped       bool
}

func (l *Logger) LogInfo(objectToLog any) {
	l.logInfo(objectToLog, "")
}

// LogUtil this method creates a new logger instance on every call, and does not have a requestId
// please prefer using the logger instance from the context instead
func LogUtil(msg string) {
	logger := New("", nil)
	logger.LogInfo(msg)
}

func (l *Logger) logInfo(objectToLog any, requestDirection string) {
	l.zapLogger.Infow(l.objectToJson(objectToLog),
		// Structured context as loosely typed key-value pairs.
		"requestId", l.requestId,
		"requestDirection", requestDirection)
}
func (l *Logger) objectToJson(object any) string {
	result, err := json.Marshal(object)
	if err != nil {
		l.LogError(fmt.Sprintf("failed to convert object to JSON, error: %s", err))
		return ""
	}
	return string(result)
}
