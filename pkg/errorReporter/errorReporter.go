package errorReporter

import (
	"fmt"

	"runtime/debug"

	"github.com/datreeio/datree/pkg/utils"

	"github.com/datreeio/admission-webhook-datree/pkg/clients"
	servicestate "github.com/datreeio/admission-webhook-datree/pkg/serviceState"
)

type ErrorReporterClient interface {
	ReportError(reportCliErrorRequest clients.ReportErrorRequest, uri string) (StatusCode int, Error error)
}

type ErrorReporter struct {
	client ErrorReporterClient
	state  *servicestate.ServiceState
}

func NewErrorReporter(client ErrorReporterClient, state *servicestate.ServiceState) *ErrorReporter {
	return &ErrorReporter{
		client: client,
		state:  state,
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
	statusCode, err := reporter.client.ReportError(clients.ReportErrorRequest{
		ClientId:       reporter.state.GetClientId(),
		Token:          reporter.state.GetToken(),
		ClusterName:    reporter.state.GetClusterName(),
		ClusterUuid:    reporter.state.GetClusterUuid(),
		K8sVersion:     reporter.state.GetK8sVersion(),
		PolicyName:     reporter.state.GetPolicyName(),
		IsEnforceMode:  reporter.state.GetIsEnforceMode(),
		WebhookVersion: reporter.state.GetServiceVersion(),
		ErrorMessage:   errorMessage,
		StackTrace:     string(debug.Stack()),
	}, uri)

	if err != nil {
		// using fmt.Println instead of logger to avoid circular dependency
		fmt.Printf("ReportError status code: %d, err: %s", statusCode, err.Error())
	}
}
