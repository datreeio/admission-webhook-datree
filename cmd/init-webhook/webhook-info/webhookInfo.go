package webhookinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"
)

func GetWebhookServiceName() string {
	webhookServiceName, isFound := os.LookupEnv("WEBHOOK_SERVICE")
	if !isFound || webhookServiceName == "" {
		logger.LogUtil("required environment variable WEBHOOK_SERVICE is missing")
		return "datree-webhook-server"
	}

	return webhookServiceName
}

func GetWebhookNamespace() string {
	webhookNamespace, isFound := os.LookupEnv("WEBHOOK_NAMESPACE")
	if !isFound || webhookNamespace == "" {
		logger.LogUtil("required environment variable WEBHOOK_NAMESPACE is missing")
		return "datree"
	}

	return webhookNamespace
}

func GetWebhookSelector() string {
	webhookSelector, isFound := os.LookupEnv("WEBHOOK_SELECTOR")
	if !isFound || webhookSelector == "" {
		logger.LogUtil("required environment variable WEBHOOK_SELECTOR is missing")
		return "admission.datree/validate"
	}

	return webhookSelector
}

func GetWebhookCABundle() ([]byte, error) {
	certPath := filepath.Join(`/etc/webhook/certs`, `ca-bundle.pem`)
	caPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	return caPEM, nil
}

func GetWebhookServerReplicas() int {
	replicasStr, isFound := os.LookupEnv("WEBHOOK_SERVER_REPLICAS")
	if !isFound || replicasStr == "" {
		logger.LogUtil("required environment variable WEBHOOK_SERVER_REPLICAS is missing")
		return 2
	}
	replicas, err := strconv.Atoi(replicasStr)
	if err != nil {
		logger.LogUtil(fmt.Sprintf("invalid value for env variable WEBHOOK_SERVER_REPLICAS, err: %v", err))
		return 2
	}

	return replicas
}
