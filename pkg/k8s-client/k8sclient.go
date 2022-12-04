package k8sclient

import (
	"context"
	"fmt"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"

	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
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

// create validating webhook configuration according to options
func (k *K8sClient) CreateValidatingWebhookConfiguration(webhook *admissionregistrationV1.ValidatingWebhookConfiguration) error {
	_, err := k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.Background(), webhook, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create validating webhook configuration: %v", err)
	}

	return nil
}

// search for validating webhook and delete if exists
func (k *K8sClient) DeleteExistingValidatingWebhook(name string) error {
	vw, err := k.GetValidatingWebhookConfiguration(name)
	if vw != nil && err == nil {
		return k.DeleteValidatingWebhookConfiguration(name)
	}

	return nil
}

// delete validating webhook configuration according to name
func (k *K8sClient) DeleteValidatingWebhookConfiguration(name string) error {
	return k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(context.Background(), name, metav1.DeleteOptions{})
}

// get validating webhook configuration according to name
func (k *K8sClient) GetValidatingWebhookConfiguration(name string) (*admissionregistrationV1.ValidatingWebhookConfiguration, error) {
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	vw, err := k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get validating webhook configuration: %v", err)
	}
	if vw == nil {
		return nil, fmt.Errorf("validating webhook configuration not found")
	}

	return vw, nil
}

// watch for pods given namespace to be ready, the selector is used to filter pods
func (k *K8sClient) WaitUntilPodsAreRunning(ctx context.Context, namespace string, podsSelector string, replicas int) error {
	// ctx := context.Background()

	watcher, err := k.createPodWatcher(ctx, namespace, podsSelector)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	runningReplicasCounter := 0
	for {
		select {
		case event := <-watcher.ResultChan():
			pod := event.Object.(*v1.Pod)

			if pod.Status.Phase == v1.PodRunning {
				if k.isPodReady(pod) {
					runningReplicasCounter++
					logger.Logf("the POD '%s' is running. UID: %v", podsSelector, pod.UID)
					if runningReplicasCounter == replicas {
						return nil
					}
				}

			}

		case <-time.After(180 * time.Second):
			logger.Logf("exit from waitPodRunning for POD '%s' because the time is over", podsSelector)
			return nil

		case <-ctx.Done():
			logger.Logf("exit from waitPodRunning for POD '%s' because the context is done", podsSelector)
			return nil
		}
	}
}

func (k *K8sClient) createPodWatcher(ctx context.Context, namespace string, selector string) (watch.Interface, error) {
	opts := metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: selector,
		FieldSelector: "",
	}

	return k.clientset.CoreV1().Pods(namespace).Watch(ctx, opts)
}

func (k *K8sClient) isPodReady(pod *v1.Pod) bool {
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
