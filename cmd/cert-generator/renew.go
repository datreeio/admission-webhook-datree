package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"path/filepath"
)

type writer interface {
	writeFile(filepath string, sCert *bytes.Buffer) error
}

type Renewer struct {
	w writer
}

func newRenewer(w writer) *Renewer {
	return &Renewer{
		w: w,
	}
}

func (r *Renewer) renewCA(tlsDir string, validity certValidityDuration) (*rsa.PrivateKey, *x509.Certificate, error) {
	caPrivKey, caCert, err := GenerateDatreeCA(nil, validity)
	if err != nil {
		return nil, nil, err
	}
	err = r.writeCABundleFile(tlsDir, caCert.Raw)
	if err != nil {
		return nil, nil, err
	}

	return caPrivKey, caCert, nil

}

func (r *Renewer) renewTLS(tlsDir string, caCert *x509.Certificate, caPrivKey *rsa.PrivateKey, validity certValidityDuration) error {
	// sign the server cert with Datree CA
	serverPrivKey, serverCert, err := GenerateTLS(caCert, caPrivKey, validity)
	if err != nil {
		return err
	}

	// PEM encode the server cert and key
	serverCertPEM, err := certificateToPem(serverCert.Raw)
	if err != nil {
		return err
	}

	err = r.w.writeFile(filepath.Join(tlsDir, `tls.crt`), serverCertPEM)
	if err != nil {
		return err
	}

	serverPrivKeyPEM, err := privateKeyToPem(serverPrivKey)
	if err != nil {
		return err
	}

	err = r.w.writeFile(filepath.Join(tlsDir, `tls.key`), serverPrivKeyPEM)
	if err != nil {
		return err
	}

	return nil
}

func (r *Renewer) writeCABundleFile(tlsDir string, caCertificate []byte) error {
	// pem encode a bundle of the CA certificates
	// this is used to verify that the server is really the correct site you're talking to
	datreeCACertPEM, err := certificateToPem(caCertificate)
	if err != nil {
		return err
	}

	err = r.w.writeFile(filepath.Join(tlsDir, "ca-bundle.pem"), datreeCACertPEM)
	if err != nil {
		return err
	}

	return nil
}
