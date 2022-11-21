package renewer

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func certificateToPem(cert []byte) (*bytes.Buffer, error) {
	certPEM := new(bytes.Buffer)
	err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	if err != nil {
		return nil, err
	}

	return certPEM, nil
}

func privateKeyToPem(rsaKey *rsa.PrivateKey) (*bytes.Buffer, error) {
	privateKeyPEM := new(bytes.Buffer)
	err := pem.Encode(privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})
	if err != nil {
		return nil, err
	}

	return privateKeyPEM, nil
}

// generate private and public key for CA with given bitKeySize
func generateRSAPrivKey(bitsSize int) (*rsa.PrivateKey, error) {
	privKey, err := rsa.GenerateKey(cryptorand.Reader, bitsSize)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}
