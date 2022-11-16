package main

import (
	"bytes"
	"os"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
)

type fileWriter struct{}

// WriteFile writes data in the file at the given path
func (fw *fileWriter) writeFile(filepath string, sCert *bytes.Buffer) error {
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

	now := time.Now()
	validity := certValidityDuration{begin: now.Add(-1 * time.Hour), end: now.Add(5 * 365 * 24 * time.Hour)} // 5 years validity

	renewer := newRenewer(&fileWriter{})

	caPrivKey, caCert, err := renewer.renewCA(tlsDir, validity)
	if err != nil {
		loggerUtil.Log(err.Error())
		return
	}

	err = renewer.renewTLS(tlsDir, caCert, caPrivKey, validity)
	if err != nil {
		loggerUtil.Log(err.Error())
		return
	}

	loggerUtil.Log("horray! successfully generated self-signed CA and signed webhook server certificate using this CA!")
}
