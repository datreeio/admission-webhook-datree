package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/datreeio/admission-webhook-datree/pkg/server"
	servicestate "github.com/datreeio/admission-webhook-datree/pkg/serviceState"

	"github.com/datreeio/datree/pkg/ciContext"
	"github.com/datreeio/datree/pkg/evaluation"

	k8sTypes "k8s.io/apimachinery/pkg/types"

	"github.com/datreeio/datree/pkg/cliClient"
	"github.com/datreeio/datree/pkg/httpClient"
)

type HTTPClient interface {
	Request(method string, resourceURI string, body interface{}, headers map[string]string) (httpClient.Response, error)
}

type CliClient struct {
	baseUrl          string
	httpClient       HTTPClient
	timeoutClient    HTTPClient
	httpErrors       []string
	networkValidator cliClient.NetworkValidator
	flagsHeaders     map[string]string
}

func NewCliServiceClient(url string, networkValidator cliClient.NetworkValidator, state *servicestate.ServiceState) *CliClient {
	httpClient := httpClient.NewClient(url, nil)
	return &CliClient{
		baseUrl:          url,
		httpClient:       httpClient,
		timeoutClient:    nil,
		httpErrors:       []string{},
		networkValidator: networkValidator,
		flagsHeaders: map[string]string{
			"x-cli-flags-policyName":  state.GetPolicyName(),
			"x-cli-flags-verbose":     state.GetVerbose(),
			"x-cli-flags-output":      state.GetOutput(),
			"x-cli-flags-noRecord":    state.GetNoRecord(),
			"x-cli-flags-enforce":     strconv.FormatBool(state.GetIsEnforceMode()),
			"x-cli-flags-clusterName": state.GetClusterName(),
		},
	}
}
func NewCustomCliServiceClient(baseUrl string, httpClient HTTPClient, timeoutClient HTTPClient, httpErrors []string, networkValidator cliClient.NetworkValidator, flagsHeaders map[string]string) *CliClient {
	return &CliClient{
		baseUrl:          baseUrl,
		httpClient:       httpClient,
		timeoutClient:    timeoutClient,
		httpErrors:       httpErrors,
		networkValidator: networkValidator,
		flagsHeaders:     flagsHeaders,
	}
}

type ClusterEvaluationPrerunDataResponse struct {
	cliClient.EvaluationPrerunDataResponse `json:",inline"`
	ActivePolicies                         []string                       `json:"activePolicies"`
	ActionOnFailure                        enums.ActionOnFailure          `json:"actionOnFailure"`
	IgnorePatterns                         []string                       `json:"ignorePatterns"`
	BypassPermissions                      servicestate.BypassPermissions `json:"bypassPermissions"`
}

func (c *CliClient) RequestClusterEvaluationPrerunData(tokenId string, clusterUuid k8sTypes.UID) (*ClusterEvaluationPrerunDataResponse, error) {
	if c.networkValidator.IsLocalMode() {
		return &ClusterEvaluationPrerunDataResponse{
			EvaluationPrerunDataResponse: cliClient.EvaluationPrerunDataResponse{
				IsPolicyAsCodeMode: true,
			},
		}, nil
	}

	res, err := c.httpClient.Request(http.MethodGet, "/cli/evaluation/policyCheck/tokens/"+tokenId+"/clusters/"+string(clusterUuid)+"/prerun?", nil, c.flagsHeaders)

	if err != nil {
		networkErr := c.networkValidator.IdentifyNetworkError(err)
		if networkErr != nil {
			return &ClusterEvaluationPrerunDataResponse{}, networkErr
		}

		if c.networkValidator.IsLocalMode() {
			return &ClusterEvaluationPrerunDataResponse{EvaluationPrerunDataResponse: cliClient.EvaluationPrerunDataResponse{
				IsPolicyAsCodeMode: true,
			}}, nil
		}

		return &ClusterEvaluationPrerunDataResponse{}, err
	}

	var evaluationPrerunDataResponse = &ClusterEvaluationPrerunDataResponse{EvaluationPrerunDataResponse: cliClient.EvaluationPrerunDataResponse{
		IsPolicyAsCodeMode: true,
	}}
	err = json.Unmarshal(res.Body, &evaluationPrerunDataResponse)
	if err != nil {
		return &ClusterEvaluationPrerunDataResponse{}, err
	}

	return evaluationPrerunDataResponse, nil
}

// SendEvaluationResult needed to override cliClient for evaluation
func (c *CliClient) SendEvaluationResult(request *cliClient.EvaluationResultRequest) (*cliClient.SendEvaluationResultsResponse, error) {
	return nil, nil
}

type OwnerReference struct {
	ApiVersion         string `json:"apiVersion"`
	Kind               string `json:"kind"`
	Name               string `json:"name"`
	Uid                string `json:"uid"`
	Controller         bool   `json:"controller"`
	BlockOwnerDeletion bool   `json:"blockOwnerDeletion"`
}

type ClusterRequestMetadata struct {
	ClusterUuid              k8sTypes.UID                        `json:"clusterUuid"`
	WebhookVersion           string                              `json:"webhookVersion"`
	CliEvaluationId          int                                 `json:"cliEvaluationId"`
	Token                    string                              `json:"token"`
	Skipped                  bool                                `json:"skipped"`
	Allowed                  bool                                `json:"allowed"`
	ResourceKind             string                              `json:"resourceKind"`
	ResourceName             string                              `json:"resourceName"`
	Managers                 []string                            `json:"managers"`
	PolicyName               string                              `json:"policyName"`
	K8sVersion               string                              `json:"k8sVersion"`
	Namespace                string                              `json:"namespace,omitempty"`
	ConfigMapScanningFilters server.ConfigMapScanningFiltersType `json:"configMapScanningFilters,omitempty"`
	Occurrences              int                                 `json:"occurrences"`
	OwnerReferences          []OwnerReference                    `json:"ownerReferences"`
}

type ClusterRequestMetadataBatchReqBody struct {
	MetadataLogs []*ClusterRequestMetadata `json:"metadataLogs"`
}

func (c *CliClient) SendRequestMetadataBatch(clusterRequestMetadataAggregator ClusterRequestMetadataBatchReqBody) {
	httpRes, err := c.httpClient.Request(http.MethodPost, "/cli/evaluation/clusterRequestMetadataBatch", clusterRequestMetadataAggregator, c.flagsHeaders)
	if err != nil {
		// using fmt.Printf instead of logger to avoid circular dependency
		fmt.Printf("SendRequestMetadataBatch status code: %d, err: %s \n", httpRes.StatusCode, err.Error())
	}
}

type WebhookEvaluationRequestData struct {
	EvaluationData evaluation.EvaluationRequestData
	WebhookVersion string
	ClusterUuid    k8sTypes.UID
	Namespace      string
	IsEnforceMode  bool
	Kind           string
	MetadataName   string
}

type BypassCriteriaType int

const (
	ServiceAccount BypassCriteriaType = iota
	UserAccount
	Group
)

func (s BypassCriteriaType) String() string {
	switch s {
	case ServiceAccount:
		return "serviceAccounts"
	case UserAccount:
		return "userAccounts"
	case Group:
		return "groups"
	default:
		return "unknown"
	}
}

type EvaluationResultRequest struct {
	ClientId                string                                      `json:"clientId"`
	Token                   string                                      `json:"token"`
	Metadata                *Metadata                                   `json:"metadata"`
	K8sVersion              string                                      `json:"k8sVersion"`
	PolicyName              string                                      `json:"policyName"`
	FailedYamlFiles         []string                                    `json:"failedYamlFiles"`
	FailedK8sFiles          []string                                    `json:"failedK8sFiles"`
	AllExecutedRules        []cliClient.RuleData                        `json:"allExecutedRules"`
	AllEvaluatedFiles       []cliClient.FileData                        `json:"allEvaluatedFiles"`
	PolicyCheckResults      map[string]map[string]*cliClient.FailedRule `json:"policyCheckResults"`
	ClusterUuid             k8sTypes.UID                                `json:"clusterUuid,omitempty"`
	Namespace               string                                      `json:"namespace,omitempty"`
	Kind                    string                                      `json:"kind"`
	MetadataName            string                                      `json:"metadataName"`
	MatchedBypassCriteria   *BypassCriteria                             `json:"matchedBypassCriteria,omitempty"`
	IsBypassedByPermissions bool                                        `json:"isBypassedByPermissions"`
}

type Metadata struct {
	Os                        string               `json:"os"`
	PlatformVersion           string               `json:"platformVersion"`
	KernelVersion             string               `json:"kernelVersion"`
	ClusterContext            *ClusterContext      `json:"clusterContext"`
	CIContext                 *ciContext.CIContext `json:"ciContext"`
	EvaluationDurationSeconds float64              `json:"evaluationDurationSeconds"`
}

type BypassCriteria struct {
	Type  BypassCriteriaType
	Value string
}

type ClusterContext struct {
	WebhookVersion string `json:"webhookVersion"`
	IsInCluster    bool   `json:"isInCluster"`
	IsEnforceMode  bool   `json:"isEnforceMode"`
}

func (c *CliClient) SaveWebhookEvaluationResults(request *EvaluationResultRequest) (*cliClient.SendEvaluationResultsResponse, error) {
	if c.networkValidator.IsLocalMode() {
		return &cliClient.SendEvaluationResultsResponse{}, nil
	}

	httpRes, err := c.httpClient.Request(http.MethodPost, "/cli/evaluation/policyCheck/result", request, c.flagsHeaders)
	if err != nil {
		networkErr := c.networkValidator.IdentifyNetworkError(err)
		if networkErr != nil {
			return &cliClient.SendEvaluationResultsResponse{}, networkErr
		}

		if c.networkValidator.IsLocalMode() {
			return &cliClient.SendEvaluationResultsResponse{}, nil
		}

		return nil, err
	}

	var res = &cliClient.SendEvaluationResultsResponse{}
	err = json.Unmarshal(httpRes.Body, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type SendEvaluationResultsResponse struct {
	EvaluationId  int    `json:"evaluationId"`
	PromptMessage string `json:"promptMessage,omitempty"`
}

type VersionRelatedMessagesResponse struct {
	CliVersion       string   `json:"cliVersion"`
	MessageTextArray []string `json:"messageTextArray"`
	MessageColor     string   `json:"messageColor"`
}

func (c *CliClient) GetVersionRelatedMessages(webhookVersion string) (*VersionRelatedMessagesResponse, error) {
	if c.networkValidator.IsLocalMode() {
		return nil, nil
	}
	if webhookVersion == "" {
		return nil, errors.New("can't get current webhook version")
	}
	httpRes, err := c.httpClient.Request(http.MethodGet, "/cli/messages/versions/"+webhookVersion+"/webhook", nil, c.flagsHeaders)
	if err != nil {
		return nil, err
	}

	var res = &VersionRelatedMessagesResponse{}

	err = json.Unmarshal(httpRes.Body, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type ReportK8sMetadataRequest struct {
	ClusterUuid     k8sTypes.UID          `json:"clusterUuid"`
	Token           string                `json:"token"`
	NodesCount      int                   `json:"nodesCount"`
	NodesCountErr   string                `json:"nodesCountErr"`
	ActionOnFailure enums.ActionOnFailure `json:"actionOnFailure"`
	K8sDistribution string                `json:"k8sDistribution"`
}

func (c *CliClient) ReportK8sMetadata(request *ReportK8sMetadataRequest) {
	_, err := c.httpClient.Request(http.MethodPost, "/cli/clusterEvents", request, c.flagsHeaders)
	if err != nil {
		fmt.Printf("Failed to report cluster metadata: %s\n", err.Error())
	}
}

type ReportErrorRequest struct {
	ClientId       string       `json:"clientId"`
	Token          string       `json:"token"`
	ClusterUuid    k8sTypes.UID `json:"clusterUuid"`
	ClusterName    string       `json:"clusterName"`
	K8sVersion     string       `json:"k8sVersion"`
	PolicyName     string       `json:"policyName"`
	IsEnforceMode  bool         `json:"isEnforceMode"`
	WebhookVersion string       `json:"webhookVersion"`
	ErrorMessage   string       `json:"errorMessage"`
	StackTrace     string       `json:"stackTrace"`
}

func (c *CliClient) ReportError(reportCliErrorRequest ReportErrorRequest, uri string) (StatusCode int, Error error) {
	headers := map[string]string{}
	res, err := c.httpClient.Request(
		http.MethodPost,
		"/cli/public"+uri,
		reportCliErrorRequest,
		headers,
	)
	return res.StatusCode, err
}
