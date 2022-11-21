package webhookinfo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Condition[T any] func(actual T) bool

type ExpectedCondition[T any] struct {
	condition Condition[T]
	message   string
}

func TestGetWebhookServerReplicas(t *testing.T) {
	type testCase struct {
		envVarValue string
		expected    ExpectedCondition[int]
	}

	tests := map[string]*testCase{
		"should return 2 when replicas is not set": {
			envVarValue: "",
			expected: ExpectedCondition[int]{
				condition: func(actual int) bool {
					return actual == 2
				},
				message: "expected replicas to be 2, got %d",
			},
		},
		"should return 2 when replicas is set to 2": {
			envVarValue: "2",
			expected: ExpectedCondition[int]{
				condition: func(actual int) bool {
					return actual == 2
				},
				message: "expected replicas to be 2, got %d",
			},
		},
		"should return 1 when replicas is set to 1": {
			envVarValue: "1",
			expected: ExpectedCondition[int]{
				condition: func(actual int) bool {
					return actual == 1
				},
				message: "expected replicas to be 1, got %d",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("WEBHOOK_REPLICAS", test.envVarValue)
			replicas := GetWebhookServerReplicas()
			assert.True(t, test.expected.condition(replicas), test.expected.message, replicas)
		})
	}
}

func TestGetWebhookSelector(t *testing.T) {
	type testCase struct {
		envVarValue string
		expected    ExpectedCondition[string]
	}

	tests := map[string]*testCase{
		"should 'admission.datree/validate' when replicas is not set": {
			envVarValue: "",
			expected: ExpectedCondition[string]{
				condition: func(actual string) bool {
					return actual == "admission.datree/validate"
				},
				message: "expected selector to be 'admission.datree/validate', got %d",
			},
		},
		"should return 'datree/validate' when selector is set to 'datree/validate'": {
			envVarValue: "datree/validate",
			expected: ExpectedCondition[string]{
				condition: func(actual string) bool {
					return actual == "datree/validate"
				},
				message: "expected selector to be 'datree/validate', got %d",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("WEBHOOK_SELECTOR", test.envVarValue)
			selector := GetWebhookSelector()
			assert.True(t, test.expected.condition(selector), test.expected.message, selector)
		})
	}
}

func TestGetWebhookNamespace(t *testing.T) {
	type testCase struct {
		envVarValue string
		expected    ExpectedCondition[string]
	}

	tests := map[string]*testCase{
		"should return datree namespace when namespace is not set": {
			envVarValue: "",
			expected: ExpectedCondition[string]{
				condition: func(actual string) bool {
					return actual == "datree"
				},
				message: "expected namespace to be 'datree', got %d",
			},
		},
		"should return default namespace when namespace is set to default": {
			envVarValue: "default",
			expected: ExpectedCondition[string]{
				condition: func(actual string) bool {
					return actual == "default"
				},
				message: "expected namesapce to be 'default', got %d",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("WEBHOOK_NAMESPACE", test.envVarValue)
			ns := GetWebhookNamespace()
			assert.True(t, test.expected.condition(ns), test.expected.message, ns)
		})
	}
}

func TestGetWebhookServiceName(t *testing.T) {
	type testCase struct {
		envVarValue string
		expected    ExpectedCondition[string]
	}

	tests := map[string]*testCase{
		"should return 'datree-webhook-server' when service name is not set": {
			envVarValue: "",
			expected: ExpectedCondition[string]{
				condition: func(actual string) bool {
					return actual == "datree-webhook-server"
				},
				message: "expected selector to be 'datree-webhook-server', got %d",
			},
		},
		"should return 'datree-service' when selector is set to 'datree-service'": {
			envVarValue: "datree-service",
			expected: ExpectedCondition[string]{
				condition: func(actual string) bool {
					return actual == "datree-service"
				},
				message: "expected selector to be 'datree-service', got %d",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("WEBHOOK_SERVICE", test.envVarValue)
			serviceName := GetWebhookServiceName()
			assert.True(t, test.expected.condition(serviceName), test.expected.message, serviceName)
		})
	}

}
