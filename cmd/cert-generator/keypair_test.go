package main

import (
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateDatreeCA(t *testing.T) {
	type condition[T any] struct {
		compareFn func(actual T) bool
		msg       string
	}

	type output struct {
		privKey *condition[*rsa.PrivateKey]
		cert    *condition[*x509.Certificate]
		err     *condition[error]
	}

	type args struct {
		caKey    *rsa.PrivateKey
		validity certValidityDuration
	}

	type test struct {
		name     string
		input    args
		expected output
	}

	privKey, _ := rsa.GenerateKey(cryptorand.Reader, 2048)
	validity := certValidityDuration{
		begin: time.Now(),
		end:   time.Now().Add(1 * time.Hour),
	}

	tests := []test{
		{
			name:  "should create ca private key when key is not provided",
			input: args{caKey: nil, validity: validity},
			expected: output{
				privKey: &condition[*rsa.PrivateKey]{
					msg:       "private key should not be nil",
					compareFn: func(actual *rsa.PrivateKey) bool { return actual != nil },
				},
			},
		},
		{
			name:  "should create certificate with provided private key",
			input: args{caKey: privKey, validity: validity},
			expected: output{
				privKey: &condition[*rsa.PrivateKey]{
					msg:       "private key should equal to provided",
					compareFn: func(actual *rsa.PrivateKey) bool { return actual == privKey },
				},
				cert: &condition[*x509.Certificate]{
					msg: "cert should contains organization and validity fileds",
					compareFn: func(actual *x509.Certificate) bool {
						return actual.Subject.Organization[0] == "/CN=Datree Admission Controller Webhook CA" && actual.NotBefore.Round(time.Second).Sub(validity.begin.Round(time.Second)) == 0
					},
				},
				err: &condition[error]{
					msg:       "err should equal to nil",
					compareFn: func(actual error) bool { return actual == nil },
				},
			},
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			actualPrivKey, actualCert, actualErr := GenerateDatreeCA(ts.input.caKey, ts.input.validity)
			if ts.expected.privKey != nil {
				assert.Condition(t, func() bool { return ts.expected.privKey.compareFn(actualPrivKey) }, ts.expected.privKey.msg)
			}
			if ts.expected.cert != nil {
				assert.Condition(t, func() bool { return ts.expected.cert.compareFn(actualCert) }, ts.expected.cert.msg)
			}
			if ts.expected.err != nil {
				assert.Condition(t, func() bool { return ts.expected.err.compareFn(actualErr) }, ts.expected.err.msg)
			}
		})
	}
}

func TestGenerateTLS(t *testing.T) {
	type condition[T any] struct {
		compareFn func(actual T) bool
		msg       string
	}

	type output struct {
		privKey *condition[*rsa.PrivateKey]
		cert    *condition[*x509.Certificate]
		err     *condition[error]
	}

	type args struct {
		caKey    *rsa.PrivateKey
		caCert   *x509.Certificate
		validity certValidityDuration
	}

	type test struct {
		name     string
		input    args
		expected output
	}

	validity := certValidityDuration{
		begin: time.Now(),
		end:   time.Now().Add(1 * time.Hour),
	}

	caPrivKey, caCert, _ := GenerateDatreeCA(nil, validity)

	tests := []test{
		{name: "should create certificate",
			input: args{caKey: caPrivKey, caCert: caCert, validity: validity},
			expected: output{
				privKey: &condition[*rsa.PrivateKey]{
					msg:       "private key should not be nil",
					compareFn: func(actual *rsa.PrivateKey) bool { return actual != nil },
				},
				cert: &condition[*x509.Certificate]{
					msg: "cert should include DNS name and provided validity times",
					compareFn: func(actual *x509.Certificate) bool {
						return actual.DNSNames[0] == "test-webhook-dns-name" &&
							actual.Subject.CommonName == "/CN=test-webhook-dns-name" &&
							actual.NotBefore.Round(time.Second).Equal(validity.begin.Round(time.Second))
					},
				},
				err: &condition[error]{
					msg:       "error should be nil",
					compareFn: func(actual error) bool { return actual == nil }},
			},
		},
	}

	os.Setenv("WEBHOOK_SERVER_DNS", "test-webhook-dns-name")
	defer os.Unsetenv("WEBHOOK_SERVER_DNS")

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			actualPrivKey, actualCert, actualErr := GenerateTLS(ts.input.caCert, ts.input.caKey, ts.input.validity)
			if ts.expected.cert != nil {
				assert.Condition(t, func() bool { return ts.expected.cert.compareFn(actualCert) }, ts.expected.cert.msg)
			}
			if ts.expected.privKey != nil {
				assert.Condition(t, func() bool { return ts.expected.privKey.compareFn(actualPrivKey) }, ts.expected.privKey.msg)
			}
			if ts.expected.err != nil {
				assert.Condition(t, func() bool { return ts.expected.err.compareFn(actualErr) }, ts.expected.err.msg)
			}
		})
	}
}
