package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/datreeio/admission-webhook-datree/pkg/deploymentConfig"
	"gopkg.in/yaml.v2"
)

type ConfigMapScanningFiltersType struct {
	SkipList []string `yaml:"skipList" json:"skipList"`
}

var ConfigMapScanningFilters = ConfigMapScanningFiltersType{}

func InitServerVars() error {
	skipList, err := readConfigScanningFilters()

	if err != nil {
		return err
	}

	ConfigMapScanningFilters.SkipList = skipList
	return nil
}

func validateFileExistence(filePath string) bool {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func getScanningFilterFromConfigMap(filePath string) ([]string, error) {
	var configMapScanningFilters []string
	fileContent, readFileError := os.ReadFile(filePath)
	if readFileError != nil {
		return nil, readFileError
	}

	fileUnmarshalError := yaml.Unmarshal(fileContent, &configMapScanningFilters)
	if fileUnmarshalError != nil {
		return nil, fileUnmarshalError
	}

	return configMapScanningFilters, nil
}

func readConfigScanningFilters() (skipList []string, err error) {
	configDir := `/config`
	configSkipListPath := filepath.Join(configDir, `skiplist`)
	datreeSkipConfigDir := filepath.Join(configDir, `datreeSkipList`)
	datreeSkipListPath := filepath.Join(datreeSkipConfigDir, `datreeSkipList`)
	skipListPaths := []string{datreeSkipListPath, configSkipListPath}
	skipLists := []string{}

	for _, skipListPath := range skipListPaths {
		if validateFileExistence(skipListPath) {
			skipList, err = getScanningFilterFromConfigMap(skipListPath)
			if err != nil {
				return nil, err
			}
			skipLists = append(skipLists, skipList...)

		}
	}
	return skipLists, nil
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
