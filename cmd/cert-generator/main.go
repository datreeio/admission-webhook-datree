package main

import (
	"bytes"
	"os"

	"github.com/datreeio/admission-webhook-datree/cmd/cert-generator/renewer"
	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
)

type fileWriter struct{}

// WriteFile writes data in the file at the given path
func (fw *fileWriter) WriteFile(filepath string, sCert *bytes.Buffer) error {
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

func main() {
	tlsDir, isFound := os.LookupEnv("WEBHOOK_CERTS_DIR")
	if !isFound {
		loggerUtil.Log("required directory for certificates is missing, verify env varaible WEBHOOK_CERTS_DIR in deployment")
		return
	}

	err := os.MkdirAll(tlsDir, 0666)
	if err != nil {
		loggerUtil.Log(err.Error())
		return
	}

	renewer := renewer.NewCertRenewer(&fileWriter{})

	caPrivKey, caCert, err := renewer.RenewCA(tlsDir)
	if err != nil {
		loggerUtil.Log(err.Error())
		return
	}

	err = renewer.RenewTLS(tlsDir, caCert, caPrivKey)
	if err != nil {
		loggerUtil.Log(err.Error())
		return
	}

	loggerUtil.Log("horray! successfully generated self-signed CA and signed webhook server certificate using this CA!")
}
