package k8sclient

import (
	"context"
	"fmt"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type K8sClient struct {
	clientset kubernetes.Interface
}

func New(c kubernetes.Interface) *K8sClient {
	if c != nil {
		return &K8sClient{
			clientset: c,
		}
	}

	c, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		return nil
	}

	return &K8sClient{
		clientset: c,
	}

}

type ValidatingWebhookOpts struct {
	MetaName    string
	CaBundle    []byte
	ServiceName string
	Selector    string
	WebhookName string
}

func (k *K8sClient) CreateValidatingWebhookConfiguration(namespace string, cfg *ValidatingWebhookOpts) (*admissionregistrationV1.ValidatingWebhookConfiguration, error) {
	if cfg == nil {
		return nil, fmt.Errorf("invalid validating webhook configuration")
	}

	path := "/validate"
	sideEffects := admissionregistrationV1.SideEffectClassNone

	vw := &admissionregistrationV1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.MetaName,
		},
		Webhooks: []admissionregistrationV1.ValidatingWebhook{{
			Name: cfg.WebhookName,
			ClientConfig: admissionregistrationV1.WebhookClientConfig{
				CABundle: cfg.CaBundle, // CA bundle created earlier
				Service: &admissionregistrationV1.ServiceReference{
					Name:      cfg.ServiceName, // datree-webhook-server
					Namespace: namespace,
					Path:      &path,
				},
			},
			Rules: []admissionregistrationV1.RuleWithOperations{{Operations: []admissionregistrationV1.OperationType{
				admissionregistrationV1.Create,
				admissionregistrationV1.Update,
			},
				Rule: admissionregistrationV1.Rule{
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
						Key:      cfg.Selector,
						Operator: metav1.LabelSelectorOpDoesNotExist,
					},
				},
			},
		}},
	}

	vw, err := k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.Background(), vw, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// should be debug
	logger.LogUtil(fmt.Sprintf("created validating webhook configuration: %v", vw))
	return vw, nil
}

// search for validating webhook and delete if exists
func (k *K8sClient) DeleteExistingValidatingWebhook(name string) error {
	vw := k.GetValidatingWebhookConfiguration(name)
	if vw != nil {
		return k.DeleteValidatingWebhookConfiguration(name)
	}
	return nil
}

func (k *K8sClient) DeleteValidatingWebhookConfiguration(name string) error {
	return k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (k *K8sClient) GetValidatingWebhookConfiguration(name string) *admissionregistrationV1.ValidatingWebhookConfiguration {
	vw, err := k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), name, metav1.GetOptions{})
	if (vw != nil && vw.Name == name) && err == nil {
		return vw
	}

	return nil
}

func (k *K8sClient) CreatePodWatcher(ctx context.Context, namespace string, selector string) (watch.Interface, error) {
	labelSelector := fmt.Sprint(selector)

	opts := metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: labelSelector,
		FieldSelector: "",
	}

	return k.clientset.CoreV1().Pods(namespace).Watch(ctx, opts)
}

func (k *K8sClient) WaitUntilPodsAreRunning(ctx context.Context, namespace string, selector string, replicas int) error {
	logger.LogUtil(fmt.Sprintf("creating watcher for POD with label:%s ...", selector))
	watcher, err := k.CreatePodWatcher(ctx, namespace, selector)
	if err != nil {
		return err
	}

	logger.LogUtil("watch out! Succuessfuly created watcher for PODs.")
	defer watcher.Stop()

	count := 0
	for {
		select {
		case event := <-watcher.ResultChan():
			pod := event.Object.(*v1.Pod)

			if pod.Status.Phase == v1.PodRunning {
				if k.IsPodReady(pod) {
					count++
					logger.LogUtil(fmt.Sprintf("the POD \"%s\" is running", selector))
					if count == replicas {
						logger.LogUtil(fmt.Sprintf("all PODs are running"))
						return nil
					}
				}

			}

		case <-time.After(180 * time.Second):
			logger.LogUtil(fmt.Sprintf("exit from waitPodRunning for POD \"%s\" because the time is over", selector))
			return nil

		case <-ctx.Done():
			logger.LogUtil(fmt.Sprintf("exit from waitPodRunning for POD \"%s\" because the context is done", selector))
			return nil
		}
	}
}

func (k *K8sClient) IsPodReady(pod *v1.Pod) bool {
	checkPodReadyCondition := func(condition v1.PodCondition) bool {
		return condition.Type == v1.PodReady && condition.Status == "True"
	}

	checkPodContainersCondition := func(condition v1.PodCondition) bool {
		return condition.Type == v1.ContainersReady && condition.Status == "True"
	}

	return findIndex(checkPodReadyCondition, pod.Status.Conditions) > 0 && findIndex(checkPodContainersCondition, pod.Status.Conditions) > 0
}

// util function, should be in a separate file but for the sake of simplicity it's here
func findIndex[T interface{}, K []T](findFn func(element T) bool, array K) (idx int) {
	for i, v := range array {
		if findFn(v) {
			return i
		}
	}
	return -1
}
