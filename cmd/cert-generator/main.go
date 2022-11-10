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
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
)

type certificate struct {
	cert    *x509.Certificate
	privKey *rsa.PrivateKey
	certPEM *bytes.Buffer
}

func main() {
	loggerUtil.Log("Start certing!")

	namespace, _ := os.LookupEnv("WEBHOOK_NAMESPACE")

	// set up CA certificate
	caCert, _ := startupCACertificate()

	// set up our server certificate
	serverCert, _ := startupServerCertificate(caCert, namespace)

	err := os.MkdirAll("/etc/webhook/certs/", 0666)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	// save ca-bundle file
	err = writeFile("/etc/webhook/certs/ca-bundle.pem", caCert.certPEM)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	err = writeFile("/etc/webhook/certs/tls.crt", serverCert.certPEM)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	serverPrivKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverCert.privKey),
	})
	err = writeFile("/etc/webhook/certs/tls.key", serverPrivKeyPEM)
	if err != nil {
		loggerUtil.Log(err.Error())
	}

	loggerUtil.Log("Successfully generated self signed CA and signed webhook server certificate")

}

func newCertificate() *x509.Certificate {
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

func startupCACertificate() (*certificate, error) {
	// set up CA certificate
	caX509Certificate := newCertificate()

	// create private and public key for CA with 2048 bitKeySize
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, caX509Certificate, caX509Certificate, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	// pem encode
	// this is a bundle of the CA certificates used to verify that the server is really the correct site you're talking to
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, err
	}

	return &certificate{
		cert:    caX509Certificate,
		privKey: caPrivKey,
		certPEM: caPEM,
	}, nil
}

func startupServerCertificate(caCert *certificate, namespace string) (*certificate, error) {
	// server cert config
	cert := &x509.Certificate{
		DNSNames: []string{
			fmt.Sprintf("datree-webhook-server.%s.svc", namespace),
		},
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: "/CN=datree-webhook-server.datree.svc",
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
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, caCert.cert, &serverPrivKey.PublicKey, caCert.privKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode the server cert and key
	serverCertPEM := new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	return &certificate{
		cert:    cert,
		privKey: serverPrivKey,
		certPEM: serverCertPEM,
	}, nil
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
