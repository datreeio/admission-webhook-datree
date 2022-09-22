package logger

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	admission "k8s.io/api/admission/v1"
)

type Logger struct {
	zapLogger *zap.SugaredLogger
	requestId string
}

type LogWithMetadata struct {
	RequestId        string `json:"requestId"`
	RequestDirection string `json:"requestDirection"` // incoming, outgoing or mid-request
	Level            string `json:"level"`            // info, error, debug
	Message          any    `json:"message"`
}

func New(requestId string) Logger {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync() // flushes buffer, if any
	sugar := zapLogger.Sugar()

	return Logger{zapLogger: sugar, requestId: requestId}
}

func (l *Logger) LogError(message string) {
	l.zapLogger.Errorw(message,
		// Structured context as loosely typed key-value pairs.
		"requestId", l.requestId,
	)
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
	l.logInfo(objectToLog, "mid-request")
}

func (l *Logger) logInfo(objectToLog any, requestDirection string) {

	l.zapLogger.Infow(l.objectToJson(objectToLog),
		// Structured context as loosely typed key-value pairs.
		"requestId", l.requestId,
		"requestDirection", requestDirection)

	// to dump all the logs from the last 72 hours, the user should run the following command:
	// for podId in $(kubectl get pods -n datree --output name); do echo $(kubectl logs -n datree --since=72h $podId); done > datree-webhook-logs.txt
}

func LogUtil(msg string) {
	fmt.Println(msg)
}

func (l *Logger) objectToJson(object any) string {
	result, err := json.Marshal(object)
	if err != nil {
		l.LogError(fmt.Sprintf("failed to convert object to JSON, error: %s", err))
		return ""
	}
	return string(result)
}
