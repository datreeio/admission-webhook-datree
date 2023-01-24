package errorReporter

import (
	"fmt"

	"runtime/debug"

	"github.com/datreeio/datree/pkg/utils"

	"github.com/datreeio/admission-webhook-datree/pkg/clients"
	servicestate "github.com/datreeio/admission-webhook-datree/pkg/serviceState"
)

type ErrorReporterClient interface {
	ReportError(reportCliErrorRequest clients.ReportCliErrorRequest, uri string) (StatusCode int, Error error)
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
	state := servicestate.GetState()
	statusCode, err := reporter.client.ReportError(clients.ReportCliErrorRequest{
		ClientId:       state.ClientId,
		Token:          state.Token,
		ClusterName:    state.ClusterName,
		ClusterUuid:    state.ClusterUuid,
		K8sVersion:     state.K8sVersion,
		PolicyName:     state.PolicyName,
		IsEnforceMode:  state.IsEnforceMode,
		ServiceVersion: state.ServiceVersion,
		ServiceType:    state.ServiceType,
		ErrorMessage:   errorMessage,
		StackTrace:     string(debug.Stack()),
	}, uri)

	if err != nil {
		// using fmt.Println instead of logger to avoid circular dependency
		fmt.Println(fmt.Sprintf("ReportError status code: %d, err: %s", statusCode, err.Error()))
	}
}
