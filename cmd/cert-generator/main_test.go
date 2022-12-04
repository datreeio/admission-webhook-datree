package main

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	tlsDir = "/tmp"
	m.Run()
}

func TestRenewCA(t *testing.T) {
	main()

	// Read server cert
	certPEM, err := ioutil.ReadFile("/tmp/tls.crt")
	assert.True(t, err == nil)

	// Read ca bundle
	caPEM, err := ioutil.ReadFile("/tmp/ca-bundle.pem")
	assert.True(t, err == nil)

	// Parse ca bundle and append to cert-pool
	roots := x509.NewCertPool()
	assert.True(t, roots.AppendCertsFromPEM([]byte(caPEM)))

	// Parse server cert
	block, _ := pem.Decode([]byte(certPEM))
	assert.True(t, block != nil)
	cert, err := x509.ParseCertificate(block.Bytes)
	assert.True(t, err == nil)

	// Verify server cert is valid and signed by ca
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: x509.NewCertPool(),
	}

	_, err = cert.Verify(opts)
	assert.True(t, err == nil)
}
