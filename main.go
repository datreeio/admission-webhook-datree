package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/config"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"github.com/datreeio/admission-webhook-datree/pkg/services"
	"github.com/robfig/cron/v3"

	"github.com/datreeio/admission-webhook-datree/pkg/controllers"
	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"github.com/datreeio/admission-webhook-datree/pkg/k8sMetadataUtil"
	"github.com/datreeio/admission-webhook-datree/pkg/server"
	"github.com/datreeio/datree/pkg/cliClient"
	"github.com/datreeio/datree/pkg/deploymentConfig"
	"github.com/datreeio/datree/pkg/localConfig"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/datreeio/datree/pkg/utils"
)

const DefaultErrExitCode = 1

func main() {
	port := os.Getenv("LISTEN_PORT")
	if port == "" {
		port = "8443"
	}

	start(port)
}

func start(port string) {
	internalLogger := logger.New("")

	basicNetworkValidator := networkValidator.NewNetworkValidator()
	basicCliClient := cliClient.NewCliClient(deploymentConfig.URL, basicNetworkValidator)
	basicLocalConfigClient := localConfig.NewLocalConfigClient(basicCliClient, basicNetworkValidator)
	errorReporter := errorReporter.NewErrorReporter(basicCliClient, basicLocalConfigClient)

	defer func() {
		if panicErr := recover(); panicErr != nil {
			errorReporter.ReportPanicError(panicErr)
			internalLogger.LogError(fmt.Sprintf("Datree Webhook failed to start due to Unexpected error: %s\n", utils.ParseErrorToString(panicErr)))
			os.Exit(DefaultErrExitCode)
		}
	}()

	k8sMetadataUtil.InitK8sMetadataUtil()
	initMetadataLogsCronjob()
	server.InitServerVars()
	certPath, keyPath, err := server.ValidateCertificate()
	if err != nil {
		panic(err)
	}

	validationController := controllers.NewValidationController(errorReporter)
	healthController := controllers.NewHealthController()
	// set routes
	http.HandleFunc("/validate", validationController.Validate)
	http.HandleFunc("/health", healthController.Health)
	http.HandleFunc("/ready", healthController.Ready)

	internalLogger.LogInfo(fmt.Sprintf("server starting in webhook-version: %s", config.WebhookVersion))

	// start server
	if err := http.ListenAndServeTLS(":"+port, certPath, keyPath, nil); err != nil {
		http.ListenAndServe(":"+port, nil)
	}
}

func initMetadataLogsCronjob() {
	cornJob := cron.New(cron.WithLocation(time.UTC))
	cornJob.AddFunc("@every 1h", services.SendMetadataInBatch)
	cornJob.Start()
}
