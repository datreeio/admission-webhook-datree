package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	certPath := filepath.Join(`/etc/webhook/certs`, `ca-bundle.pem`)
	caPEM, _ := os.ReadFile(certPath)

	time.Sleep(180 * time.Second)
	loggerUtil.Log("Sleep Over.....")

	err := createValidationWebhookConfig(caPEM)

	if err != nil {
		loggerUtil.Log(fmt.Sprintf("failed to create validation webhook config, err: %v", err))
	} else {
		loggerUtil.Log("created validating webhook configuration")
	}
}

func createValidationWebhookConfig(caBundle []byte) error {
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err // panic("failed to set go -client")
	}

	path := "/validate"
	sideEffects := admissionregistrationv1.SideEffectClassNone

	validationWebhookConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "datree-webhook",
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name: "webhook-server.datree.svc",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caBundle, // CA bundle created earlier
				Service: &admissionregistrationv1.ServiceReference{
					Name:      "datree-webhook-server", // datree-webhook-server
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
						Key:      "admission.datree/validate",
						Operator: metav1.LabelSelectorOpDoesNotExist,
					},
				},
			},
		}},
	}

	if _, err = kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.Background(), validationWebhookConfig, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}
