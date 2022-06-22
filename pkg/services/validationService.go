package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/config"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"

	policyFactory "github.com/datreeio/datree/bl/policy"
	"github.com/datreeio/datree/pkg/ciContext"
	baseCliClient "github.com/datreeio/datree/pkg/cliClient"
	"github.com/datreeio/datree/pkg/deploymentConfig"
	"github.com/datreeio/datree/pkg/evaluation"
	"github.com/datreeio/datree/pkg/extractor"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/datreeio/datree/pkg/printer"
	"github.com/datreeio/datree/pkg/utils"

	cliClient "github.com/datreeio/admission-webhook-datree/pkg/clients"

	"github.com/ghodss/yaml"
	"github.com/lithammer/shortuuid"
	admission "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type AdmissionReqOptions struct {
	Kind         string `json:"kind"`
	ApiVersion   string `json:"apiVersion"`
	FieldManager string `json:"fieldManager,omitempty"`
}

func Validate(admissionReviewReq *admission.AdmissionReview) *admission.AdmissionReview {
	startTime := time.Now()
	msg := "We're good!"
	allowed := true
	var err error
	warningMessages := []string{}

	validator := networkValidator.NewNetworkValidator()
	cliClient := cliClient.NewCliServiceClient(deploymentConfig.URL, validator)

	ciContext := ciContext.Extract()

	clusterK8sVersion := getK8sVersion()

	var reqOptions AdmissionReqOptions
	if err := json.Unmarshal(admissionReviewReq.Request.Options.Raw, &reqOptions); err != nil {
		panic(err)
	}

	if reqOptions.FieldManager == "" {
		return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, allowed, msg, warningMessages)
	}

	token, err := getToken(cliClient)
	if err != nil {
		panic(err)
	}

	clientId := getClientId()

	policyName := os.Getenv(enums.Policy)
	prerunData, err := cliClient.RequestEvaluationPrerunData(token)
	if err != nil {
		warningMessages = append(warningMessages, err.Error())
	}
	policy, err := policyFactory.CreatePolicy(prerunData.PoliciesJson, policyName, prerunData.RegistrationURL)
	if err != nil {
		panic(err)
	}

	filesConfigurations := getFileConfiguration(admissionReviewReq.Request)

	policyCheckData := evaluation.PolicyCheckData{
		FilesConfigurations: filesConfigurations,
		IsInteractiveMode:   false,
		PolicyName:          policy.Name,
		Policy:              policy,
	}

	evaluator := evaluation.New(cliClient, ciContext)
	policyCheckResults, _ := evaluator.Evaluate(policyCheckData)

	results := policyCheckResults.FormattedResults
	passedPolicyCheckCount := 0
	if results.EvaluationResults != nil {
		passedPolicyCheckCount = results.EvaluationResults.Summary.FilesPassedCount
	}

	evaluationSummary := getEvaluationSummary(policyCheckResults, passedPolicyCheckCount)

	evaluationRequestData := getEvaluationRequestData(token, clientId, clusterK8sVersion, policy.Name, startTime,
		policyCheckResults)

	verifyVersionResponse, err := cliClient.VerifyWebhookVersion(evaluationRequestData.WebhookVersion)
	if err != nil {
		warningMessages = append(warningMessages, err.Error())
	} else {
		if verifyVersionResponse != nil {
			for i := range verifyVersionResponse.MessageTextArray {
				warningMessages = append(warningMessages, verifyVersionResponse.MessageTextArray[i])
			}
		}
	}
	noRecords := os.Getenv(enums.NoRecord)
	if noRecords != "true" {
		_, err = sendEvaluationResult(cliClient, evaluationRequestData)
		if err != nil {
			fmt.Println("saving evaluation results failed")
			warningMessages = append(warningMessages, "saving evaluation results failed")
		}
	}

	resultStr, err := evaluation.GetResultsText(&evaluation.PrintResultsData{
		Results:           results,
		EvaluationSummary: evaluationSummary,
		LoginURL:          prerunData.RegistrationURL,
		Printer:           printer.CreateNewPrinter(),
		K8sVersion:        clusterK8sVersion,
		Verbose:           os.Getenv(enums.Verbose) == "true",
		PolicyName:        policy.Name,
		OutputFormat:      os.Getenv(enums.Output),
	})

	if evaluationSummary.PassedPolicyCheckCount == 0 {
		allowed = false

		sb := strings.Builder{}
		sb.WriteString("\n---\n")
		sb.WriteString(resultStr)
		msg = sb.String()
	}
	return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, allowed, msg, warningMessages)
}

func sendEvaluationResult(cliServiceClient *cliClient.CliClient, evaluationRequestData cliClient.WebhookEvaluationRequestData) (*baseCliClient.SendEvaluationResultsResponse, error) {
	var OSInfoFn = utils.NewOSInfo
	osInfo := OSInfoFn()

	sendEvaluationResultsResponse, err := cliServiceClient.SendWebhookEvaluationResult(&cliClient.EvaluationResultRequest{
		K8sVersion: evaluationRequestData.EvaluationData.K8sVersion,
		ClientId:   evaluationRequestData.EvaluationData.ClientId,
		Token:      evaluationRequestData.EvaluationData.Token,
		PolicyName: evaluationRequestData.EvaluationData.PolicyName,
		Metadata: &cliClient.Metadata{
			Os:              osInfo.OS,
			PlatformVersion: osInfo.PlatformVersion,
			KernelVersion:   osInfo.KernelVersion,
			ClusterContext: &cliClient.ClusterContext{
				IsInCluster:    true,
				WebhookVersion: evaluationRequestData.WebhookVersion,
			},
			EvaluationDurationSeconds: evaluationRequestData.EvaluationData.EvaluationDurationSeconds,
		},
		AllExecutedRules:   evaluationRequestData.EvaluationData.RulesData,
		AllEvaluatedFiles:  evaluationRequestData.EvaluationData.FilesData,
		PolicyCheckResults: evaluationRequestData.EvaluationData.PolicyCheckResults,
	})

	return sendEvaluationResultsResponse, err
}

func ParseEvaluationResponseIntoAdmissionReview(requestUID types.UID, allowed bool, msg string, warningMessages []string) *admission.AdmissionReview {
	statusCode := http.StatusOK
	message := msg

	if !allowed {
		statusCode = http.StatusInternalServerError
		message = msg
	}

	if len(warningMessages) > 0 {
		return &admission.AdmissionReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
			Response: &admission.AdmissionResponse{
				UID:      requestUID,
				Warnings: warningMessages,
				Allowed:  allowed,
				Result: &metav1.Status{
					Code:    int32(statusCode),
					Message: message,
				},
			},
		}
	}

	return &admission.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admission.AdmissionResponse{
			UID:     requestUID,
			Allowed: allowed,
			Result: &metav1.Status{
				Code:    int32(statusCode),
				Message: message,
			},
		},
	}
}

func getClusterK8sVersion() (string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	discClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return "", err
	}

	serverInfo, err := discClient.ServerVersion()
	if err != nil {
		return "", err
	}

	return serverInfo.GitVersion, nil
}

func getK8sVersion() string {
	var err error
	clusterK8sVersion := os.Getenv("CLUSTER_K8S_VERSION")
	if clusterK8sVersion == "" {
		clusterK8sVersion, err = getClusterK8sVersion()
		if err != nil {
			clusterK8sVersion = "unknown k8s version"
		}

		err = os.Setenv("CLUSTER_K8S_VERSION", clusterK8sVersion)
		if err != nil {
			fmt.Println(fmt.Errorf("couldn't set CLUSTER_K8S_VERSION env variable %s", err))
		}
	}
	return clusterK8sVersion
}

func getToken(cliClient *cliClient.CliClient) (string, error) {
	token := os.Getenv(enums.Token)

	if token == "" {
		newToken, err := cliClient.CreateToken()
		if err != nil {
			return "", err
		}

		err = os.Setenv(enums.Token, newToken.Token)
		if err != nil {
			fmt.Println(fmt.Errorf("couldn't set DATREE_TOKEN env variable %s", err))
		}
		token = newToken.Token
	}
	return token, nil
}

func getClientId() string {
	clientId := os.Getenv(enums.ClientId)
	if clientId == "" {
		clientId = shortuuid.New()

		err := os.Setenv(enums.ClientId, clientId)
		if err != nil {
			fmt.Println(fmt.Errorf("couldn't set DATREE_CLIENT_ID env variable %s", err))
		}
	}
	return clientId
}

func getFileConfiguration(admissionReviewReq *admission.AdmissionRequest) []*extractor.FileConfigurations {
	yamlSchema, _ := yaml.JSONToYAML(admissionReviewReq.Object.Raw)

	config := extractor.Configuration{
		MetadataName: admissionReviewReq.Name,
		Kind:         admissionReviewReq.Kind.Kind,
		ApiVersion:   admissionReviewReq.Kind.Version,
		Payload:      yamlSchema,
	}

	var filesConfigurations []*extractor.FileConfigurations
	filesConfigurations = append(filesConfigurations, &extractor.FileConfigurations{
		FileName:       fmt.Sprintf("webhook-%s-%s.tmp.yaml\n\n", admissionReviewReq.Name, admissionReviewReq.Kind.Kind),
		Configurations: []extractor.Configuration{config},
	})

	return filesConfigurations
}

func getEvaluationSummary(policyCheckResults evaluation.PolicyCheckResultData, passedPolicyCheckCount int) printer.EvaluationSummary {
	// the webhook receives one configuration at a time, we know it already passed yaml and k8s validation
	evaluationSummary := printer.EvaluationSummary{
		FilesCount:                1,
		RulesCount:                policyCheckResults.RulesCount,
		PassedYamlValidationCount: 1,
		K8sValidation:             "1/1",
		ConfigsCount:              1,
		PassedPolicyCheckCount:    passedPolicyCheckCount,
	}

	return evaluationSummary
}

func getEvaluationRequestData(token string, clientId string, clusterK8sVersion string, policyName string,
	startTime time.Time, policyCheckResults evaluation.PolicyCheckResultData) cliClient.WebhookEvaluationRequestData {
	endEvaluationTime := time.Now()

	evaluationDurationSeconds := endEvaluationTime.Sub(startTime).Seconds()
	evaluationRequestData := cliClient.WebhookEvaluationRequestData{
		EvaluationData: evaluation.EvaluationRequestData{
			Token:                     token,
			ClientId:                  clientId,
			K8sVersion:                clusterK8sVersion,
			PolicyName:                policyName,
			RulesData:                 policyCheckResults.RulesData,
			FilesData:                 policyCheckResults.FilesData,
			PolicyCheckResults:        policyCheckResults.RawResults,
			EvaluationDurationSeconds: evaluationDurationSeconds,
		},
		WebhookVersion: config.WebhookVersion,
	}

	return evaluationRequestData
}
