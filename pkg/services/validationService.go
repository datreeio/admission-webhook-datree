package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/k8sMetadataUtil"
	"github.com/datreeio/datree/pkg/deploymentConfig"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"github.com/datreeio/admission-webhook-datree/pkg/server"

	cliDefaultRules "github.com/datreeio/datree/pkg/defaultRules"

	"github.com/datreeio/admission-webhook-datree/pkg/config"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"

	policyFactory "github.com/datreeio/datree/bl/policy"
	"github.com/datreeio/datree/pkg/ciContext"
	baseCliClient "github.com/datreeio/datree/pkg/cliClient"
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
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type ManagedFields struct {
	Manager string `json:"manager"`
}

type Metadata struct {
	Name              string            `json:"name"`
	DeletionTimestamp string            `json:"deletionTimestamp"`
	ManagedFields     []ManagedFields   `json:"managedFields"`
	Labels            map[string]string `json:"labels"`
}

type ValidationService struct {
	CliServiceClient *cliClient.CliClient
}

var cliServiceClient = cliClient.NewCliServiceClient(deploymentConfig.URL, networkValidator.NewNetworkValidator())

func isEnforceMode() bool {
	return os.Getenv(enums.Enforce) == "true"
}

func NewValidationService() *ValidationService {
	return &ValidationService{
		CliServiceClient: cliServiceClient,
	}
}

func NewValidationServiceWithCustomCliServiceClient(cliServiceClient *cliClient.CliClient) *ValidationService {
	return &ValidationService{
		CliServiceClient: cliServiceClient,
	}
}

func (vs *ValidationService) Validate(admissionReviewReq *admission.AdmissionReview, warningMessages *[]string, internalLogger logger.Logger) (admissionReview *admission.AdmissionReview, isSkipped bool) {
	startTime := time.Now()
	msg := "We're good!"
	cliEvaluationId := -1
	var err error

	ciContext := ciContext.Extract()

	clusterK8sVersion := getK8sVersion()
	token, err := getToken(cliServiceClient)
	if err != nil {
		panic(err)
	}

	rootObject := getResourceRootObject(admissionReviewReq)
	namespace, resourceKind, resourceName, managers := getResourceMetadata(admissionReviewReq, rootObject)
	if !ShouldResourceBeValidated(admissionReviewReq, rootObject) {
		clusterRequestMetadata := getClusterRequestMetadata(cliEvaluationId, token, true, true, resourceKind, resourceName, managers, clusterK8sVersion, "", namespace, server.ConfigMapScanningFilters)
		saveRequestMetadataLogInAggregator(clusterRequestMetadata)
		return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, true, msg, *warningMessages), true
	}

	clientId := getClientId()
	policyName := os.Getenv(enums.Policy)

	prerunData, err := vs.CliServiceClient.RequestEvaluationPrerunData(token)
	if err != nil {
		internalLogger.LogError(fmt.Sprintf("Getting prerun data err: %s", err.Error()))
		*warningMessages = append(*warningMessages, err.Error())
	}

	// convert default rules string into DefaultRulesDefinitions structure
	defaultRules, err := cliDefaultRules.YAMLToStruct(prerunData.DefaultRulesYaml)
	if err != nil {
		// get default rules from cli binary on failure
		defaultRules, err = cliDefaultRules.GetDefaultRules()
		// panic if didn't manage to get default rules
		if err != nil {
			panic(err)
		}
	}

	policy, err := policyFactory.CreatePolicy(prerunData.PoliciesJson, policyName, prerunData.RegistrationURL, defaultRules)
	if err != nil {
		*warningMessages = append(*warningMessages, err.Error())
		/*this flow runs when user enter none existing policy name (we wouldn't like to fail the validation for this reason)
		so we are validating against default policy */
		internalLogger.LogError(fmt.Sprintf("Extracting policy out of policies yaml err1: %s", err.Error()))

		for _, policy := range prerunData.PoliciesJson.Policies {
			if policy.IsDefault {
				policyName = policy.Name
			}
		}

		policy, err = policyFactory.CreatePolicy(prerunData.PoliciesJson, policyName, prerunData.RegistrationURL, defaultRules)
		if err != nil {
			internalLogger.LogError(fmt.Sprintf("Extracting policy out of policies yaml err2: %s", err.Error()))
			*warningMessages = append(*warningMessages, err.Error())
			panic(err.Error())
		}
	}

	filesConfigurations := getFileConfiguration(admissionReviewReq.Request)

	policyCheckData := evaluation.PolicyCheckData{
		FilesConfigurations: filesConfigurations,
		IsInteractiveMode:   false,
		PolicyName:          policy.Name,
		Policy:              policy,
	}

	evaluator := evaluation.New(vs.CliServiceClient, ciContext)
	policyCheckResults, err := evaluator.Evaluate(policyCheckData)
	if err != nil {
		internalLogger.LogError(fmt.Sprintf("Evaluate err: %s", err.Error()))
	}

	results := policyCheckResults.FormattedResults
	passedPolicyCheckCount := 0
	if results.EvaluationResults != nil {
		passedPolicyCheckCount = results.EvaluationResults.Summary.FilesPassedCount
	}

	evaluationSummary := getEvaluationSummary(policyCheckResults, passedPolicyCheckCount)

	resource := filesConfigurations[0].Configurations[0]
	kind := resource.Kind
	metadataName := resource.MetadataName
	evaluationRequestData := getEvaluationRequestData(token, clientId, clusterK8sVersion, policy.Name, startTime,
		policyCheckResults, namespace, kind, metadataName)

	verifyVersionResponse, err := cliServiceClient.GetVersionRelatedMessages(evaluationRequestData.WebhookVersion)
	if err != nil {
		*warningMessages = append(*warningMessages, err.Error())
	} else {
		if verifyVersionResponse != nil {
			for i := range verifyVersionResponse.MessageTextArray {
				*warningMessages = append(*warningMessages, verifyVersionResponse.MessageTextArray[i])
			}
		}
	}

	noRecords := os.Getenv(enums.NoRecord)
	if noRecords != "true" {
		evaluationResultResp, err := sendEvaluationResult(cliServiceClient, evaluationRequestData)
		if err == nil {
			cliEvaluationId = evaluationResultResp.EvaluationId
		} else {
			cliEvaluationId = -2
			internalLogger.LogError("saving evaluation results failed")
			*warningMessages = append(*warningMessages, "saving evaluation results failed")
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

	if err != nil {
		internalLogger.LogError(fmt.Sprintf("GetResultsText err: %s", err.Error()))
	}

	var allowed bool
	var isFailedPolicyCheck = evaluationSummary.PassedPolicyCheckCount == 0
	if isFailedPolicyCheck {
		allowed = false

		sb := strings.Builder{}
		sb.WriteString("\n---\n")
		sb.WriteString(resultStr)
		msg = sb.String()
	} else {
		allowed = true
	}

	if !isEnforceMode() {
		allowed = true
		if isFailedPolicyCheck {
			baseUrl := strings.Split(prerunData.RegistrationURL, "datree.io")[0] + "datree.io"
			warningUTMMessage := fmt.Sprintf("🚩 Object with name %s and kind %s failed the policy check, get the full report at: %s/cli/invocations/%d?webhook=true", resourceName, resourceKind, baseUrl, cliEvaluationId)
			*warningMessages = append([]string{warningUTMMessage}, *warningMessages...)
		}
	}

	clusterRequestMetadata := getClusterRequestMetadata(cliEvaluationId, token, false, allowed, resourceKind, resourceName, managers, clusterK8sVersion, policy.Name, namespace, server.ConfigMapScanningFilters)
	saveRequestMetadataLogInAggregator(clusterRequestMetadata)
	return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, allowed, msg, *warningMessages), false
}

type ClusterRequestMetadataAggregator = map[string]*cliClient.ClusterRequestMetadata

var clusterRequestMetadataAggregator = make(ClusterRequestMetadataAggregator)

func saveRequestMetadataLogInAggregator(clusterRequestMetadata *cliClient.ClusterRequestMetadata) {
	logJsonInBytes, err := json.Marshal(clusterRequestMetadata)
	if err != nil {
		logger.LogUtil(err.Error())
		return
	}
	logJson := string(logJsonInBytes)
	currentValue := clusterRequestMetadataAggregator[logJson]
	if currentValue == nil {
		clusterRequestMetadataAggregator[logJson] = clusterRequestMetadata
	} else {
		currentValue.Occurrences++
	}

	if len(clusterRequestMetadataAggregator) >= 500 {
		SendMetadataInBatch()
	}
}

func SendMetadataInBatch() {
	clusterRequestMetadataArray := make([]*cliClient.ClusterRequestMetadata, 0, len(clusterRequestMetadataAggregator))
	for _, value := range clusterRequestMetadataAggregator {
		clusterRequestMetadataArray = append(clusterRequestMetadataArray, value)
	}
	go cliServiceClient.SendRequestMetadataBatch(cliClient.ClusterRequestMetadataBatchReqBody{MetadataLogs: clusterRequestMetadataArray})
	clusterRequestMetadataAggregator = make(ClusterRequestMetadataAggregator) // clear the hash table
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
				IsEnforceMode:  evaluationRequestData.IsEnforceMode,
			},
			EvaluationDurationSeconds: evaluationRequestData.EvaluationData.EvaluationDurationSeconds,
		},
		AllExecutedRules:   evaluationRequestData.EvaluationData.RulesData,
		AllEvaluatedFiles:  evaluationRequestData.EvaluationData.FilesData,
		PolicyCheckResults: evaluationRequestData.EvaluationData.PolicyCheckResults,
		ClusterUuid:        evaluationRequestData.ClusterUuid,
		Namespace:          evaluationRequestData.Namespace,
		Kind:               evaluationRequestData.Kind,
		MetadataName:       evaluationRequestData.MetadataName,
	})

	return sendEvaluationResultsResponse, err
}

func ParseEvaluationResponseIntoAdmissionReview(requestUID k8sTypes.UID, allowed bool, msg string, warningMessages []string) *admission.AdmissionReview {
	statusCode := http.StatusOK
	message := msg

	if !allowed {
		statusCode = http.StatusInternalServerError
		message = msg
	}

	return &admission.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admission.AdmissionResponse{
			UID:      requestUID,
			Allowed:  allowed,
			Warnings: warningMessages,
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
			logger.LogUtil(fmt.Sprintf("couldn't set CLUSTER_K8S_VERSION env variable %s", err))
		}
	}
	return clusterK8sVersion
}

func getToken(cliClient *cliClient.CliClient) (string, error) {
	token := os.Getenv(enums.Token)

	if token == "" {
		logger.LogUtil("No DATREE_TOKEN was found in env")
	}
	return token, nil
}

func getClientId() string {
	clientId := os.Getenv(enums.ClientId)
	if clientId == "" {
		clientId = shortuuid.New()

		err := os.Setenv(enums.ClientId, clientId)
		if err != nil {
			logger.LogUtil(fmt.Sprintf("couldn't set DATREE_CLIENT_ID env variable %s", err))
		}
	}
	return clientId
}

func getFileConfiguration(admissionReviewReq *admission.AdmissionRequest) []*extractor.FileConfigurations {
	yamlSchema, _ := yaml.JSONToYAML(admissionReviewReq.Object.Raw)
	var annotations map[string]interface{}
	var rawYaml map[string]interface{}
	err := yaml.Unmarshal(yamlSchema, &rawYaml)
	if err == nil {
		metadata := rawYaml["metadata"].(map[string]interface{})
		if metadata != nil && metadata["annotations"] != nil {
			annotations = metadata["annotations"].(map[string]interface{})
		}
	}

	config := extractor.Configuration{
		MetadataName: admissionReviewReq.Name,
		Kind:         admissionReviewReq.Kind.Kind,
		ApiVersion:   admissionReviewReq.Kind.Version,
		Payload:      yamlSchema,
		Annotations:  annotations,
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

func getResourceRootObject(admissionReviewReq *admission.AdmissionReview) RootObject {
	var rootObject RootObject
	if err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &rootObject); err != nil {
		panic(err)
	}

	return rootObject
}

func getResourceMetadata(admissionReviewReq *admission.AdmissionReview, rootObject RootObject) (string, string, string, []string) {
	resourceKind := admissionReviewReq.Request.Kind.Kind
	managedFields := rootObject.Metadata.ManagedFields
	namespace := admissionReviewReq.Request.Namespace

	var managers []string
	for _, manager := range managedFields {
		managers = append(managers, manager.Manager)
	}

	resourceName := admissionReviewReq.Request.Name

	return namespace, resourceKind, resourceName, managers
}

func getEvaluationRequestData(token string, clientId string, clusterK8sVersion string, policyName string,
	startTime time.Time, policyCheckResults evaluation.PolicyCheckResultData, evaluationNamespace string, kind string, metadataName string) cliClient.WebhookEvaluationRequestData {
	evaluationDurationSeconds := time.Now().Sub(startTime).Seconds()
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
		ClusterUuid:    k8sMetadataUtil.ClusterUuid,
		IsEnforceMode:  isEnforceMode(),
		Namespace:      evaluationNamespace,
		Kind:           kind,
		MetadataName:   metadataName,
	}

	return evaluationRequestData
}

func getClusterRequestMetadata(cliEvaluationId int, token string, skipped bool, allowed bool, resourceKind string, resourceName string,
	managers []string, clusterK8sVersion string, policyName string, namespace string, configMapScanningFilters server.ConfigMapScanningFiltersType) *cliClient.ClusterRequestMetadata {

	clusterRequestMetadata := &cliClient.ClusterRequestMetadata{
		CliEvaluationId:          cliEvaluationId,
		Token:                    token,
		Skipped:                  skipped,
		Allowed:                  allowed,
		ResourceKind:             resourceKind,
		ResourceName:             resourceName,
		Managers:                 managers,
		PolicyName:               policyName,
		K8sVersion:               clusterK8sVersion,
		Namespace:                namespace,
		ConfigMapScanningFilters: configMapScanningFilters,
		Occurrences:              1,
	}

	return clusterRequestMetadata
}
