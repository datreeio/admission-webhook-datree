package servicestate

import (
	"os"

	"github.com/datreeio/admission-webhook-datree/pkg/config"
	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/lithammer/shortuuid"
	"k8s.io/apimachinery/pkg/types"
)

type ServiceState struct {
	clientId       string
	token          string
	clusterUuid    types.UID
	clusterName    string
	k8sVersion     string
	configFromHelm bool
	policyName     string
	isEnforceMode  bool
	serviceVersion string
	noRecord       string
	output         string
	verbose        string
}

func New() *ServiceState {
	return &ServiceState{
		clientId:       shortuuid.New(),
		token:          os.Getenv(enums.Token),
		clusterName:    os.Getenv(enums.ClusterName),
		configFromHelm: os.Getenv(enums.ConfigFromHelm) != "false",
		policyName:     os.Getenv(enums.Policy),
		isEnforceMode:  os.Getenv(enums.Enforce) == "true",
		serviceVersion: config.WebhookVersion,
		noRecord:       os.Getenv(enums.NoRecord),
		output:         os.Getenv(enums.Output),
		verbose:        os.Getenv(enums.Verbose),
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

// SetPolicyName to override when we get cluster config in /prerun
func (s *ServiceState) SetPolicyName(policyName string) {
	s.policyName = policyName
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
