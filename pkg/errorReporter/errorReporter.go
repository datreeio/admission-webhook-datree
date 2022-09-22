package errorReporter

import (
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"os"

	"runtime/debug"

	"github.com/datreeio/datree/cmd"
	"github.com/datreeio/datree/pkg/localConfig"
	"github.com/datreeio/datree/pkg/utils"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"

	"github.com/datreeio/datree/pkg/cliClient"
)

type LocalConfig interface {
	GetLocalConfiguration() (*localConfig.LocalConfig, error)
}

type CliClient interface {
	ReportCliError(reportCliErrorRequest cliClient.ReportCliErrorRequest, uri string) (StatusCode int, Error error)
}

type ErrorReporter struct {
	config LocalConfig
	client CliClient
}

func NewErrorReporter(client CliClient, localConfig LocalConfig) *ErrorReporter {
	return &ErrorReporter{
		client: client,
		config: localConfig,
	}
}

func (reporter *ErrorReporter) ReportPanicError(panicErr interface{}) {
	reporter.ReportError(panicErr, "/report-webhook-panic-error")
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
		logger.LogUtil(fmt.Sprintf("ReportError status code: %d, err: %s", statusCode, err.Error()))
	}
}

func (reporter *ErrorReporter) getLocalConfig() (unknownLocalConfig *localConfig.LocalConfig) {
	unknownLocalConfig = &localConfig.LocalConfig{ClientId: "unknown", Token: "unknown"}
	defer func() {
		_ = recover()

	}()

	config, err := reporter.config.GetLocalConfiguration()
	if err != nil {
		return unknownLocalConfig
	} else {
		return config
	}

}
