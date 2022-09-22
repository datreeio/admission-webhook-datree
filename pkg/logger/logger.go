package logger

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

type Logger struct {
	requestId string
}

type LogWithMetadata struct {
	RequestId    string `json:"requestId"`
	RequestStage string `json:"requestDirection"` // incoming, outgoing or mid-request
	Level        string `json:"level"`            // info, error, debug
	Message      any    `json:"message"`
}

func New(requestId string) Logger {
	return Logger{requestId: requestId}
}

func (l *Logger) Log(object any) {
	// when querying the logs, start by with this:
	// const logsDump = fs.readFileSync('./path/to/logs/file.txt', 'utf8')
	// const logsArray = logsDump.split("___DATREE_SEPARATOR___")
	// const logsArrayWithoutLastEmptyElement = logsArray.slice(0, logsArray.length - 1)
	// const logsArrayWithParsedJson = logsArrayWithoutLastEmptyElement.map(log => JSON.parse(log, null, 2))
	//
	// and then start querying the logs with plain JS, e.g.:
	// logsArrayWithParsedJson.filter(log => log.requestId === "some-request-id")
	// logsArrayWithParsedJson.filter(log => log.message.someProperty === "some-value")

	logWithRequestId := LogWithMetadata{
		RequestId: l.requestId,
		Message:   object,
	}

	result, err := json.Marshal(logWithRequestId)
	if err != nil {
		LogUtil(fmt.Sprintf("failed to convert logWithRequestId to JSON, error: %s", err))
		return
	}
	LogUtil(string(result))
}

func LogUtil(msg string) {
	fmt.Printf("%s___DATREE_SEPARATOR___", msg)
}

// TODO remove all the unnecessary logs
// TODO add a log level for errors
// TODO add a log level for info
// TODO consider using Zap


loggerr, _ := zap.NewProduction()
defer logger.Sync() // flushes buffer, if any
sugar := logger.Sugar()
sugar.Infow("failed to fetch URL",
// Structured context as loosely typed key-value pairs.
"url", url,
"attempt", 3,
"backoff", time.Second,
)
sugar.Infof("Failed to fetch URL: %s", url)
