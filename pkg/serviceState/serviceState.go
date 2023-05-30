package servicestate

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"os"
	"path/filepath"

	"github.com/datreeio/admission-webhook-datree/pkg/config"
	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/lithammer/shortuuid"
	"k8s.io/apimachinery/pkg/types"
)

var DATREE_CONFIG_FILE_DIR = `/config`

type ServiceState struct {
	clientId         string
	token            string
	clusterUuid      types.UID
	clusterName      string
	k8sVersion       string
	configFromHelm   bool
	policyName       string // strictly represents the policy name from the values.yaml file, we don't actually use it. We use the policies from prerunResponse.activePolicies for evaluation
	multiplePolicies *MultiplePolicies
	isEnforceMode    bool
	serviceVersion   string
	noRecord         string
	output           string
	verbose          string
}

func New() *ServiceState {
	return &ServiceState{
		clientId:         shortuuid.New(),
		token:            os.Getenv(enums.Token),
		clusterName:      os.Getenv(enums.ClusterName),
		configFromHelm:   os.Getenv(enums.ConfigFromHelm) != "false",
		policyName:       os.Getenv(enums.Policy),
		multiplePolicies: readMultiplePolicies(),
		isEnforceMode:    os.Getenv(enums.Enforce) == "true",
		serviceVersion:   config.WebhookVersion,
		noRecord:         os.Getenv(enums.NoRecord),
		output:           os.Getenv(enums.Output),
		verbose:          os.Getenv(enums.Verbose),
	}
}

func (s *ServiceState) SetClusterUuid(clusterUuid types.UID) {
	s.clusterUuid = clusterUuid
}

func (s *ServiceState) SetK8sVersion(k8sVersion string) {
	s.k8sVersion = k8sVersion
}

func (s *ServiceState) GetClientId() string {
	return s.clientId
}

func (s *ServiceState) GetToken() string {
	return s.token
}

func (s *ServiceState) GetClusterUuid() types.UID {
	return s.clusterUuid
}

func (s *ServiceState) GetClusterName() string {
	return s.clusterName
}

func (s *ServiceState) GetK8sVersion() string {
	return s.k8sVersion
}

func (s *ServiceState) GetConfigFromHelm() bool {
	return s.configFromHelm
}

func (s *ServiceState) GetPolicyName() string {
	return s.policyName
}

func (s *ServiceState) GetIsEnforceMode() bool {
	return s.isEnforceMode
}

// SetIsEnforceMode to override when we get cluster config in /prerun
func (s *ServiceState) SetIsEnforceMode(isEnforceMode bool) {
	s.isEnforceMode = isEnforceMode
}

func (s *ServiceState) GetServiceVersion() string {
	return s.serviceVersion
}

func (s *ServiceState) GetNoRecord() string {
	return s.noRecord
}

func (s *ServiceState) GetOutput() string {
	return s.output
}

func (s *ServiceState) GetVerbose() string {
	return s.verbose
}

func (s *ServiceState) GetMultiplePolicies() *MultiplePolicies {
	return s.multiplePolicies
}

type Namespaces struct {
	IncludePatterns []string `yaml:"includePatterns" json:"includePatterns"`
	ExcludePatterns []string `yaml:"excludePatterns" json:"excludePatterns"`
}

type PolicyWithNamespaces struct {
	Policy     string     `yaml:"policy" json:"policy"`
	Namespaces Namespaces `yaml:"namespaces" json:"namespaces"`
}

type MultiplePolicies = []PolicyWithNamespaces

func readMultiplePolicies() *MultiplePolicies {
	datreeMultiplePoliciesPath := filepath.Join(DATREE_CONFIG_FILE_DIR, "datreeMultiplePolicies")

	if _, err := os.Stat(datreeMultiplePoliciesPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println(fmt.Errorf("error read multiplePolicies on path : %s", datreeMultiplePoliciesPath))
		return nil
	}

	fileContent, readFileError := os.ReadFile(datreeMultiplePoliciesPath)
	if readFileError != nil {
		fmt.Println(readFileError)
		return nil
	}

	result := &MultiplePolicies{}
	fileUnmarshalError := yaml.Unmarshal(fileContent, &result)

	if fileUnmarshalError != nil {
		fmt.Println(fileUnmarshalError)
		return nil
	}

	return result
}
