package logger

import (
	"encoding/json"
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
	printLogsSeparator()
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

func (l *Logger) LogInfo(objectToLog any) {
	l.logInfo(objectToLog, "mid-request")
}

func (l *Logger) logInfo(objectToLog any, requestDirection string) {

	l.zapLogger.Infow(l.objectToJson(objectToLog),
		// Structured context as loosely typed key-value pairs.
		"requestId", l.requestId,
		"requestDirection", requestDirection)
	printLogsSeparator()

	// to dump all the logs from the last 72 hours, the user should run the following command:
	// for podId in $(kubectl get pods -n datree --output name); do echo $(kubectl logs -n datree --since=72h $podId); done > datree-webhook-logs.txt
	//
	// when querying the logs, you can start with something like this javascript code:
	// type "node" in the terminal to open the node console
	/*
		const fs = require('fs');
		const items = fs.readFileSync('./datree-webhook-logs.txt', 'utf8')
			.split("\n\r")
			.slice(0, -1)
			.map(JSON.parse)
			.filter((value) => value.requestDirection === 'incoming')
			.map((value) => value.msg)
		console.log(items)
	*/

}

func LogUtil(msg string) {
	fmt.Println(msg)
	printLogsSeparator()
}

func (l *Logger) objectToJson(object any) string {
	result, err := json.Marshal(object)
	if err != nil {
		l.LogError(fmt.Sprintf("failed to convert object to JSON, error: %s", err))
		return ""
	}
	return string(result)
}

func printLogsSeparator() {
	// this is needed in order to parse the logs correctly
	fmt.Print("___DATREE_LOGS_SEPARATOR___")
}
