package utils

import (
	"os"
	"testing"
)

func TestGetWebhookServerReplicas(t *testing.T) {
	t.Run("should return 2 when replicas is not set", func(t *testing.T) {
		os.Setenv("WEBHOOK_SERVER_REPLICAS", "")
		replicas := GetWebhookServerReplicas()
		if replicas != 2 {
			t.Errorf("expected replicas to be 2, got %d", replicas)
		}
	})

	t.Run("should return 2 when replicas is not valid", func(t *testing.T) {
		os.Setenv("WEBHOOK_SERVER_REPLICAS", "hh")
		replicas := GetWebhookServerReplicas()
		if replicas != 2 {
			t.Errorf("expected replicas to be 2, got %d", replicas)
		}
	})

	t.Run("should return 5 when replicas is set to 5", func(t *testing.T) {
		os.Setenv("WEBHOOK_SERVER_REPLICAS", "5")
		replicas := GetWebhookServerReplicas()
		if replicas != 5 {
			t.Errorf("expected replicas to be 5, got %d", replicas)
		}
	})
}

func TestGetWebhookSelector(t *testing.T) {
	t.Run("should return default selector when selector is not set", func(t *testing.T) {
		selector := GetWebhookSelector()
		if selector != "admission.datree/validate" {
			t.Errorf("expected selector to be admission.datree/validate, got %s", selector)
		}
	})

	t.Run("should return selector when selector is set", func(t *testing.T) {
		os.Setenv("WEBHOOK_SELECTOR", "my-selector")
		selector := GetWebhookSelector()
		if selector != "my-selector" {
			t.Errorf("expected selector to be my-selector, got %s", selector)
		}
	})
}

func TestGetWebhookNamespace(t *testing.T) {
	t.Run("should return default namespace when namespace is not set", func(t *testing.T) {
		namespace := GetWebhookNamespace()
		if namespace != "datree" {
			t.Errorf("expected namespace to be datree, got %s", namespace)
		}
	})

	t.Run("should return namespace when namespace is set", func(t *testing.T) {
		os.Setenv("WEBHOOK_NAMESPACE", "my-namespace")
		namespace := GetWebhookNamespace()
		if namespace != "my-namespace" {
			t.Errorf("expected namespace to be my-namespace, got %s", namespace)
		}
	})
}

func TestGetWebhookServiceName(t *testing.T) {
	t.Run("should return default service name when service name is not set", func(t *testing.T) {
		serviceName := GetWebhookServiceName()
		if serviceName != "datree-webhook-server" {
			t.Errorf("expected service name to be datree-webhook-server, got %s", serviceName)
		}
	})

	t.Run("should return service name when service name is set", func(t *testing.T) {
		os.Setenv("WEBHOOK_SERVICE", "my-service")
		serviceName := GetWebhookServiceName()
		if serviceName != "my-service" {
			t.Errorf("expected service name to be my-service, got %s", serviceName)
		}
	})
}
