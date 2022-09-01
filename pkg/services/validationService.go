package services

import (
	"encoding/json"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
	"net/http"
	"os"
	"strings"
	"time"

	cliDefaultRules "github.com/datreeio/datree/pkg/defaultRules"

	"k8s.io/utils/strings/slices"

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

type RootObject struct {
	Metadata Metadata `json:"metadata"`
}

func Validate(admissionReviewReq *admission.AdmissionReview, warningMessages *[]string) *admission.AdmissionReview {
	startTime := time.Now()
	msg := "We're good!"
	allowed := true
	var err error

	validator := networkValidator.NewNetworkValidator()
	cliClient := cliClient.NewCliServiceClient(deploymentConfig.URL, validator)
	ciContext := ciContext.Extract()

	loggerUtil.Log("Getting k8s version")
	clusterK8sVersion := getK8sVersion()
	loggerUtil.Log(fmt.Sprintf("k8s version: %s", clusterK8sVersion))

	var rootObject RootObject
	if err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &rootObject); err != nil {
		fmt.Println(err)
		panic(err)
	}

	resourceKind := admissionReviewReq.Request.Kind.Kind
	loggerUtil.Log(fmt.Sprintf("resource kind: %s", resourceKind))

	loggerUtil.Log("Starting filtering process")
	if !isMetadataNameExists(rootObject.Metadata) ||
		!shouldEvaluateResourceByKind(resourceKind) ||
		!shouldEvaluateArgoCRDResources(resourceKind, admissionReviewReq.Request.Operation) ||
		!shouldEvaluateFluxCDResources(*admissionReviewReq.Request.DryRun, rootObject.Metadata.Labels, admissionReviewReq.Request.Namespace) ||
		!shouldEvaluateResourceByManager(rootObject.Metadata.ManagedFields) ||
		rootObject.Metadata.DeletionTimestamp != "" {

		loggerUtil.Log("Resource needs to be skipped")
		return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, allowed, msg, *warningMessages)
	}

	loggerUtil.Log("Resource needs to be scan")
	token, err := getToken(cliClient)
	if err != nil {
		panic(err)
	}

	clientId := getClientId()

	policyName := os.Getenv(enums.Policy)
	loggerUtil.Log(fmt.Sprintf("policyName: %s", policyName))

	loggerUtil.Log("Getting prerun data")
	prerunData, err := cliClient.RequestEvaluationPrerunData(token)
	if err != nil {
		loggerUtil.Log(fmt.Sprintf("Getting prerun data err: %s", err.Error()))
		*warningMessages = append(*warningMessages, err.Error())
	}

	loggerUtil.Log("convert default rules string into DefaultRulesDefinitions structure")
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

	loggerUtil.Log("Extracting policy out of policies yaml")
	policy, err := policyFactory.CreatePolicy(prerunData.PoliciesJson, policyName, prerunData.RegistrationURL, defaultRules)
	if err != nil {
		*warningMessages = append(*warningMessages, err.Error())
		/*this flow runs when user enter none existing policy name (we wouldn't like to fail the validation for this reason)
		so we are validating against default policy */
		loggerUtil.Log(fmt.Sprintf("Extracting policy out of policies yaml err1: %s", err.Error()))

		for _, policy := range prerunData.PoliciesJson.Policies {
			if policy.IsDefault {
				policyName = policy.Name
			}
		}

		policy, err = policyFactory.CreatePolicy(prerunData.PoliciesJson, policyName, prerunData.RegistrationURL, defaultRules)
		if err != nil {
			loggerUtil.Log(fmt.Sprintf("Extracting policy out of policies yaml err2: %s", err.Error()))
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

	evaluator := evaluation.New(cliClient, ciContext)
	policyCheckResults, err := evaluator.Evaluate(policyCheckData)
	if err != nil {
		loggerUtil.Log(fmt.Sprintf("Evaluate err: %s", err.Error()))
	}

	results := policyCheckResults.FormattedResults
	passedPolicyCheckCount := 0
	if results.EvaluationResults != nil {
		passedPolicyCheckCount = results.EvaluationResults.Summary.FilesPassedCount
	}

	evaluationSummary := getEvaluationSummary(policyCheckResults, passedPolicyCheckCount)

	evaluationRequestData := getEvaluationRequestData(token, clientId, clusterK8sVersion, policy.Name, startTime,
		policyCheckResults)

	verifyVersionResponse, err := cliClient.GetVersionRelatedMessages(evaluationRequestData.WebhookVersion)
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
		_, err = sendEvaluationResult(cliClient, evaluationRequestData)
		if err != nil {
			fmt.Println("saving evaluation results failed")
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
		loggerUtil.Log(fmt.Sprintf("GetResultsText err: %s", err.Error()))
	}

	if evaluationSummary.PassedPolicyCheckCount == 0 {
		allowed = false

		sb := strings.Builder{}
		sb.WriteString("\n---\n")
		sb.WriteString(resultStr)
		msg = sb.String()
	}

	return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, allowed, msg, *warningMessages)
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

func ParseEvaluationResponseIntoAdmissionReview(requestUID k8sTypes.UID, allowed bool, msg string, warningMessages []string) *admission.AdmissionReview {
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
	loggerUtil.Log("getToken")
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
	loggerUtil.Log("getClientId")
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
	loggerUtil.Log("getFileConfiguration")
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
	loggerUtil.Log("getEvaluationSummary")
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
	loggerUtil.Log("getEvaluationRequestData")
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

func shouldEvaluateResourceByKind(resourceKind string) bool {
	loggerUtil.Log("Filtering - shouldEvaluateResourceByKind")
	unsupportedResourceKinds := []string{"Event", "GitRepository"}
	return !slices.Contains(unsupportedResourceKinds, resourceKind)
}

func shouldEvaluateArgoCRDResources(resourceKind string, operation admission.Operation) bool {
	loggerUtil.Log("Filtering - shouldEvaluateArgoCRDResources")
	argoCRDList := []string{"Application", "Workflow", "Rollout"}
	isKindInArgoCRDList := slices.Contains(argoCRDList, resourceKind)
	return (isKindInArgoCRDList && operation == admission.Create) || !isKindInArgoCRDList
}

func shouldEvaluateFluxCDResources(isDryRun bool, labels map[string]string, namespace string) bool {
	loggerUtil.Log("Filtering - shouldEvaluateFluxCDResources")
	isFluxObject := isFluxObject(labels, namespace)
	badFluxObject := (isFluxObject && len(labels) == 0) || (isFluxObject && isDryRun)

	return !isFluxObject || !badFluxObject
}

func isFluxObject(labels map[string]string, namespace string) bool {
	loggerUtil.Log("Filtering - isFluxObject")
	if namespace == "flux-system" {
		return true
	}

	for label := range labels {
		if strings.Contains(label, "kustomize.toolkit.fluxcd.io") {
			return true
		}
	}

	return false
}

func shouldEvaluateResourceByManager(fields []ManagedFields) bool {
	loggerUtil.Log("Filtering - shouldEvaluateResourceByManager")
	supportedPrefixes := []string{"kubectl", "argocd", "argo", "kustomize-controller"}
	for _, field := range fields {
		for _, prefix := range supportedPrefixes {
			if strings.HasPrefix(field.Manager, prefix) {
				return true
			}
		}
	}
	return false
}

func isMetadataNameExists(metadata Metadata) bool {
	loggerUtil.Log("Filtering - isMetadataNameExists")
	return metadata.Name != ""
}
