package k8sclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	type testCase struct {
		webhookArg  *admissionregistrationV1.ValidatingWebhookConfiguration
		expectedErr ExpectedCondition[error]
	}

	tests := map[string]*testCase{
		"should call client to create a validating webhook configuration": {
			webhookArg: &admissionregistrationV1.ValidatingWebhookConfiguration{},
			expectedErr: ExpectedCondition[error]{
				condition: func(actual error) bool {
					return actual == nil
				},
				message: "expected no error",
			},
		},
		"should not create validating webhook since opts is nil": {
			webhookArg: nil,
			expectedErr: ExpectedCondition[error]{
				condition: func(actual error) bool {
					return actual != nil
				},
				message: "expected error",
			},
		},
	}

	client := testclient.NewSimpleClientset()
	k8sClient := New(client)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualErr := k8sClient.CreateValidatingWebhookConfiguration(&admissionregistrationV1.ValidatingWebhookConfiguration{})
			assert.True(t, tc.expectedErr.condition(actualErr), tc.expectedErr.message)
		})
	}
}

// test get validating webhook configuration
func TestGetValidatingWebhookConfiguration(t *testing.T) {
	type testArgs struct {
		name            string
		clientsetReturn *admissionregistrationV1.ValidatingWebhookConfiguration
	}

	type testCase struct {
		args        testArgs
		expectedErr ExpectedCondition[error]
		expected    Expected[*admissionregistrationV1.ValidatingWebhookConfiguration]
	}

	tests := map[string]*testCase{
		"should call client to get a validating webhook configuration": {
			args: testArgs{
				name: "test",
				clientsetReturn: &admissionregistrationV1.ValidatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
			},
			expectedErr: ExpectedCondition[error]{
				condition: func(actual error) bool {
					return actual == nil
				},
				message: "expected no error",
			},
			expected: Expected[*admissionregistrationV1.ValidatingWebhookConfiguration]{
				value: &admissionregistrationV1.ValidatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
				message: "expected a validating webhook configuration",
			},
		},
		"should not get validating webhook since name is empty": {
			args: testArgs{
				name:            "",
				clientsetReturn: &admissionregistrationV1.ValidatingWebhookConfiguration{},
			},
			expectedErr: ExpectedCondition[error]{
				condition: func(actual error) bool {
					return actual != nil
				},
				message: "expected error",
			},
			expected: Expected[*admissionregistrationV1.ValidatingWebhookConfiguration]{
				value:   nil,
				message: "expected no validating webhook configuration",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := testclient.NewSimpleClientset(tc.args.clientsetReturn)
			k8sClient := New(client)

			actual, actualErr := k8sClient.GetValidatingWebhookConfiguration(tc.args.name)
			assert.True(t, tc.expectedErr.condition(actualErr), tc.expectedErr.message)
			assert.Equal(t, tc.expected.value, actual, tc.expected.message)
		})
	}
}
