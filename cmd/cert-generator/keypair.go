package main

import (
	"crypto/rand"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"time"
)

type certValidityDuration struct {
	begin time.Time
	end   time.Time
}

// generateCA creates the self-signed CA cert and private key
// it will be used to sign the webhook server certificate
func GenerateDatreeCA(key *rsa.PrivateKey, validity certValidityDuration) (*rsa.PrivateKey, *x509.Certificate, error) {
	if key == nil {
		newKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
		if err != nil {
			return nil, nil, err
		}
		key = newKey
	}

	templ := buildDatreeWebhookCATemplate(validity)
	der, err := x509.CreateCertificate(cryptorand.Reader, templ, templ, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, err
	}

	return key, cert, nil
}

// generateTLS takes the results of GenerateCACert and uses it to create the
// PEM-encoded public certificate and private key, respectively
func GenerateTLS(caCert *x509.Certificate, caKey *rsa.PrivateKey, validity certValidityDuration) (*rsa.PrivateKey, *x509.Certificate, error) {
	templ := buildDatreeWebhookServerCertificateTemplate(validity)
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	der, err := x509.CreateCertificate(rand.Reader, templ, caCert, key.Public(), caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create certificate failed, err: %v", err)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, fmt.Errorf("parse certificate failed, err: %v", err)
	}
	return key, cert, nil
}

func buildDatreeWebhookCATemplate(validity certValidityDuration) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		Subject: pkix.Name{
			Organization: []string{"/CN=Datree Admission Controller Webhook CA"},
		},
		NotBefore:   validity.begin,
		NotAfter:    validity.end,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		// keyCertSign bit is for verifying signatures on public key certificates. (e.g. CA certificate)
		// KeyUsageDigitalSignature bit is for verifying signatures on digital signatures. Read more: https://ldapwiki.com/wiki/KeyCertSign
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}

func buildDatreeWebhookServerCertificateTemplate(validity certValidityDuration) *x509.Certificate {
	// webhookServerNamespace, _ := os.LookupEnv("WEBHOOK_NAMESPACE")
	// webhookServerDNS := fmt.Sprintf("datree-webhook-server.%s.svc", webhookServerNamespace)
	webhookDNS, _ := os.LookupEnv("WEBHOOK_SERVER_DNS")

	// server cert config
	return &x509.Certificate{
		DNSNames: []string{
			webhookDNS,
		},
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("/CN=%v", webhookDNS),
		},
		NotBefore:    validity.begin,
		NotAfter:     validity.end,
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
}
