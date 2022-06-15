package clients

import (
	"encoding/json"
	"github.com/datreeio/datree/pkg/ciContext"
	"github.com/datreeio/datree/pkg/evaluation"
	"net/http"

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

func (c *CliClient) CreateToken() (*cliClient.CreateTokenResponse, error) {
	if c.networkValidator.IsLocalMode() {
		return &cliClient.CreateTokenResponse{}, nil
	}

	headers := map[string]string{}
	res, err := c.httpClient.Request(http.MethodPost, "/cli/tokens/", nil, headers)

	if err != nil {
		networkErr := c.networkValidator.IdentifyNetworkError(err.Error())
		if networkErr != nil {
			return nil, networkErr
		}

		if c.networkValidator.IsLocalMode() {
			return &cliClient.CreateTokenResponse{}, nil
		}

		return nil, err
	}

	createTokenResponse := &cliClient.CreateTokenResponse{}
	err = json.Unmarshal(res.Body, &createTokenResponse)

	if err != nil {
		return nil, err
	}

	return createTokenResponse, nil
}

func (c *CliClient) RequestEvaluationPrerunData(tokenId string) (*cliClient.EvaluationPrerunDataResponse, error) {
	if c.networkValidator.IsLocalMode() {
		return &cliClient.EvaluationPrerunDataResponse{IsPolicyAsCodeMode: true}, nil
	}

	res, err := c.httpClient.Request(http.MethodGet, "/cli/evaluation/policyCheck/tokens/"+tokenId+"/prerun?", nil, c.flagsHeaders)

	if err != nil {
		networkErr := c.networkValidator.IdentifyNetworkError(err.Error())
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

type WebhookEvaluationRequestData struct {
	EvaluationData evaluation.EvaluationRequestData
	WebhookVersion string
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
}

func (c *CliClient) SendWebhookEvaluationResult(request *EvaluationResultRequest) (*cliClient.SendEvaluationResultsResponse, error) {
	if c.networkValidator.IsLocalMode() {
		return &cliClient.SendEvaluationResultsResponse{}, nil
	}
	httpRes, err := c.httpClient.Request(http.MethodPost, "/cli/evaluation/policyCheck/result", request, c.flagsHeaders)
	if err != nil {
		networkErr := c.networkValidator.IdentifyNetworkError(err.Error())
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
