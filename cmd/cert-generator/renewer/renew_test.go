package renewer

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Condition[T any] func(actual T) bool

type ExpectedCondition[T any] struct {
	condition Condition[T]
	message   string
}

type Expected[T any] struct {
	value   T
	message string
}

type FileWriterMock struct {
	mock.Mock
}

func (m *FileWriterMock) WriteFile(filepath string, sCert *bytes.Buffer) error {
	args := m.Called(filepath, sCert)
	return args.Error(0)
}

func TestRenewCA(t *testing.T) {
	type testArgs struct {
		tlsDir string
	}

	type testCase struct {
		args                    *testArgs
		expectedErr             Expected[error]
		expectedPrivKey         ExpectedCondition[*rsa.PrivateKey]
		expectedCert            ExpectedCondition[*x509.Certificate]
		writerMockExpectedCalls *[]*mock.Call
	}

	tlsDir := "testdata"

	tests := map[string]*testCase{
		"should create CA certificate and write file": {
			args: &testArgs{tlsDir: tlsDir},
			expectedPrivKey: ExpectedCondition[*rsa.PrivateKey]{
				condition: func(actual *rsa.PrivateKey) bool {
					return actual.Validate() == nil
				},
				message: "should return a valid private key",
			},
			expectedErr: Expected[error]{
				value:   nil,
				message: "should not return an error",
			},
			expectedCert: ExpectedCondition[*x509.Certificate]{
				condition: func(actual *x509.Certificate) bool {
					return actual != nil && actual.IsCA && actual.Subject.Organization[0] == "/CN=Datree Admission Controller Webhook CA" && actual.NotBefore.Add(5*24*365*time.Hour).Equal(actual.NotAfter)
				},
				message: "should return a certificate",
			},
			writerMockExpectedCalls: &[]*mock.Call{
				{
					Method: "WriteFile",
					Arguments: mock.Arguments{
						filepath.Join(tlsDir, "ca-bundle.pem"),
						mock.AnythingOfType("*bytes.Buffer"),
					},
				},
			},
		},
	}

	writerMock := &FileWriterMock{}
	writerMock.On("WriteFile", mock.Anything, mock.Anything).Return(nil)
	renewer := CertRenewer{w: writerMock, tlsValidityDuration: 5 * 365 * 24 * time.Hour}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualPrivKey, actualCert, actualErr := renewer.RenewCA(tc.args.tlsDir)
			assert.Equal(t, tc.expectedErr.value, actualErr, tc.expectedErr.message)
			assert.True(t, tc.expectedCert.condition(actualCert), tc.expectedCert.message)
			assert.True(t, tc.expectedPrivKey.condition(actualPrivKey), tc.expectedPrivKey.message)
			writerMock.ExpectedCalls = *tc.writerMockExpectedCalls
			writerMock.AssertExpectations(t)
		})
	}

}

func TestRenewTLS(t *testing.T) {
	type testArgs struct {
		tlsDir    string
		caCert    *x509.Certificate
		caPrivKey *rsa.PrivateKey
	}

	type testCase struct {
		args                    *testArgs
		expectedErr             Expected[error]
		writerMockExpectedCalls *[]*mock.Call
	}

	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	tlsDir := "testdata"

	tests := map[string]*testCase{
		"should create server certificate and write cert and key": {
			args: &testArgs{tlsDir: tlsDir, caCert: serverCertificateMock(), caPrivKey: privKey},
			expectedErr: Expected[error]{
				value:   nil,
				message: "should not return an error",
			},
			writerMockExpectedCalls: &[]*mock.Call{
				{
					Method: "WriteFile",
					Arguments: mock.Arguments{
						filepath.Join(tlsDir, "tls.crt"),
						mock.AnythingOfType("*bytes.Buffer"),
					},
				},
				{
					Method: "WriteFile",
					Arguments: mock.Arguments{
						filepath.Join(tlsDir, "tls.key"),
						mock.AnythingOfType("*bytes.Buffer"),
					},
				},
			},
		},
	}

	writerMock := &FileWriterMock{}
	writerMock.On("WriteFile", mock.Anything, mock.Anything).Return(nil)
	renewer := CertRenewer{w: writerMock, tlsValidityDuration: 5 * 365 * 24 * time.Hour}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualErr := renewer.RenewTLS(tc.args.tlsDir, tc.args.caCert, tc.args.caPrivKey)
			assert.Equal(t, tc.expectedErr.value, actualErr, tc.expectedErr.message)
			writerMock.ExpectedCalls = *tc.writerMockExpectedCalls
			writerMock.AssertExpectations(t)
		})
	}

}

func serverCertificateMock() *x509.Certificate {
	return &x509.Certificate{
		DNSNames: []string{
			"localhost",
		},
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("/CN=%v", "localhost"),
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * time.Hour),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
}
