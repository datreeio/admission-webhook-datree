package main

import (
	"context"

	k8sclient "github.com/datreeio/admission-webhook-datree/cmd/webhook-init/k8s-client"
	"github.com/datreeio/admission-webhook-datree/cmd/webhook-init/utils"
	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func main() {
	k8sClient := k8sclient.New(nil)
	if k8sClient == nil {
		loggerUtil.Logf("failed to set k8s go -client, err")
	}

	err := InitWebhook(k8sClient)
	if err != nil {
		loggerUtil.Logf("failed to init webhook, err: %v", err)
	}
	loggerUtil.Log("horray! succesfully created datree validating admission webhook")

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
		loggerUtil.Logf("failed to delete existed validation webhook config, err: %v", err)
		return err
	}

	err = k8sClient.WaitUntilPodsAreRunning(context.Background(), utils.GetWebhookNamespace(), utils.GetWebhookSelector(), utils.GetWebhookServerReplicas())
	if err != nil {
		loggerUtil.Logf("failed to wait for pods, err: %v", err)
		return err
	}

	caBundle, _ := utils.GetWebhookCABundle()
	_, err = k8sClient.CreateValidatingWebhookConfiguration(utils.GetWebhookNamespace(), &k8sclient.ValidatingWebhookOpts{
		MetaName:    "datree-webhook",
		ServiceName: utils.GetWebhookServiceName(),
		CaBundle:    caBundle,
		Selector:    utils.GetWebhookSelector(),
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
