package k8sClient2

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

func (kc *k8sClient) DoesValidatingWebhookConfigurationExist() (any, error) {

	result, err := kc.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.TODO(), "datree-webhook", metav1.GetOptions{})

	fmt.Println("*******************")
	fmt.Println(result)
	fmt.Println(err)
	fmt.Println("*******************")

	result.Webhooks[0].ClientConfig.CABundle = []byte("bbbbbbbbbb")

	result2, err2 := kc.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(
		context.TODO(),
		result,
		metav1.UpdateOptions{},
	)

	fmt.Println("*******************")
	fmt.Println(result2)
	fmt.Println(err2)
	fmt.Println("*******************")

	return nil, nil
}

func (kc *k8sClient) applyValidatingWebhookConfiguration() error {
	return nil
}
