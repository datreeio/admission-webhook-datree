package main

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
	"k8s.io/client-go/kubernetes"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	caCert, _ := generateSelfSignedCAAndSignWebhookServerCertificate()
	createValidationWebhookConfig(caCert)
}

func createValidationWebhookConfig(caCert *bytes.Buffer) {
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic("failed to set go -client")
	}

	webhookNamespace, _ := os.LookupEnv("WEBHOOK_NAMESPACE")
	webhookService, _ := os.LookupEnv("WEBHOOK_SERVICE")
	validationCfgName := "datree-webhook"

	path := "/validate"
	sideEffects := admissionregistrationv1.SideEffectClassNone

	validationWebhookConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: validationCfgName,
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name: "webhook-server.datree.svc",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caCert.Bytes(), // CA bundle created earlier
				Service: &admissionregistrationv1.ServiceReference{
					Name:      webhookService,
					Namespace: webhookNamespace,
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
		panic(err)
	}
}

func generateSelfSignedCAAndSignWebhookServerCertificate() (*bytes.Buffer, error) {
	ca, caPrivKey, err := createCA("/CN=Admission Controller Webhook Demo CA", time.Now().AddDate(5, 0, 0))
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	// PEM encode CA cert, this is a bundle of the CA certificates used to verify that the server is really the correct site you're talking to
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}
	caPEM := new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	dnsNames := []string{"datree-webhook-server",
		"datree-webhook-server.default",
		"datree-webhook-server.default.svc",
		"datree-webhook-server.datree",
		"datree-webhook-server.datree.svc",
	}

	commonName := "/CN=datree-webhook-server.datree.svc"

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: commonName,
			// Organization: []string{"datree.io"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(5, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode the  server cert and key
	serverCertPEM := new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	err = os.MkdirAll("/etc/webhook/certs/", 0666)
	if err != nil {
		loggerUtil.Log(err.Error())
	}
	err = WriteFile("/etc/webhook/certs/tls.crt", serverCertPEM)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	err = WriteFile("/etc/webhook/certs/tls.key", serverPrivKeyPEM)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	return serverCertPEM, nil

}

// WriteFile writes data in the file at the given path
func WriteFile(filepath string, sCert *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(sCert.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// createCA creates a CA certificate and private key
func createCA(organization string, validity time.Time) (*x509.Certificate, *rsa.PrivateKey, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore:   time.Now(),
		NotAfter:    validity, // 5 years validity
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		// keyCertSign bit is for verifying signatures on public key certificates. (e.g. CA certificate)
		// KeyUsageDigitalSignature bit is for verifying signatures on digital signatures. Read more: https://ldapwiki.com/wiki/KeyCertSign
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create RSA CA private key with 2048 bitKeySize
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	return ca, caPrivKey, nil
}
