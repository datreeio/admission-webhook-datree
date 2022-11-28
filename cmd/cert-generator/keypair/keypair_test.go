package keypair

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestGenerateCA(t *testing.T) {
	type testArgs struct {
		key   *rsa.PrivateKey
		templ *x509.Certificate
	}

	type testCase struct {
		args         *testArgs
		expectedErr  Expected[error]
		expectedKey  ExpectedCondition[*rsa.PrivateKey]
		expectedCert ExpectedCondition[*x509.Certificate]
	}

	tests := map[string]*testCase{
		"should create CA certificate and private key when key is nil": {
			args: &testArgs{
				key:   nil,
				templ: caCertificateTemplateMock(),
			},
			expectedCert: ExpectedCondition[*x509.Certificate]{
				condition: func(actual *x509.Certificate) bool {
					return actual.IsCA && actual.SerialNumber.Cmp(big.NewInt(1)) == 0 && actual.Subject.Organization[0] == "/CN=datree.io" && (actual.NotAfter.Sub(actual.NotBefore) == 10*time.Hour)
				},
				message: "certificate should be a valid CA certificate and contain given template values",
			},
			expectedErr: Expected[error]{
				value:   nil,
				message: "should not return an error",
			},
			expectedKey: ExpectedCondition[*rsa.PrivateKey]{
				condition: func(actual *rsa.PrivateKey) bool {
					return actual.Validate() == nil
				},
				message: "private key should be valid",
			},
		},
		"should create CA certificate and private key when key is not nil": {
			args: &testArgs{
				key:   nil,
				templ: caCertificateTemplateMock(),
			},
			expectedCert: ExpectedCondition[*x509.Certificate]{
				condition: func(actual *x509.Certificate) bool {
					return actual.IsCA
				},
				message: "certificate should be a valid CA certificate",
			},
			expectedErr: Expected[error]{
				value:   nil,
				message: "should not return an error",
			},
			expectedKey: ExpectedCondition[*rsa.PrivateKey]{
				condition: func(actual *rsa.PrivateKey) bool {
					return actual.Validate() == nil
				},
				message: "private key should be valid",
			},
		},
		"should return an error when template is nil": {
			args: &testArgs{key: nil, templ: nil},
			expectedCert: ExpectedCondition[*x509.Certificate]{
				condition: func(actual *x509.Certificate) bool { return actual == nil },
				message:   "certificate should be nil",
			},
			expectedErr: Expected[error]{
				value:   fmt.Errorf(ErrInvalidTemplate),
				message: "should return an error",
			},
			expectedKey: ExpectedCondition[*rsa.PrivateKey]{
				condition: func(actual *rsa.PrivateKey) bool { return actual == nil },
				message:   "private key should be nil",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualyPrivKey, actualCert, actualErr := GenerateCA(tc.args.key, tc.args.templ)
			assert.Equal(t, tc.expectedErr.value, actualErr, tc.expectedErr.message)
			assert.True(t, tc.expectedCert.condition(actualCert), tc.expectedCert.message)
			assert.True(t, tc.expectedKey.condition(actualyPrivKey), tc.expectedKey.message)
		})
	}
}

func TestGenerateTLS(t *testing.T) {
	type testArgs struct {
		caCert *x509.Certificate
		caKey  *rsa.PrivateKey
		templ  *x509.Certificate
	}

	type testCase struct {
		args         *testArgs
		expectedErr  Expected[error]
		expectedKey  ExpectedCondition[*rsa.PrivateKey]
		expectedCert ExpectedCondition[*x509.Certificate]
	}

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	serverCert := serverCertificateMock()

	tests := map[string]*testCase{
		"should create CA certificate and private key when key is nil": {
			args: &testArgs{
				caKey:  privateKey,
				caCert: caCertificateTemplateMock(),
				templ:  serverCert,
			},
			expectedCert: ExpectedCondition[*x509.Certificate]{
				condition: func(actual *x509.Certificate) bool {
					return !actual.IsCA && actual.SerialNumber.Cmp(serverCert.SerialNumber) == 0 && actual.DNSNames[0] == serverCert.DNSNames[0] && (actual.NotAfter.Sub(actual.NotBefore) == 10*time.Hour)
				},
				message: "server certificate should be valid certificate and contain given template values",
			},
			expectedErr: Expected[error]{
				value:   nil,
				message: "should not return an error",
			},
			expectedKey: ExpectedCondition[*rsa.PrivateKey]{
				condition: func(actual *rsa.PrivateKey) bool {
					return actual.Validate() == nil
				},
				message: "private key should be valid",
			},
		},
		"should return error when private key is nil": {
			args: &testArgs{
				caKey:  nil,
				caCert: caCertificateTemplateMock(),
				templ:  serverCert,
			},
			expectedCert: ExpectedCondition[*x509.Certificate]{
				condition: func(actual *x509.Certificate) bool {
					return actual == nil
				},
				message: "server certificate should be nil",
			},
			expectedErr: Expected[error]{
				value:   fmt.Errorf(ErrInvalidArgs),
				message: "should return an error",
			},
			expectedKey: ExpectedCondition[*rsa.PrivateKey]{
				condition: func(actual *rsa.PrivateKey) bool {
					return actual == nil
				},
				message: "private key should be nil",
			},
		},
		"should return an error when template is nil": {
			args: &testArgs{caKey: privateKey, caCert: caCertificateTemplateMock(), templ: nil},
			expectedCert: ExpectedCondition[*x509.Certificate]{
				condition: func(actual *x509.Certificate) bool { return actual == nil },
				message:   "certificate should be nil",
			},
			expectedErr: Expected[error]{
				value:   fmt.Errorf(ErrInvalidArgs),
				message: "should return an error",
			},
			expectedKey: ExpectedCondition[*rsa.PrivateKey]{
				condition: func(actual *rsa.PrivateKey) bool { return actual == nil },
				message:   "private key should be nil",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualPrivKey, actualCert, actualErr := GenerateTLS(tc.args.caCert, tc.args.caKey, tc.args.templ)
			assert.Equal(t, tc.expectedErr.value, actualErr, tc.expectedErr.message)
			assert.True(t, tc.expectedCert.condition(actualCert), tc.expectedCert.message)
			assert.True(t, tc.expectedKey.condition(actualPrivKey), tc.expectedKey.message)
		})
	}
}

func caCertificateTemplateMock() *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"/CN=datree.io"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
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
