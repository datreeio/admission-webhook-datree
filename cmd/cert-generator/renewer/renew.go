package renewer

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/datreeio/admission-webhook-datree/cmd/cert-generator/keypair"
)

type Writer interface {
	WriteFile(filepath string, sCert *bytes.Buffer) error
}

type CertRenewer struct {
	w                   Writer
	tlsValidityDuration time.Duration
}

func NewCertRenewer(w Writer) *CertRenewer {
	return &CertRenewer{
		w:                   w,
		tlsValidityDuration: 5 * 365 * 24 * time.Hour, // 5 years
	}
}

func (r *CertRenewer) RenewCA(tlsDir string) (*rsa.PrivateKey, *x509.Certificate, error) {
	templ := buildDatreeWebhookCATemplate(r.tlsValidityDuration)

	caPrivKey, caCert, err := keypair.GenerateCA(nil, templ)
	if err != nil {
		return nil, nil, err
	}
	err = r.writeCABundleFile(tlsDir, caCert.Raw)
	if err != nil {
		return nil, nil, err
	}

	return caPrivKey, caCert, nil

}

func (r *CertRenewer) RenewTLS(tlsDir string, caCert *x509.Certificate, caPrivKey *rsa.PrivateKey) error {
	// sign the server cert with Datree CA

	templ := buildDatreeWebhookServerCertificateTemplate(r.tlsValidityDuration)
	serverPrivKey, serverCert, err := keypair.GenerateTLS(caCert, caPrivKey, templ)
	if err != nil {
		return err
	}

	// PEM encode the server cert and key
	serverCertPEM, err := certificateToPem(serverCert.Raw)
	if err != nil {
		return err
	}

	err = r.w.WriteFile(filepath.Join(tlsDir, `tls.crt`), serverCertPEM)
	if err != nil {
		return err
	}

	serverPrivKeyPEM, err := privateKeyToPem(serverPrivKey)
	if err != nil {
		return err
	}

	err = r.w.WriteFile(filepath.Join(tlsDir, `tls.key`), serverPrivKeyPEM)
	if err != nil {
		return err
	}

	return nil
}

func (r *CertRenewer) writeCABundleFile(tlsDir string, caCertificate []byte) error {
	// pem encode a bundle of the CA certificates
	// this is used to verify that the server is really the correct site you're talking to
	datreeCACertPEM, err := certificateToPem(caCertificate)
	if err != nil {
		return err
	}

	err = r.w.WriteFile(filepath.Join(tlsDir, "ca-bundle.pem"), datreeCACertPEM)
	if err != nil {
		return err
	}

	return nil
}

func buildDatreeWebhookCATemplate(certValidityDuration time.Duration) *x509.Certificate {
	now := time.Now()
	begin, end := now, now.Add(certValidityDuration)

	return &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		Subject: pkix.Name{
			Organization: []string{"/CN=Datree Admission Controller Webhook CA"},
		},
		NotBefore:   begin,
		NotAfter:    end,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		// keyCertSign bit is for verifying signatures on public key certificates. (e.g. CA certificate)
		// KeyUsageDigitalSignature bit is for verifying signatures on digital signatures. Read more: https://ldapwiki.com/wiki/KeyCertSign
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}

func buildDatreeWebhookServerCertificateTemplate(certValidityDuration time.Duration) *x509.Certificate {
	webhookDNS := getWebhookServerDNSName()

	now := time.Now()
	begin, end := now, now.Add(certValidityDuration)

	return &x509.Certificate{
		DNSNames: []string{
			webhookDNS,
		},
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("/CN=%v", webhookDNS),
		},
		NotBefore:    begin,
		NotAfter:     end,
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
}

func getWebhookServerDNSName() string {
	webhookDNS, isFound := os.LookupEnv("WEBHOOK_SERVER_DNS")
	if !isFound {
		return "datree-webhook-server.datree.svc"
	}

	return webhookDNS
}
