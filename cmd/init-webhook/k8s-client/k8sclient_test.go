package k8sclient

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

type Condition[T any] func(actual T) bool

type ExpectedCondition[T any] struct {
	condition Condition[T]
	message   string
}

type Expected[T any] struct {
	value   T
	message string
}

func TestCreateValidatingWebhookConfiguration(t *testing.T) {
	type testArgs struct {
		namespace string
		cfg       *ValidatingWebhookOpts
	}

	type testCase struct {
		args            *testArgs
		expectedErr     Expected[error]
		expectedWebhook ExpectedCondition[*admissionregistrationV1.ValidatingWebhookConfiguration]
	}

	tests := map[string]*testCase{
		"should create a validating webhook configuration": {
			args: &testArgs{namespace: "test-namespace", cfg: &ValidatingWebhookOpts{
				MetaName:    "datree-webhook",
				ServiceName: "webhook-server",
				CaBundle:    []byte("caBundle"),
				Selector:    "app=webhook-server",
				WebhookName: "webhook-server.datree.svc",
			}},
			expectedWebhook: ExpectedCondition[*admissionregistrationV1.ValidatingWebhookConfiguration]{
				condition: func(actual *admissionregistrationV1.ValidatingWebhookConfiguration) bool {
					isNameValid := actual.Name == "datree-webhook"
					isNamespaceValid := actual.Webhooks[0].ClientConfig.Service.Namespace == "test-namespace"
					isWebhookNameValid := actual.Webhooks[0].Name == "webhook-server.datree.svc"
					isWebhookServiceNameValid := actual.Webhooks[0].ClientConfig.Service.Name == "webhook-server"
					isWebhookCABundleValid := string(actual.Webhooks[0].ClientConfig.CABundle) == string([]byte("caBundle"))
					isWebhookSelectorValid := actual.Webhooks[0].NamespaceSelector.MatchExpressions[0].Key == "app=webhook-server"
					return isNameValid && isNamespaceValid && isWebhookNameValid && isWebhookServiceNameValid && isWebhookCABundleValid && isWebhookSelectorValid
				},
				message: "webhook should use the correct values from the args",
			},
			expectedErr: Expected[error]{
				value:   nil,
				message: "should not return an error",
			},
		},
		"should not create validating webhook since opts is nil": {
			args: &testArgs{namespace: "test-namespace", cfg: nil},
			expectedWebhook: ExpectedCondition[*admissionregistrationV1.ValidatingWebhookConfiguration]{
				message: "webhook should be nil",
				condition: func(actual *admissionregistrationV1.ValidatingWebhookConfiguration) bool {
					return actual == nil
				},
			},
			expectedErr: Expected[error]{
				value:   fmt.Errorf("invalid validating webhook configuration"),
				message: "should return an error",
			},
		},
	}

	client := testclient.NewSimpleClientset()
	k8sClient := New(client)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualVW, actualErr := k8sClient.CreateValidatingWebhookConfiguration(tc.args.namespace, tc.args.cfg)
			assert.Equal(t, tc.expectedErr.value, actualErr, tc.expectedErr.message)
			assert.True(t, tc.expectedWebhook.condition(actualVW), tc.expectedWebhook.message)
		})
	}
}
