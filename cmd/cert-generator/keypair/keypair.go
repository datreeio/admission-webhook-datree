package keypair

import (
	"crypto/rand"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
)

const ErrInvalidTemplate = "invalid template"
const ErrInvalidArgs = "invalid args, templ, caCert and caKey must not be nil"

// generateCA creates the self-signed CA cert and private key
// it will be used to sign the webhook server certificate
func GenerateCA(key *rsa.PrivateKey, templ *x509.Certificate) (*rsa.PrivateKey, *x509.Certificate, error) {
	if templ == nil {
		return nil, nil, fmt.Errorf(ErrInvalidTemplate)
	}

	if key == nil {
		newKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
		if err != nil {
			return nil, nil, err
		}
		key = newKey
	}

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
func GenerateTLS(caCert *x509.Certificate, caKey *rsa.PrivateKey, templ *x509.Certificate) (*rsa.PrivateKey, *x509.Certificate, error) {
	if caCert == nil || caKey == nil || templ == nil {
		return nil, nil, fmt.Errorf(ErrInvalidArgs)
	}

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
