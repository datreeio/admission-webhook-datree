package cert_manager

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/big"
	"os"
	"time"
)

const certsFolder = "/app/folder/certs"
const CertPath = certsFolder + "/tls.crt"
const KeyPath = certsFolder + "/tls.key"
const CaCertPath = certsFolder + "/ca.crt"
const CaKeyPath = certsFolder + "/ca.key"

type AllCertificates struct {
	Cert   []byte
	Key    []byte
	CaCert []byte
	CaKey  []byte
}

func GenerateCertificatesIfTheyAreMissing() error {
	if !doCertificatesExist() {
		generateCertificates()
	}
	return nil
}

func doCertificatesExist() bool {
	doesFileExist := func(filePath string) bool {
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			return false
		}
		return true
	}

	return doesFileExist(CertPath) && doesFileExist(KeyPath) && doesFileExist(CaCertPath) && doesFileExist(CaKeyPath)
}

func generateCertificates() {
	var caPEM, serverCertPEM, serverPrivKeyPEM *bytes.Buffer
	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{"velotio.com"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode CA cert
	caPEM = new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	dnsNames := []string{"webhook-service",
		"webhook-service.default", "webhook-service.default.svc"}
	commonName := "datree-webhook-server.datree.svc" // TODO use namespace from config

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"velotio.com"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
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
	serverCertPEM = new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	err = os.MkdirAll("/app/folder/certs/", 0666)
	if err != nil {
		log.Panic(err)
	}
	err = writeFile("/app/folder/certs/tls.crt", serverCertPEM)
	if err != nil {
		log.Panic(err)
	}

	err = writeFile("/app/folder/certs/tls.key", serverPrivKeyPEM)
	if err != nil {
		log.Panic(err)
	}

}

// writeFile writes data in the file at the given path
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
