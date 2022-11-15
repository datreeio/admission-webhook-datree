package main

import (
	"context"
	"fmt"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	_admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type KubernetesClient interface {
	CoreV1() corev1.CoreV1Interface
	AdmissionregistrationV1() _admissionregistrationv1.AdmissionregistrationV1Interface
}

type k8sClient struct {
	clientset KubernetesClient
	config    struct {
		namespace string
	}
}

func newK8sClient(c KubernetesClient, ns string) *k8sClient {
	return &k8sClient{
		clientset: c,
		config: struct{ namespace string }{
			ns,
		},
	}
}

func (k *k8sClient) createPodWatcher(ctx context.Context, selector string) (watch.Interface, error) {
	labelSelector := fmt.Sprint(selector)

	opts := metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: labelSelector,
		FieldSelector: "",
	}

	return k.clientset.CoreV1().Pods(k.config.namespace).Watch(ctx, opts)
}

func (k *k8sClient) waitPodsRunning(ctx context.Context, selector string, replicas int) error {
	loggerUtil.Debugf("creating watcher for POD with label:%s ...", selector)
	watcher, err := k.createPodWatcher(ctx, selector)
	if err != nil {
		return err
	}

	loggerUtil.Debug("watch out! Succuessfuly created watcher for PODs.")
	defer watcher.Stop()

	count := 0
	for {
		select {
		case event := <-watcher.ResultChan():
			pod := event.Object.(*v1.Pod)

			if pod.Status.Phase == v1.PodRunning {
				if isPodReady(pod) {
					count++
					loggerUtil.Debugf("the POD \"%s\" is running", selector)
					if count == replicas {
						loggerUtil.Debugf("all PODs are running", selector)
						return nil
					}
				}

			}

		case <-time.After(180 * time.Second):
			loggerUtil.Debug("exit from waitPodRunning for POD \"%s\" because the time is over")
			return nil

		case <-ctx.Done():
			loggerUtil.Debugf("exit from waitPodRunning for POD \"%s\" because the context is done", selector)
			return nil
		}
	}
}

func isPodReady(pod *v1.Pod) bool {
	checkPodReadyCondition := func(condition v1.PodCondition) bool {
		return condition.Type == v1.PodReady && condition.Status == "True"
	}

	checkPodContainersCondition := func(condition v1.PodCondition) bool {
		return condition.Type == v1.ContainersReady && condition.Status == "True"
	}

	return findIndex(checkPodReadyCondition, pod.Status.Conditions) > 0 && findIndex(checkPodContainersCondition, pod.Status.Conditions) > 0
}

type validationWebhookConfig struct {
	name        string
	caBundle    []byte
	serviceName string
	namesapce   string
	selector    string
}

func (k *k8sClient) createValidationWebhookConfig(cfg *validationWebhookConfig) error {
	path := "/validate"
	sideEffects := admissionregistrationv1.SideEffectClassNone

	validationWebhookConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "datree-webhook",
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name: "webhook-server.datree.svc",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: cfg.caBundle, // CA bundle created earlier
				Service: &admissionregistrationv1.ServiceReference{
					Name:      cfg.serviceName, // datree-webhook-server
					Namespace: "datree",
					Path:      &path,
				},
			},
			Rules: []admissionregistrationv1.RuleWithOperations{{Operations: []admissionregistrationv1.OperationType{
				admissionregistrationv1.Create,
				admissionregistrationv1.Update,
			},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{"*"},
					APIVersions: []string{"*"},
					Resources:   []string{"*"},
				},
			}},
			SideEffects:             &sideEffects,
			AdmissionReviewVersions: []string{"v1", "v1beta1"},
			TimeoutSeconds:          &[]int32{30}[0],
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{ // only validate pods in namespaces with the label "admission.datree/validate"
						Key:      cfg.selector,
						Operator: metav1.LabelSelectorOpDoesNotExist,
					},
				},
			},
		}},
	}

	_, err := k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.Background(), validationWebhookConfig, metav1.CreateOptions{})
	return err
}

func (k *k8sClient) deleteExistingValidationAdmissionWebhook(name string) error {
	vw := k.getValidationAdmissionWebhook("datree-webhook")
	if vw != nil {
		return k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(context.Background(), name, metav1.DeleteOptions{})
	}
	return nil
}

func (k *k8sClient) getValidationAdmissionWebhook(name string) *admissionregistrationv1.ValidatingWebhookConfiguration {
	vw, err := k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), name, metav1.GetOptions{})
	if vw.Name == name && err == nil {
		return vw
	}

	return nil
}
