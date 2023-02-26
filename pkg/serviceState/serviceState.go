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
	policyName     string
	isEnforceMode  bool
	serviceVersion string
	NoRecord       string
	Output         string
	Verbose        string
}

func New() *ServiceState {

	return &ServiceState{
		clientId:       shortuuid.New(),
		token:          os.Getenv(enums.Token),
		clusterName:    os.Getenv(enums.ClusterName),
		policyName:     os.Getenv(enums.Policy),
		isEnforceMode:  os.Getenv(enums.Enforce) == "true",
		serviceVersion: config.WebhookVersion,
		NoRecord:       os.Getenv(enums.NoRecord),
		Output:         os.Getenv(enums.Output),
		Verbose:        os.Getenv(enums.Verbose),
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

func (s *ServiceState) GetPolicyName() string {
	return s.policyName
}

func (s *ServiceState) GetIsEnforceMode() bool {
	return s.isEnforceMode
}

func (s *ServiceState) GetServiceVersion() string {
	return s.serviceVersion
}

func (s *ServiceState) GetNoRecord() string {
	return s.NoRecord
}

func (s *ServiceState) GetOutput() string {
	return s.Output
}

func (s *ServiceState) GetVerbose() string {
	return s.Verbose
}
