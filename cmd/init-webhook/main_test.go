package main

import (
	"context"
	"testing"

	config "github.com/datreeio/admission-webhook-datree/pkg/config"
	k8sclient "github.com/datreeio/admission-webhook-datree/pkg/k8s-client"
	"github.com/stretchr/testify/mock"
	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	testclient "k8s.io/client-go/kubernetes/fake"
)

type mockK8sClient struct {
	mock.Mock
	clientset *testclient.Clientset
}

func (m *mockK8sClient) CreateValidatingWebhookConfiguration(namespace string, cfg *k8sclient.ValidatingWebhookOpts) (*admissionregistrationV1.ValidatingWebhookConfiguration, error) {
	args := m.Called(namespace, cfg)
	return args.Get(0).(*admissionregistrationV1.ValidatingWebhookConfiguration), args.Error(1)
}

func (m *mockK8sClient) DeleteExistingValidatingWebhook(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *mockK8sClient) WaitUntilPodsAreRunning(ctx context.Context, namespace string, selector string, replicas int) error {
	args := m.Called(ctx, namespace, selector, replicas)
	return args.Error(0)
}

func (m *mockK8sClient) GetValidatingWebhookConfiguration(name string) *admissionregistrationV1.ValidatingWebhookConfiguration {
	args := m.Called(name)
	return args.Get(0).(*admissionregistrationV1.ValidatingWebhookConfiguration)
}

func (m *mockK8sClient) CreatePodWatcher(ctx context.Context, namespace string, selector string) (watch.Interface, error) {
	args := m.Called(ctx, namespace, selector)
	return args.Get(0).(watch.Interface), args.Error(1)
}

func (m *mockK8sClient) IsPodReady(pod *v1.Pod) bool {
	args := m.Called(pod)
	return args.Bool(0)
}

func (m *mockK8sClient) LabelNamespace(namespace string, labels map[string]string) error {
	args := m.Called(namespace, labels)
	return args.Error(0)
}

func (m *mockK8sClient) RemoveNamespaceLabels(ns string, labels map[string]string) error {
	args := m.Called(ns, labels)
	return args.Error(0)
}

func TestInitWebhook(t *testing.T) {
	k8sClient := &mockK8sClient{
		clientset: testclient.NewSimpleClientset(),
	}

	k8sClient.On("DeleteExistingValidatingWebhook", mock.Anything).Return(nil)
	k8sClient.On("WaitUntilPodsAreRunning", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	k8sClient.On("GetWebhookCABundle", mock.Anything).Return([]byte{}, nil)
	k8sClient.On("CreateValidatingWebhookConfiguration", mock.Anything, mock.Anything).Return(&admissionregistrationV1.ValidatingWebhookConfiguration{}, nil)

	t.Run("should delete existing webhook, call wait for pods and then create the validation webhook", func(t *testing.T) {
		_ = InitWebhook(k8sClient)

		k8sClient.AssertCalled(t, "DeleteExistingValidatingWebhook", "datree-webhook")
		k8sClient.AssertCalled(t, "WaitUntilPodsAreRunning", mock.Anything, config.GetDatreeValidatingWebhookNamespace(), config.GetDatreeValidatingWebhookPodsSelector(), config.GetDatreeValidatingWebhookServerReplicas())
		k8sClient.AssertCalled(t, "CreateValidatingWebhookConfiguration", config.GetDatreeValidatingWebhookNamespace(), &k8sclient.ValidatingWebhookOpts{
			MetaName:    config.GetDatreeValidatingWebhookName(),
			ServiceName: config.GetDatreeValidatingWebhookServiceName(),
			CaBundle:    nil,
			Selector:    config.GetDatreeValidatingWebhookNamespaceSelector(),
			WebhookName: "webhook-server.datree.svc",
		})
	})
}
