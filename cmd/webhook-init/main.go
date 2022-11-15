package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	// create validation webhook config
	datreeValidationWebhookConfig, err := castDatreeValidationWebhookConfig()
	if err != nil {
		loggerUtil.Logf("failed cast webhook configuration, err: %v", err)
		return
	}

	// create k8s client
	kubeClient, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		loggerUtil.Logf("failed to set k8s go -client, err: %v", err)
		return
	}
	k8sClient := newK8sClient(kubeClient, "datree")

	// wait for webhook-server pods to be ready
	err = k8sClient.waitPodRunning(context.Background(), "app=datree-webhook-server")
	if err != nil {
		loggerUtil.Logf("failed to wait for pods, err: %v", err)
		return
	}

	// create validation webhook, if one already exists delete before creating
	loggerUtil.Log("creating ValidatingWebhook...")
	err = k8sClient.deleteExistingValidationAdmissionWebhook("datree-webhook")
	if err != nil {
		loggerUtil.Logf("failed to delete existed validation webhook config, err: %v", err)
		return
	}

	err = k8sClient.createValidationWebhookConfig(datreeValidationWebhookConfig)
	if err != nil {
		loggerUtil.Logf("failed to create validation webhook config, err: %v", err)
	}

	loggerUtil.Log("Horray! Succesfully initiaded datree validation admission webhook.")
	os.Exit(0)
}

func castDatreeValidationWebhookConfig() (*validationWebhookConfig, error) {
	webhookServiceName, isFound := os.LookupEnv("WEBHOOK_SERVICE")
	if !isFound {
		return nil, fmt.Errorf("required environment variable WEBHOOK_SERVICE is missing")
	}

	webhookNamespace, isFound := os.LookupEnv("WEBHOOK_NAMESPACE")
	if !isFound {
		return nil, fmt.Errorf("required environment variable WEBHOOK_NAMESPACE is missing")
	}

	webhookSelector, isFound := os.LookupEnv("WEBHOOK_SELECTOR")
	if !isFound {
		return nil, fmt.Errorf("required environment variable WEBHOOK_SELECTOR is missing")
	}

	// read CA bundle
	caBundle, err := readCABundle()
	if err != nil {
		return nil, fmt.Errorf("failed to read caBundle, err: %v", err)
	}

	return &validationWebhookConfig{
		name:        "datree-webhook",
		serviceName: webhookServiceName,
		namesapce:   webhookNamespace,
		caBundle:    caBundle,
		selector:    webhookSelector,
	}, nil
}

func readCABundle() ([]byte, error) {
	certPath := filepath.Join(`/etc/webhook/certs`, `ca-bundle.pem`)
	caPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	return caPEM, nil
}
