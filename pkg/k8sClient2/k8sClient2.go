package k8sClient2

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

type k8sClientInterface interface {
	doesValidatingWebhookConfigurationExist() (any, error)
	applyValidatingWebhookConfiguration() (any, error)
}

type k8sClient struct {
	clientset *kubernetes.Clientset
}

func NewK8sClient() (*k8sClient, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientsetInstance, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &k8sClient{
		clientset: clientsetInstance,
	}, nil
}

func (kc *k8sClient) ActivateValidatingWebhookConfiguration(certPath string) error {
	certificateContent, readFileError := os.ReadFile(certPath)
	if readFileError != nil {
		return readFileError
	}x

	result, err := kc.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.TODO(), "datree-webhook", metav1.GetOptions{})
	if err != nil {
		return err
	}

	// update the CABundle from PLACEHOLDER to the actual certificate from cert-manager
	result.Webhooks[0].ClientConfig.CABundle = certificateContent

	// remove the match expression at index 1, which is responsible for disabling the webhook
	matchExpressions := result.Webhooks[0].NamespaceSelector.MatchExpressions
	if len(matchExpressions) > 1 {
		result.Webhooks[0].NamespaceSelector.MatchExpressions = append(matchExpressions[:1], matchExpressions[2:]...)
	}

	_, err = kc.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(context.TODO(), result, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (kc *k8sClient) applyValidatingWebhookConfiguration() error {
	return nil
}
