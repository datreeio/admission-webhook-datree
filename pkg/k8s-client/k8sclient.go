package k8sclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

func (k *K8sClient) LabelNamespace(ns string, labels map[string]string) error {
	_, err := k.clientset.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to get namespace %s: %v", ns, err))
	}

	jsonLabels, _ := json.Marshal(labels)
	_, err = k.clientset.CoreV1().Namespaces().Patch(context.Background(), ns, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":%s}}`, jsonLabels)), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to add label to namespace %s: %v", ns, err))
	}

	logger.Logf("added label to namespace %s", ns)
	return nil
}

func (k *K8sClient) RemoveNamespaceLabels(ns string, labels map[string]string) error {
	namespace, err := k.clientset.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to get namespace %s: %v", ns, err))
	}

	nsLabels := namespace.GetLabels()
	if nsLabels == nil {
		return fmt.Errorf(fmt.Sprintf("namespace %s has no labels", ns))
	}

	for k := range labels {
		delete(nsLabels, k)
	}

	jsonLabels, _ := json.Marshal(nsLabels)
	_, err = k.clientset.CoreV1().Namespaces().Patch(context.Background(), ns, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":%s}}`, jsonLabels)), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to remove label from namespace %s: %v", ns, err))
	}

	return nil
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
			Rules: []admissionregistrationV1.RuleWithOperations{
				{
					Operations: []admissionregistrationV1.OperationType{
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

	existed := k.GetValidatingWebhookConfiguration(cfg.MetaName)
	if existed == nil {
		vw, err := k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.Background(), vw, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		return vw, nil
	}

	logger.Logf("validating webhook configuration %s already exists", existed.Name)
	return existed, nil
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
	if err != nil && !errors.IsNotFound(err) {
		fmt.Printf("failed to get validating webhook configuration %s: %v", name, err)
	}
	if (vw != nil && vw.Name == name) && err == nil {
		fmt.Printf("found validating webhook configuration %s", vw.Name)
		return vw
	}

	return nil
}

func (k *K8sClient) CreatePodWatcher(ctx context.Context, namespace string, selector string) (watch.Interface, error) {
	opts := metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: selector,
		FieldSelector: "",
	}

	return k.clientset.CoreV1().Pods(namespace).Watch(ctx, opts)
}

func (k *K8sClient) WaitUntilPodsAreRunning(ctx context.Context, namespace string, selector string, replicas int) error {
	watcher, err := k.CreatePodWatcher(ctx, namespace, selector)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	count := 0
	for {
		select {
		case event := <-watcher.ResultChan():
			pod := event.Object.(*v1.Pod)

			if pod.Status.Phase == v1.PodRunning {
				if k.IsPodReady(pod) {
					count++
					logger.Logf("the POD '%s' is running. UID: %v", selector, pod.UID)
					if count == replicas {
						return nil
					}
				}

			}

		case <-time.After(180 * time.Second):
			logger.Logf("exit from waitPodRunning for POD '%s' because the time is over", selector)
			return nil

		case <-ctx.Done():
			logger.Logf("exit from waitPodRunning for POD '%s' because the context is done", selector)
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
