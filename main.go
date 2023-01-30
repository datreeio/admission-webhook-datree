package main

import (
	"errors"
	"fmt"

	"net/http"
	"os"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/admission-webhook-datree/pkg/leaderElection"
	servicestate "github.com/datreeio/admission-webhook-datree/pkg/serviceState"
	v1 "k8s.io/client-go/kubernetes/typed/coordination/v1"

	"github.com/datreeio/admission-webhook-datree/pkg/k8sClient"

	"github.com/datreeio/admission-webhook-datree/pkg/config"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"github.com/datreeio/admission-webhook-datree/pkg/services"
	"github.com/robfig/cron/v3"

	"github.com/datreeio/admission-webhook-datree/pkg/controllers"
	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"github.com/datreeio/admission-webhook-datree/pkg/k8sMetadataUtil"
	"github.com/datreeio/admission-webhook-datree/pkg/server"
	"github.com/datreeio/datree/pkg/deploymentConfig"
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
	state := servicestate.New()
	basicNetworkValidator := networkValidator.NewNetworkValidator()
	basicCliClient := clients.NewCliServiceClient(deploymentConfig.URL, basicNetworkValidator)
	errorReporter := errorReporter.NewErrorReporter(basicCliClient, state)
	internalLogger := logger.New("", errorReporter)

	defer func() {
		if panicErr := recover(); panicErr != nil {
			errorReporter.ReportPanicError(panicErr)
			internalLogger.LogError(fmt.Sprintf("Datree Webhook failed to start due to Unexpected error: %s\n", utils.ParseErrorToString(panicErr)))
			os.Exit(DefaultErrExitCode)
		}
	}()

	k8sClientInstance, err := k8sClient.NewK8sClient()
	var leaderElectionLeaseGetter v1.LeasesGetter = nil
	if err == nil && k8sClientInstance != nil {
		leaderElectionLeaseGetter = k8sClientInstance.CoordinationV1()
	}
	leaderElectionInstance := leaderElection.New(&leaderElectionLeaseGetter, internalLogger)
	k8sMetadataUtilInstance := k8sMetadataUtil.NewK8sMetadataUtil(k8sClientInstance, err, leaderElectionInstance, internalLogger)
	k8sMetadataUtilInstance.InitK8sMetadataUtil()

	clusterUuid, err := k8sMetadataUtilInstance.GetClusterUuid()
	if err != nil {
		formattedErrorMessage := fmt.Sprintf("couldn't get cluster uuid %s", err)
		errorReporter.ReportUnexpectedError(errors.New(formattedErrorMessage))
		internalLogger.LogInfo(formattedErrorMessage)
	}

	k8sVersion, err := k8sMetadataUtilInstance.GetClusterK8sVersion()
	if err != nil {
		formattedErrorMessage := fmt.Sprintf("couldn't get k8s version %s", err)
		errorReporter.ReportUnexpectedError(errors.New(formattedErrorMessage))
		internalLogger.LogInfo(formattedErrorMessage)
	}

	state.SetClusterUuid(clusterUuid)
	state.SetK8sVersion(k8sVersion)

	server.InitServerVars()
	certPath, keyPath, err := server.ValidateCertificate()
	if err != nil {
		panic(err)
	}

	validationController := controllers.NewValidationController(basicCliClient, state, errorReporter, k8sMetadataUtilInstance)
	healthController := controllers.NewHealthController()
	// set routes
	http.HandleFunc("/validate", validationController.Validate)
	http.HandleFunc("/health", healthController.Health)
	http.HandleFunc("/ready", healthController.Ready)

	// use validation service to send metadata in batch
	initMetadataLogsCronjob(validationController.ValidationService)

	internalLogger.LogInfo(fmt.Sprintf("server starting in webhook-version: %s", config.WebhookVersion))

	// start server
	if err := http.ListenAndServeTLS(":"+port, certPath, keyPath, nil); err != nil {
		http.ListenAndServe(":"+port, nil)
	}
}

func initMetadataLogsCronjob(validationService *services.ValidationService) {
	cornJob := cron.New(cron.WithLocation(time.UTC))
	cornJob.AddFunc("@every 1h", validationService.SendMetadataInBatch)
	cornJob.Start()
}
