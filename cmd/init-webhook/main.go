package main

import (
	"context"
	"fmt"

	k8sclient "github.com/datreeio/admission-webhook-datree/cmd/init-webhook/k8s-client"
	webhookinfo "github.com/datreeio/admission-webhook-datree/cmd/init-webhook/webhook-info"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func main() {
	k8sClient := k8sclient.New(nil)
	if k8sClient == nil {
		logger.LogUtil("failed to set k8s go -client, err")
	}

	err := InitWebhook(k8sClient)
	if err != nil {
		logger.LogUtil(fmt.Sprintf("failed to init webhook, err: %v", err))
	}
	logger.LogUtil("horray! succesfully created datree validating admission webhook")

	// wait forever to prevent the container from restrating
	waitForever()
}

type k8sClientInterface interface {
	CreateValidatingWebhookConfiguration(namespace string, cfg *k8sclient.ValidatingWebhookOpts) (*admissionregistrationV1.ValidatingWebhookConfiguration, error)
	DeleteExistingValidatingWebhook(name string) error
	WaitUntilPodsAreRunning(ctx context.Context, namespace string, selector string, replicas int) error
	GetValidatingWebhookConfiguration(name string) *admissionregistrationV1.ValidatingWebhookConfiguration
	CreatePodWatcher(ctx context.Context, namespace string, selector string) (watch.Interface, error)
	IsPodReady(pod *v1.Pod) bool
}

func InitWebhook(k8sClient k8sClientInterface) error {
	err := k8sClient.DeleteExistingValidatingWebhook("datree-webhook")
	if err != nil {
		logger.LogUtil(fmt.Sprintf("failed to delete existed validation webhook config, err: %v", err))
		return err
	}

	err = k8sClient.WaitUntilPodsAreRunning(context.Background(), webhookinfo.GetWebhookNamespace(), webhookinfo.GetWebhookSelector(), webhookinfo.GetWebhookServerReplicas())
	if err != nil {
		logger.LogUtil(fmt.Sprintf("failed to wait for pods, err: %v", err))
		return err
	}

	caBundle, _ := webhookinfo.GetWebhookCABundle()
	_, err = k8sClient.CreateValidatingWebhookConfiguration(webhookinfo.GetWebhookNamespace(), &k8sclient.ValidatingWebhookOpts{
		MetaName:    "datree-webhook",
		ServiceName: webhookinfo.GetWebhookServiceName(),
		CaBundle:    caBundle,
		Selector:    webhookinfo.GetWebhookSelector(),
		WebhookName: "webhook-server.datree.svc",
	})
	if err != nil {
		return err
	}

	return nil
}

func waitForever() {
	select {}
}
