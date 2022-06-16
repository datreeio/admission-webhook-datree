package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/datreeio/admission-webhook-datree/pkg/deploymentConfig"
)

func ValidateCertificate() (certPath string, keyPath string, err error) {
	tlsDir := `/run/secrets/tls`
	tlsCertFile := `tls.crt`
	tlsKeyFile := `tls.key`

	certPath = filepath.Join(tlsDir, tlsCertFile)
	keyPath = filepath.Join(tlsDir, tlsKeyFile)

	if deploymentConfig.ShouldValidateCertificate {
		if _, err := os.Stat(certPath); errors.Is(err, os.ErrNotExist) {
			return "", "", fmt.Errorf("cert file doesn't exist")
		}

		if _, err := os.Stat(keyPath); errors.Is(err, os.ErrNotExist) {
			return "", "", fmt.Errorf("key file doesn't exist")
		}
	}

	return certPath, keyPath, nil
}
