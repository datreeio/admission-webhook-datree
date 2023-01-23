package servicestate

import "k8s.io/apimachinery/pkg/types"

const (
	WEBHOOK string = "Webhook"
)

type ServiceState struct {
	ClientId       string    `json:"clientId"`
	Token          string    `json:"token"`
	ClusterUuid    types.UID `json:"clusterUuid"`
	ClusterName    string    `json:"clusterName"`
	K8sVersion     string    `json:"k8sVersion"`
	PolicyName     string    `json:"policyName"`
	IsEnforceMode  bool      `json:"isEnforceMode"`
	ServiceVersion string    `json:"serviceVersion"`
	ServiceType    string    `json:"serviceType"`
}

var state *ServiceState

func GetState() *ServiceState {
	if state == nil {
		state = &ServiceState{}
	}

	return state
}

func (s *ServiceState) Get() *ServiceState {
	return s
}

func (s *ServiceState) SetClientId(clientId string) {
	s.ClientId = clientId
}

func (s *ServiceState) SetToken(token string) {
	s.Token = token
}

func (s *ServiceState) SetClusterUuid(clusterUuid types.UID) {
	s.ClusterUuid = clusterUuid
}

func (s *ServiceState) SetClusterName(clusterName string) {
	s.ClusterName = clusterName
}

func (s *ServiceState) SetK8sVersion(k8sVersion string) {
	s.K8sVersion = k8sVersion
}

func (s *ServiceState) SetPolicyName(policyName string) {
	s.PolicyName = policyName
}

func (s *ServiceState) SetIsEnforceMode(isEnforceMode bool) {
	s.IsEnforceMode = isEnforceMode
}

func (s *ServiceState) SetServiceVersion(serviceVersion string) {
	s.ServiceVersion = serviceVersion
}

func (s *ServiceState) SetServiceType(serviceType string) {
	s.ServiceType = serviceType
}
