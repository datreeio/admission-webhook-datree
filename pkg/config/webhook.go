package config

import (
	"os"
	"strconv"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"
)

const (
	datreeWebhookConfigMapName       = "datree-webhook-config"
	datreeValidatingWebhookName      = "datree-webhook"
	datreeValidatingWebhookSkipLabel = "admission.datree/skip"
)

func GetDatreeValidatingWebhookName() string {
	return datreeValidatingWebhookName
}

func GetDatreeValidatingWebhookServiceName() string {
	return getEnvWithFallback("WEBHOOK_SERVICE_NAME", "datree-webhook-server")
}

func GetDatreeValidatingWebhookNamespace() string {
	return getEnvWithFallback("WEBHOOK_NAMESPACE", "datree")
}

func GetDatreeValidatingWebhookNamespaceSelector() string {
	return getEnvWithFallback("WEBHOOK_NAMESPACE_SELECTOR", "admission.datree/validate")
}

func GetDatreeValidatingWebhookServerReplicas() int {
	replicas, err := strconv.Atoi(getEnvWithFallback("WEBHOOK_SERVER_REPLICAS", "2"))

	if err != nil {
		logger.Logf("invalid value for env variable WEBHOOK_SERVER_REPLICAS, err: %v", err)
		return 2
	}

	return replicas
}

func GetDatreeValidatingWebhookPodsSelector() string {
	return getEnvWithFallback("WEBHOOK_POD_SELECTOR", "app=datree-webhook-server")
}

func getEnvWithFallback(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
