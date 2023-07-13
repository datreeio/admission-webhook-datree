package k8sClient2

import (
	"context"
	cert_manager "github.com/datreeio/admission-webhook-datree/pkg/cert-manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

type K8sClient struct {
	clientset *kubernetes.Clientset
}

func NewK8sClient() (*K8sClient, error) {
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

	return &K8sClient{
		clientset: clientsetInstance,
	}, nil
}

func (kc *K8sClient) ActivateValidatingWebhookConfiguration() error {
	certificateContent, readFileError := os.ReadFile(cert_manager.CaCertPath)
	if readFileError != nil {
		return readFileError
	}

	existingValidatingWebhookConfiguration, err := kc.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.TODO(), "datree-webhook", metav1.GetOptions{})
	if err != nil {
		return err
	}

	// update the CABundle from PLACEHOLDER to the actual certificate from persistent volume
	existingValidatingWebhookConfiguration.Webhooks[0].ClientConfig.CABundle = certificateContent

	// remove the match expression at index 1, which is responsible for disabling the webhook
	matchExpressions := existingValidatingWebhookConfiguration.Webhooks[0].NamespaceSelector.MatchExpressions
	if len(matchExpressions) > 1 {
		existingValidatingWebhookConfiguration.Webhooks[0].NamespaceSelector.MatchExpressions = append(matchExpressions[:1], matchExpressions[2:]...)
	}

	_, err = kc.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(context.TODO(), existingValidatingWebhookConfiguration, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
