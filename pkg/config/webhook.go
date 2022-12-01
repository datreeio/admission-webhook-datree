package config

import (
	"os"
	"strconv"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	admission "k8s.io/api/admission/v1"
	"sigs.k8s.io/yaml"
)

const (
	datreeWebhookConfigMapName       = "datree-webhook-config"
	datreeValidatingWebhookName      = "datree-webhook"
	datreeValidatingWebhookSkipLabel = "admission.datree/skip"
)

func IsConfigurationChangeEvent(request *admission.AdmissionRequest) bool {
	return request.Resource.Resource == "configmaps" && request.Name == datreeWebhookConfigMapName
}

func GetDatreeValidatingWebhookName() string {
	return datreeValidatingWebhookName
}

func GetDatreeValidatingWebhookServiceName() string {
	return getEnvWithFallback("WEBHOOK_SERVICE_NAME", "datree-webhook-server")
}

func GetDatreeValidatingWebhookNamespace() string {
	return getEnvWithFallback("WEBHOOK_NAMESPACE", "datree")
}

func GetDatreeValidatingWebhookLabels() map[string]string {
	return map[string]string{datreeValidatingWebhookSkipLabel: "true"}
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

type configuration struct {
	Detach bool `json:"enable"`
}

func parseWebhookConfiguration(request *admission.AdmissionRequest) (*configuration, error) {
	config := &configuration{}

	var resource map[string]interface{}
	yamlObj, _ := yaml.JSONToYAML(request.Object.Raw)
	err := yaml.Unmarshal(yamlObj, &resource)
	if err != nil {
		return nil, err
	}

	configMapData := resource["data"].(map[string]interface{})
	if configMapData != nil {
		config.Detach = configMapData["detach"] == "true"
	}

	return config, nil
}
