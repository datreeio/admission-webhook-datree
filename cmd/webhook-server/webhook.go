package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/datreeio/admission-webhook-datree/pkg/config"
	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type k8sClientInterface interface {
	CreateValidatingWebhookConfiguration(webhook *admissionregistrationV1.ValidatingWebhookConfiguration) error
	DeleteExistingValidatingWebhook(name string) error
	WaitUntilPodsAreRunning(ctx context.Context, namespace string, selector string, replicas int) error
}

const (
	datreeValidatingAdmissionWebhookName = "webhook-server.datree.svc"
)

func InitDatreeValidatingWebhook(k8sClient k8sClientInterface) error {
	caBundle, err := getWebhookCABundle()
	if err != nil {
		return err
	}

	err = k8sClient.DeleteExistingValidatingWebhook(config.GetDatreeValidatingWebhookName())
	if err != nil {
		return err
	}

	// Wait until webhook-server is ready
	err = k8sClient.WaitUntilPodsAreRunning(context.Background(), config.GetDatreeValidatingWebhookNamespace(), config.GetDatreeValidatingWebhookPodsSelector(), config.GetDatreeValidatingWebhookServerReplicas())
	if err != nil {
		return err
	}

	// Create validating webhook configuration
	err = k8sClient.CreateValidatingWebhookConfiguration(buildDatreeValidatingWebhookConfiguration(caBundle))
	if err != nil {
		return err
	}

	return nil

}

func buildDatreeValidatingWebhookConfiguration(caBundle []byte) *admissionregistrationV1.ValidatingWebhookConfiguration {
	path := "/validate"
	sideEffects := admissionregistrationV1.SideEffectClassNone

	return &admissionregistrationV1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetDatreeValidatingWebhookName(),
		},
		Webhooks: []admissionregistrationV1.ValidatingWebhook{
			{
				Name: datreeValidatingAdmissionWebhookName,
				ClientConfig: admissionregistrationV1.WebhookClientConfig{
					CABundle: caBundle, // CA bundle created earlier
					Service: &admissionregistrationV1.ServiceReference{
						Name:      config.GetDatreeValidatingWebhookServiceName(),
						Namespace: config.GetDatreeValidatingWebhookNamespace(),
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
							Key:      config.GetDatreeValidatingWebhookNamespaceSelector(),
							Operator: metav1.LabelSelectorOpDoesNotExist,
						},
					},
				},
			},
		},
	}
}

func getWebhookCABundle() ([]byte, error) {
	certPath := filepath.Join(`/etc/webhook/certs`, `ca-bundle.pem`)
	caPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	return caPEM, nil
}
