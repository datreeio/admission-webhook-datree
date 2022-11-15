package main

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
)

func main() {
	// set up self signed CA certificate
	datreeCAPrivKey, err := generatePrivKey(2048)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	// create the CA
	datreeCA := getDatreeAdmissionWebhookCAConfig()
	datreeCACertBytes, err := x509.CreateCertificate(cryptorand.Reader, datreeCA, datreeCA, &datreeCAPrivKey.PublicKey, datreeCAPrivKey)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	// set up server certificate
	serverPrivKey, err := generatePrivKey(4096)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	// sign the server cert with Datree CA
	serverCert := getServerCertificateConfig()
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, serverCert, datreeCA, &serverPrivKey.PublicKey, datreeCAPrivKey)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	tlsDir, isFound := os.LookupEnv("WEBHOOK_CERTS_DIR")
	if !isFound {
		loggerUtil.Log("Required directory path for webhook certificates is missing, verify env varaible WEBHOOK_CERTS_DIR in deployment.")
	}

	err = os.MkdirAll(tlsDir, 0666)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	err = saveCACertificateCABundle(tlsDir, datreeCACertBytes)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	err = saveServerTLSCertificate(tlsDir, serverCertBytes, serverPrivKey)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	loggerUtil.Log("Successfully generated self-signed CA and signed webhook server certificate using this CA!")
	os.Exit(0)
}

// generate private and public key for CA with given bitKeySize
func generatePrivKey(bitsSize int) (*rsa.PrivateKey, error) {
	privKey, err := rsa.GenerateKey(cryptorand.Reader, bitsSize)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func getDatreeAdmissionWebhookCAConfig() *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		Subject: pkix.Name{
			Organization: []string{"/CN=Datree Admission Controller Webhook CA"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(5, 0, 0), // 5 years validity
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		// keyCertSign bit is for verifying signatures on public key certificates. (e.g. CA certificate)
		// KeyUsageDigitalSignature bit is for verifying signatures on digital signatures. Read more: https://ldapwiki.com/wiki/KeyCertSign
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}

func getServerCertificateConfig() *x509.Certificate {
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
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(5, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
}

func saveServerTLSCertificate(tlsDir string, serverCertificate []byte, serverPrivKey *rsa.PrivateKey) error {
	// PEM encode the server cert and key
	serverCertPEM := new(bytes.Buffer)
	err := pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertificate,
	})
	if err != nil {
		return err
	}

	err = writeFile(filepath.Join(tlsDir, `tls.crt`), serverCertPEM)
	if err != nil {
		return err
	}

	serverPrivKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	err = writeFile(filepath.Join(tlsDir, `tls.key`), serverPrivKeyPEM)
	if err != nil {
		return err
	}

	return nil
}

func saveCACertificateCABundle(tlsDir string, caCertificate []byte) error {
	// pem encode a bundle of the CA certificates
	// this is used to verify that the server is really the correct site you're talking to
	datreeCACertPEM := new(bytes.Buffer)
	err := pem.Encode(datreeCACertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertificate,
	})
	if err != nil {
		return err
	}

	err = writeFile(filepath.Join(tlsDir, "ca-bundle.pem"), datreeCACertPEM)
	if err != nil {
		return err
	}

	return nil
}

// WriteFile writes data in the file at the given path
func writeFile(filepath string, sCert *bytes.Buffer) error {
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
