package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"github.com/datreeio/admission-webhook-datree/pkg/server"

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

func NewCliServiceClient(url string, networkValidator cliClient.NetworkValidator) *CliClient {
	httpClient := httpClient.NewClient(url, nil)
	return &CliClient{
		baseUrl:          url,
		httpClient:       httpClient,
		timeoutClient:    nil,
		httpErrors:       []string{},
		networkValidator: networkValidator,
		flagsHeaders:     make(map[string]string),
	}
}

func (c *CliClient) RequestEvaluationPrerunData(tokenId string) (*cliClient.EvaluationPrerunDataResponse, error) {
	if c.networkValidator.IsLocalMode() {
		return &cliClient.EvaluationPrerunDataResponse{IsPolicyAsCodeMode: true}, nil
	}

	res, err := c.httpClient.Request(http.MethodGet, "/cli/evaluation/policyCheck/tokens/"+tokenId+"/prerun?", nil, c.flagsHeaders)

	if err != nil {
		networkErr := c.networkValidator.IdentifyNetworkError(err)
		if networkErr != nil {
			return &cliClient.EvaluationPrerunDataResponse{}, networkErr
		}

		if c.networkValidator.IsLocalMode() {
			return &cliClient.EvaluationPrerunDataResponse{IsPolicyAsCodeMode: true}, nil
		}

		return &cliClient.EvaluationPrerunDataResponse{}, err
	}

	var evaluationPrerunDataResponse = &cliClient.EvaluationPrerunDataResponse{IsPolicyAsCodeMode: true}
	err = json.Unmarshal(res.Body, &evaluationPrerunDataResponse)
	if err != nil {
		return &cliClient.EvaluationPrerunDataResponse{}, err
	}

	return evaluationPrerunDataResponse, nil
}

// SendEvaluationResult needed to override cliClient for evaluation
func (c *CliClient) SendEvaluationResult(request *cliClient.EvaluationResultRequest) (*cliClient.SendEvaluationResultsResponse, error) {
	return nil, nil
}

type ClusterRequestMetadata struct {
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
}

type ClusterRequestMetadataBatchReqBody struct {
	MetadataLogs []*ClusterRequestMetadata `json:"metadataLogs"`
}

func (c *CliClient) SendRequestMetadataBatch(clusterRequestMetadataAggregator ClusterRequestMetadataBatchReqBody) {
	httpRes, err := c.httpClient.Request(http.MethodPost, "/cli/evaluation/clusterRequestMetadataBatch", clusterRequestMetadataAggregator, c.flagsHeaders)
	if err != nil {
		logger.LogUtil(fmt.Sprintf("SendRequestMetadataBatch status code: %d, err: %s", httpRes.StatusCode, err.Error()))
	}
}

type WebhookEvaluationRequestData struct {
	EvaluationData evaluation.EvaluationRequestData
	WebhookVersion string
	ClusterUuid    k8sTypes.UID
	Namespace      string
	IsEnforceMode  bool
}

type EvaluationResultRequest struct {
	ClientId           string                                      `json:"clientId"`
	Token              string                                      `json:"token"`
	Metadata           *Metadata                                   `json:"metadata"`
	K8sVersion         string                                      `json:"k8sVersion"`
	PolicyName         string                                      `json:"policyName"`
	FailedYamlFiles    []string                                    `json:"failedYamlFiles"`
	FailedK8sFiles     []string                                    `json:"failedK8sFiles"`
	AllExecutedRules   []cliClient.RuleData                        `json:"allExecutedRules"`
	AllEvaluatedFiles  []cliClient.FileData                        `json:"allEvaluatedFiles"`
	PolicyCheckResults map[string]map[string]*cliClient.FailedRule `json:"policyCheckResults"`
	ClusterUuid        k8sTypes.UID                                `json:"clusterUuid,omitempty"`
	Namespace          string                                      `json:"namespace,omitempty"`
}

type Metadata struct {
	Os                        string               `json:"os"`
	PlatformVersion           string               `json:"platformVersion"`
	KernelVersion             string               `json:"kernelVersion"`
	ClusterContext            *ClusterContext      `json:"clusterContext"`
	CIContext                 *ciContext.CIContext `json:"ciContext"`
	EvaluationDurationSeconds float64              `json:"evaluationDurationSeconds"`
}

type ClusterContext struct {
	WebhookVersion string `json:"webhookVersion"`
	IsInCluster    bool   `json:"isInCluster"`
	IsEnforceMode  bool   `json:"isEnforceMode"`
}

func (c *CliClient) SendWebhookEvaluationResult(request *EvaluationResultRequest) (*cliClient.SendEvaluationResultsResponse, error) {
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
	ClusterUuid   k8sTypes.UID `json:"clusterUuid"`
	Token         string       `json:"token"`
	NodesCount    int          `json:"nodesCount"`
	NodesCountErr string       `json:"nodesCountErr"`
}

func (c *CliClient) ReportK8sMetadata(request *ReportK8sMetadataRequest) {
	c.httpClient.Request(http.MethodPost, "/cli/clusterEvents", request, c.flagsHeaders)
}
