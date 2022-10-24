package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/datreeio/admission-webhook-datree/pkg/deploymentConfig"
	"gopkg.in/yaml.v2"
)

type ConfigAllowedListsType struct {
	SkipList     []string `yaml:"skipList"`
	ValidateList []string `yaml:"validateList"`
}

var ConfigAllowedLists ConfigAllowedListsType

func InitServerVars() error {
	skipList, validateList, err := readDatreeWebhookConfigMap()

	ConfigAllowedLists = ConfigAllowedListsType{
		SkipList:     skipList,
		ValidateList: validateList,
	}

	if err != nil {
		return err
	}

	return nil
}

func validateFileExistence(filePath string) bool {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func getConfigmapFromPath(filePath string) ([]string, error) {
	var configMap []string
	fileContent, readFileError := os.ReadFile(filePath)
	if readFileError != nil {
		return nil, readFileError
	}

	fileUnmarshalError := yaml.Unmarshal([]byte(fileContent), &configMap)
	if fileUnmarshalError != nil {
		return nil, fileUnmarshalError
	}

	return configMap, nil
}

func readDatreeWebhookConfigMap() (skipList []string, validateList []string, err error) {
	configDir := `/config`
	configSkipListPath := filepath.Join(configDir, `skiplist`)
	validateListPath := filepath.Join(configDir, `validatelist`)

	if validateFileExistence(configSkipListPath) {
		skipList, err = getConfigmapFromPath(configSkipListPath)
		if err != nil {
			return nil, nil, err
		}
	}

	if validateFileExistence(validateListPath) {
		validateList, err = getConfigmapFromPath(validateListPath)
		if err != nil {
			return nil, nil, err
		}
	}

	return skipList, validateList, nil
}

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
