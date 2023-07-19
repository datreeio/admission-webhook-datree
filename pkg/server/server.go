package server

import (
	"errors"
	servicestate "github.com/datreeio/admission-webhook-datree/pkg/serviceState"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ConfigMapScanningFiltersType struct {
	SkipList []string `yaml:"skipList" json:"skipList"`
}

var ConfigMapScanningFilters = ConfigMapScanningFiltersType{}

func InitSkipList() error {
	skipList, err := readConfigScanningFilters()

	if err != nil {
		return err
	}

	ConfigMapScanningFilters.SkipList = skipList
	return nil
}

func OverrideSkipList(skipList []string) {
	ConfigMapScanningFilters.SkipList = skipList
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
	configSkipListPath := filepath.Join(servicestate.DATREE_CONFIG_FILE_DIR, `skiplist`)
	datreeSkipListPath := filepath.Join(servicestate.DATREE_CONFIG_FILE_DIR, `datreeSkipList`)
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
