package errorReporter

import (
	"fmt"
	"os"

	"runtime/debug"

	"github.com/datreeio/datree/cmd"
	"github.com/datreeio/datree/pkg/utils"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"

	"github.com/datreeio/datree/pkg/cliClient"
)

type ErrorReporterClient interface {
	ReportCliError(reportCliErrorRequest cliClient.ReportCliErrorRequest, uri string) (StatusCode int, Error error)
}

type ErrorReporter struct {
	client ErrorReporterClient
}

func NewErrorReporter(client ErrorReporterClient) *ErrorReporter {
	return &ErrorReporter{
		client: client,
	}
}

func (reporter *ErrorReporter) ReportPanicError(panicErr interface{}) {
	reporter.ReportError(panicErr, "/report-webhook-panic-error")
}
func (reporter *ErrorReporter) ReportUnexpectedError(unexpectedError error) {
	reporter.ReportError(unexpectedError, "/report-webhook-unexpected-error")
}

func (reporter *ErrorReporter) ReportError(error interface{}, uri string) {
	errorMessage := utils.ParseErrorToString(error)
	statusCode, err := reporter.client.ReportCliError(cliClient.ReportCliErrorRequest{
		ClientId:     os.Getenv(enums.ClientId),
		Token:        os.Getenv(enums.Token),
		CliVersion:   cmd.CliVersion,
		ErrorMessage: errorMessage,
		StackTrace:   string(debug.Stack()),
	}, uri)

	if err != nil {
		// using fmt.Println instead of logger to avoid circular dependency
		fmt.Println(fmt.Sprintf("ReportError status code: %d, err: %s", statusCode, err.Error()))
	}
}
