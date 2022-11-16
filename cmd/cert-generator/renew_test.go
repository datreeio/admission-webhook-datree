package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFileWriter struct {
	mock.Mock
}

func (m *mockFileWriter) writeFile(filepath string, sCert *bytes.Buffer) error {
	args := m.Called(filepath, sCert)
	return args.Error(0)
}

func TestRenewCA(t *testing.T) {
	type args struct {
		tlsDir   string
		validity certValidityDuration
	}

	type condition[T any] struct {
		compareFn func(actual T) bool
		msg       string
	}

	type output[T any, K any, U any] struct {
		privKey condition[T]
		cert    condition[K]
		err     condition[U]
	}

	type test struct {
		name     string
		input    args
		expected output[*rsa.PrivateKey, *x509.Certificate, error]
	}

	validity := certValidityDuration{
		begin: time.Now(),
		end:   time.Now().Add(1 * time.Hour),
	}

	tlsDir := t.TempDir()
	tests := []test{
		{
			name:  "should generate CA write caBundle and return datree CA ",
			input: args{tlsDir: tlsDir, validity: validity},
			expected: output[*rsa.PrivateKey, *x509.Certificate, error]{
				err:     condition[error]{msg: "err should be nil", compareFn: func(actual error) bool { return actual == nil }},
				privKey: condition[*rsa.PrivateKey]{msg: "priv key should be valid and not nil", compareFn: func(actual *rsa.PrivateKey) bool { err := actual.Validate(); return err == nil && actual != nil }},
				cert:    condition[*x509.Certificate]{msg: "cert should be not nil", compareFn: func(actual *x509.Certificate) bool { return actual != nil }},
			},
		},
	}

	fileWriter := &mockFileWriter{}
	fileWriter.On("writeFile", mock.Anything, mock.Anything).Return(nil)

	r := newRenewer(fileWriter)

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			actualPrivKey, actualCert, actualErr := r.renewCA(ts.input.tlsDir, ts.input.validity)
			assert.Condition(t, func() bool { return ts.expected.privKey.compareFn(actualPrivKey) }, ts.expected.privKey.msg)
			assert.Condition(t, func() bool { return ts.expected.cert.compareFn(actualCert) }, ts.expected.cert.msg)
			assert.Condition(t, func() bool { return ts.expected.err.compareFn(actualErr) }, ts.expected.err.msg)
			fileWriter.AssertCalled(t, "writeFile", filepath.Join(tlsDir, "ca-bundle.pem"), mock.Anything)
		})
	}
}
