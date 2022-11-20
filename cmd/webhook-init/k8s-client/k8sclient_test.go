package k8sclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

type condition[T any] struct {
	compareFn func(actual T) bool
	msg       string
}

type output struct {
	err     *condition[error]
	webhook *condition[*admissionregistrationV1.ValidatingWebhookConfiguration]
}

type input struct {
	namespace string
	cfg       *ValidatingWebhookOpts
}

type test struct {
	name     string
	args     input
	expected output
}

func TestCreateValidatingWebhookConfiguration(t *testing.T) {
	tests := []test{
		{
			name: "should create validating webhook configuration",
			args: input{
				namespace: "datree",
				cfg: &ValidatingWebhookOpts{
					MetaName:    "datree-webhook",
					ServiceName: "webhook-server",
					CaBundle:    []byte("caBundle"),
					Selector:    "app=webhook-server",
					WebhookName: "webhook-server.datree.svc",
				}},
			expected: output{
				err: &condition[error]{compareFn: func(actual error) bool { return actual == nil }, msg: "should not return error"},
				webhook: &condition[*admissionregistrationV1.ValidatingWebhookConfiguration]{compareFn: func(actual *admissionregistrationV1.ValidatingWebhookConfiguration) bool {
					return actual != nil && actual.Name == "datree-webhook" && actual.Webhooks[0].Name == "webhook-server.datree.svc" && actual.Webhooks[0].ClientConfig.Service.Name == "webhook-server" && actual.Webhooks[0].ClientConfig.Service.Namespace == "datree" && actual.Webhooks[0].ClientConfig.Service.Path != nil && *actual.Webhooks[0].ClientConfig.Service.Path == "/validate" && actual.Webhooks[0].ClientConfig.CABundle != nil && string(actual.Webhooks[0].ClientConfig.CABundle) == "caBundle"
				}, msg: "should return webhook configuration"},
			},
		},
		{
			name: "should not return error when namespace is empty",
			args: input{
				namespace: "",
				cfg: &ValidatingWebhookOpts{
					MetaName:    "datree-webhook",
					ServiceName: "webhook-server",
					CaBundle:    []byte("caBundle"),
					Selector:    "app=webhook-server",
					WebhookName: "webhook-server.datree.svc",
				},
			},
			expected: output{err: &condition[error]{compareFn: func(actual error) bool { return actual == nil }, msg: "should return error"}},
		},
		{
			name: "should return error when cfg is nil",
			args: input{
				namespace: "datree",
				cfg:       nil,
			},
			expected: output{err: &condition[error]{compareFn: func(actual error) bool { return actual != nil }, msg: "should return error"}},
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			client := testclient.NewSimpleClientset()

			k8sClient := New(client)

			actualVW, actualErr := k8sClient.CreateValidatingWebhookConfiguration(ts.args.namespace, ts.args.cfg)
			if ts.expected.err != nil {
				assert.Condition(t, func() bool { return ts.expected.err.compareFn(actualErr) }, ts.expected.err.msg)
			}
			if ts.expected.webhook != nil {
				assert.Condition(t, func() bool { return ts.expected.webhook.compareFn(actualVW) }, ts.expected.webhook.msg)
			}
		})
	}
}
