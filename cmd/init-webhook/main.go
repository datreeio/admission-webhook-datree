package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/datreeio/admission-webhook-datree/pkg/config"
	k8sclient "github.com/datreeio/admission-webhook-datree/pkg/k8s-client"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func main() {
	k8sClient := k8sclient.New(nil)
	if k8sClient == nil {
		fmt.Print("failed to set k8s go -client, err")
	}

	err := InitWebhook(k8sClient)
	if err != nil {
		logger.Logf("failed to init webhook, err: %v", err)
	} else {
		logger.Logf("horray! succesfully created datree validating admission webhook")
	}

	// wait forever to prevent the container from restrating
	waitForever()
	// var configHandler = config.NewConfigurationClient(k8sClient)
	// listenToSignals(configHandler)

}

type k8sClientInterface interface {
	CreateValidatingWebhookConfiguration(namespace string, cfg *k8sclient.ValidatingWebhookOpts) (*admissionregistrationV1.ValidatingWebhookConfiguration, error)
	DeleteExistingValidatingWebhook(name string) error
	WaitUntilPodsAreRunning(ctx context.Context, namespace string, selector string, replicas int) error
	GetValidatingWebhookConfiguration(name string) *admissionregistrationV1.ValidatingWebhookConfiguration
	CreatePodWatcher(ctx context.Context, namespace string, selector string) (watch.Interface, error)
	IsPodReady(pod *v1.Pod) bool
	LabelNamespace(namespace string, labels map[string]string) error
	RemoveNamespaceLabels(ns string, labels map[string]string) error
}

func InitWebhook(k8sClient k8sClientInterface) error {
	err := k8sClient.DeleteExistingValidatingWebhook("datree-webhook")
	if err != nil {
		logger.Logf("failed to delete existed validation webhook config, err: %v", err)
		return err
	}

	err = k8sClient.WaitUntilPodsAreRunning(context.Background(), config.GetDatreeValidatingWebhookNamespace(), config.GetDatreeValidatingWebhookPodsSelector(), config.GetDatreeValidatingWebhookServerReplicas())
	if err != nil {
		logger.Logf("failed to wait for pods, err: %v", err)
		return err
	} else {
		logger.Logf("pods are running")
	}

	err = k8sClient.LabelNamespace(config.GetDatreeValidatingWebhookNamespace(), config.GetDatreeValidatingWebhookLabels())
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to label namespace, err: %v", err))
	}

	err = k8sClient.LabelNamespace("kube-system", config.GetDatreeValidatingWebhookLabels())
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to label kube-system namespace, err: %v", err))
	}

	caBundle, _ := getWebhookCABundle()
	logger.Logf("got ca bundle")
	if k8sClient.GetValidatingWebhookConfiguration("datree-webhook") != nil {
		logger.Logf("webhook already exists")
		return nil
	}

	_, err = k8sClient.CreateValidatingWebhookConfiguration(config.GetDatreeValidatingWebhookNamespace(), &k8sclient.ValidatingWebhookOpts{
		MetaName:    config.GetDatreeValidatingWebhookName(),
		ServiceName: config.GetDatreeValidatingWebhookServiceName(),
		CaBundle:    caBundle,
		Selector:    config.GetDatreeValidatingWebhookNamespaceSelector(),
		WebhookName: "webhook-server.datree.svc",
	})
	if err != nil {
		logger.Logf("failed to create validating webhook config, err: %v", err)
		return err
	}

	return nil
}

func waitForever() {
	select {}
}

func getWebhookCABundle() ([]byte, error) {
	certPath := filepath.Join(`/etc/webhook/certs`, `ca-bundle.pem`)
	caPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	return caPEM, nil
}

// func listenToSignals(configHandler *config.ConfigurationClient) {
// 	sigs := make(chan os.Signal, 1)
// 	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

// 	done := make(chan bool, 1)

// 	go func() {
// 		sig := <-sigs
// 		fmt.Println()
// 		fmt.Println(sig)
// 		done <- true
// 	}()

// 	fmt.Println("awaiting signal")
// 	<-done
// 	configHandler.DeleteWebhook()
// 	fmt.Println("exiting")
// }
