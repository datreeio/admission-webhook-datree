package logger

import (
	"fmt"
	"go.uber.org/zap"
	v1 "k8s.io/api/admission/v1"
)

type Logger struct {
	zapLogger *zap.SugaredLogger
	requestId string
}

// enum iota

type RequestDirection string

const (
	incoming = "dfdf" // a network error while validating the resource
	Skipped  = "df"   // resource has been skipped, for example if its kind was not found and the user added the --ignore-missing-schemas flag
)

const (
	RequestDirectionInbound  = "inbound"
	RequestDirectionOutbound = "outbound"
)

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

func (l *Logger) LogIncoming(admissionReview *v1.AdmissionReview) {
	l.logInfo(admissionReview, "incoming")
}
func (l *Logger) LogOutgoing(admissionReview *v1.AdmissionReview, isSkipped bool) {
	l.logInfo(outgoingLog{
		AdmissionReview: admissionReview,
		IsSkipped:       isSkipped,
	}, "outgoing")
}

type outgoingLog struct {
	AdmissionReview *v1.AdmissionReview
	IsSkipped       bool
}

func (l *Logger) LogInfo(object any) {
	l.logInfo(object, "mid-request")
}

func (l *Logger) logInfo(object any, requestDirection string) {

	l.zapLogger.Infow("info log",
		// Structured context as loosely typed key-value pairs.
		"requestId", l.requestId,
		"requestDirection", requestDirection,
		"message", object)

	// when querying the logs, start by with this:
	// const logsDump = fs.readFileSync('./path/to/logs/file.txt', 'utf8')
	// const logsArray = logsDump.split("___DATREE_SEPARATOR___")
	// const logsArrayWithoutLastEmptyElement = logsArray.slice(0, logsArray.length - 1)
	// const logsArrayWithParsedJson = logsArrayWithoutLastEmptyElement.map(log => JSON.parse(log, null, 2))
	//
	// and then start querying the logs with plain JS, e.g.:
	// logsArrayWithParsedJson.filter(log => log.requestId === "some-request-id")
	// logsArrayWithParsedJson.filter(log => log.message.someProperty === "some-value")

	//logWithRequestId := LogWithMetadata{
	//	RequestId: l.requestId,
	//	Message:   object,
	//}
	//
	//result, err := json.Marshal(logWithRequestId)
	//if err != nil {
	//	LogUtil(fmt.Sprintf("failed to convert logWithRequestId to JSON, error: %s", err))
	//	return
	//}
	//LogUtil(string(result))
}

func LogUtil(msg string) {
	fmt.Printf("%s___DATREE_SEPARATOR___", msg)
}

// TODO remove all the unnecessary logs
// TODO add a log level for errors
// TODO add a log level for info
// TODO consider using Zap
