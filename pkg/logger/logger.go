package logger

import (
	"encoding/json"
	"fmt"
)

type Logger struct {
	requestId string
}

type LogWithRequestId struct {
	RequestId string `json:"requestId"`
	Message   any    `json:"message"`
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

	logWithRequestId := LogWithRequestId{
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
