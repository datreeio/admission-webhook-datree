package config

import (
	k8sclient "github.com/datreeio/admission-webhook-datree/pkg/k8s-client"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	admission "k8s.io/api/admission/v1"
)

type kubernetesClient interface {
	DeleteExistingValidatingWebhook(name string) error
	RemoveNamespaceLabels(ns string, labels map[string]string) error
}

type ConfigurationClient struct {
	k8s kubernetesClient
}

func NewConfigurationClient(k8s kubernetesClient) *ConfigurationClient {
	if k8s == nil {
		k8s = k8sclient.New(nil)
	}

	return &ConfigurationClient{
		k8s: k8s,
	}
}

func (c *ConfigurationClient) DeleteWebhook() {
	err := c.k8s.DeleteExistingValidatingWebhook("datree-webhook")
	if err != nil {
		logger.Logf("failed to delete existed validation webhook config, err: %v", err)
	}
}

func (c *ConfigurationClient) HandleConfigurationChange(request *admission.AdmissionRequest) {
	if request.Operation == admission.Delete {
		// config, err := parseWebhookConfiguration(request)
		// if err != nil {
		// 	return
		// }

		// if config.Detach {
		err := c.k8s.DeleteExistingValidatingWebhook(GetDatreeValidatingWebhookName())
		if err != nil {
			logger.Logf("failed to delete existing validating webhook, err: %v", err)
		}

		// 	err = c.k8s.RemoveNamespaceLabel(GetDatreeValidatingWebhookNamespace(), GetDatreeValidatingWebhookLabels())
		// 	if err != nil {
		// 		logger.Logf("failed to remove namespace label, err: %v", err)
		// 	}

		// 	err = c.k8s.RemoveNamespaceLabel("kube-system", GetDatreeValidatingWebhookLabels())
		// 	if err != nil {
		// 		logger.Logf("failed to remove namespace label, err: %v", err)
		// 	}
		// }
	}
}
