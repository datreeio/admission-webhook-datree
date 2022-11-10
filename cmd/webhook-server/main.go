package main

import (
	"context"
	"fmt"

	"github.com/datreeio/admission-webhook-datree/pkg/config"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"github.com/datreeio/admission-webhook-datree/pkg/services"
	"github.com/robfig/cron/v3"
	"k8s.io/client-go/kubernetes"

	"net/http"
	"os"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/controllers"
	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"github.com/datreeio/admission-webhook-datree/pkg/k8sMetadataUtil"
	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
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
	defer func() {
		if panicErr := recover(); panicErr != nil {
			validator := networkValidator.NewNetworkValidator()
			newCliClient := cliClient.NewCliClient(deploymentConfig.URL, validator)
			newLocalConfigClient := localConfig.NewLocalConfigClient(newCliClient, validator)
			reporter := errorReporter.NewErrorReporter(newCliClient, newLocalConfigClient)
			reporter.ReportPanicError(panicErr)
			internalLogger.LogError(fmt.Sprintf("Datree Webhook failed to start due to Unexpected error: %s\n", utils.ParseErrorToString(panicErr)))
			os.Exit(DefaultErrExitCode)
		}
	}()

	loggerUtil.Log("initializing k8s metadata")
	k8sMetadataUtil.InitK8sMetadataUtil()
	initMetadataLogsCronjob()
	server.InitServerVars()
	certPath, keyPath, err := server.ValidateCertificate()
	if err != nil {
		panic(err)
	}

	validationController := controllers.NewValidationController()
	healthController := controllers.NewHealthController()
	// set routes
	http.HandleFunc("/validate", validationController.Validate)
	http.HandleFunc("/health", healthController.Health)
	http.HandleFunc("/ready", healthController.Ready)

	internalLogger.LogInfo(fmt.Sprintf("server starting in webhook-version: %s", config.WebhookVersion))

	// start server
	err = http.ListenAndServeTLS(":"+port, certPath, keyPath, nil)
	if err != nil {
		loggerUtil.Log(err.Error())
		http.ListenAndServe(":"+port, nil)
	}
}

func initMetadataLogsCronjob() {
	cornJob := cron.New(cron.WithLocation(time.UTC))
	cornJob.AddFunc("@every 1h", services.SendMetadataInBatch)
	cornJob.Start()
}

func createValidationWebhookConfig(caCert []byte) error {
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err // panic("failed to set go -client")
	}
	// webhookNamespace, _ := os.LookupEnv("WEBHOOK_NAMESPACE")
	validationCfgName := "datree-webhook"

	path := "/validate"
	sideEffects := admissionregistrationv1.SideEffectClassNone

	validationWebhookConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: validationCfgName,
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name: "webhook-server.datree.svc",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caCert, // CA bundle created earlier
				Service: &admissionregistrationv1.ServiceReference{
					Name:      "datree-webhook-server", // datree-webhook-server
					Namespace: "datree",
					Path:      &path,
				},
			},
			Rules: []admissionregistrationv1.RuleWithOperations{{Operations: []admissionregistrationv1.OperationType{
				admissionregistrationv1.Create,
				admissionregistrationv1.Update,
			},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{"*"},
					APIVersions: []string{"*"},
					Resources:   []string{"*"},
				},
			}},
			SideEffects:             &sideEffects,
			AdmissionReviewVersions: []string{"v1", "v1beta1"},
			TimeoutSeconds:          &[]int32{30}[0],
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{ // only validate pods in namespaces with the label "admission.datree/validate"
						Key:      "admission.datree/validate",
						Operator: metav1.LabelSelectorOpDoesNotExist,
					},
				},
			},
		}},
	}

	if _, err = kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.Background(), validationWebhookConfig, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}
