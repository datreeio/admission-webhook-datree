package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	servicestate "github.com/datreeio/admission-webhook-datree/pkg/serviceState"

	"github.com/datreeio/admission-webhook-datree/pkg/k8sMetadataUtil"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"github.com/datreeio/admission-webhook-datree/pkg/server"

	cliDefaultRules "github.com/datreeio/datree/pkg/defaultRules"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"

	policyFactory "github.com/datreeio/datree/bl/policy"
	"github.com/datreeio/datree/pkg/ciContext"
	baseCliClient "github.com/datreeio/datree/pkg/cliClient"
	"github.com/datreeio/datree/pkg/evaluation"
	"github.com/datreeio/datree/pkg/extractor"
	"github.com/datreeio/datree/pkg/printer"
	"github.com/datreeio/datree/pkg/utils"

	cliClient "github.com/datreeio/admission-webhook-datree/pkg/clients"

	"github.com/ghodss/yaml"
	admission "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
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
	K8sMetadataUtil  *k8sMetadataUtil.K8sMetadataUtil
	ErrorReporter    *errorReporter.ErrorReporter
	State            *servicestate.ServiceState
}

func (vs *ValidationService) Validate(admissionReviewReq *admission.AdmissionReview, warningMessages *[]string, internalLogger logger.Logger) (admissionReview *admission.AdmissionReview, isSkipped bool) {
	startTime := time.Now()
	msg := "We're good!"
	cliEvaluationId := -1
	var err error

	ciContext := ciContext.Extract()

	clusterK8sVersion := vs.State.GetK8sVersion()
	token := vs.State.GetToken()
	if token == "" {
		errorMessage := "no DATREE_TOKEN was found in env"
		vs.ErrorReporter.ReportUnexpectedError(errors.New(errorMessage))
		logger.LogUtil(errorMessage)
	}

	rootObject := getResourceRootObject(admissionReviewReq)
	namespace, resourceKind, resourceName, managers := getResourceMetadata(admissionReviewReq, rootObject)

	saveMetadataAndReturnAResponseForSkippedResource := func() (admissionReview *admission.AdmissionReview, isSkipped bool) {
		clusterRequestMetadata := getClusterRequestMetadata(cliEvaluationId, token, true, true, resourceKind, resourceName, managers, clusterK8sVersion, "", namespace, server.ConfigMapScanningFilters)
		vs.saveRequestMetadataLogInAggregator(clusterRequestMetadata)
		return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, true, msg, *warningMessages), true
	}

	if !ShouldResourceBeValidated(admissionReviewReq, rootObject) {
		return saveMetadataAndReturnAResponseForSkippedResource()
	}

	prerunData, err := vs.CliServiceClient.RequestClusterEvaluationPrerunData(token, vs.State.GetClusterUuid())
	if err != nil {
		internalLogger.LogAndReportUnexpectedError(fmt.Sprintf("Getting prerun data err: %s", err.Error()))

		prerunWarningMsg := "Datree failed to run policy check - an error occurred when pulling your policy"
		*warningMessages = append(*warningMessages, prerunWarningMsg)
		return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, true, msg, *warningMessages), true
	}
	if !vs.State.GetConfigFromHelm() {
		vs.State.SetIsEnforceMode(prerunData.ActionOnFailure == enums.EnforceActionOnFailure)
		server.OverrideSkipList(prerunData.IgnorePatterns)
	}

	if ShouldResourceBeSkippedByConfigMapScanningFilters(admissionReviewReq, rootObject) {
		return saveMetadataAndReturnAResponseForSkippedResource()
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

	filesConfigurations := getFileConfiguration(admissionReviewReq.Request)

	evaluator := evaluation.New(vs.CliServiceClient, ciContext)

	didFailAtLeastOnePolicyCheck := false

	for _, policyName := range prerunData.ActivePolicies {
		if !vs.shouldPolicyRunForNamespace(policyName, namespace) {
			continue
		}

		// create policy
		policy, err := policyFactory.CreatePolicy(prerunData.PoliciesJson, policyName, prerunData.RegistrationURL, defaultRules, false)
		if err != nil {
			*warningMessages = append(*warningMessages, fmt.Sprintf("Policy %s not found, skipping evaluation", policyName))
			continue
		}

		policyCheckData := evaluation.PolicyCheckData{
			FilesConfigurations: filesConfigurations,
			IsInteractiveMode:   false,
			PolicyName:          policy.Name,
			Policy:              policy,
		}

		// evaluate policy against configuration
		policyCheckResults, err := evaluator.Evaluate(policyCheckData)
		if err != nil {
			internalLogger.LogAndReportUnexpectedError(fmt.Sprintf("Evaluate err: %s", err.Error()))
		}

		results := policyCheckResults.FormattedResults
		passedPolicyCheckCount := 0
		if results.EvaluationResults != nil {
			passedPolicyCheckCount = results.EvaluationResults.Summary.FilesPassedCount
		}

		evaluationSummary := getEvaluationSummary(policyCheckResults, passedPolicyCheckCount)

		// send results to backend
		noRecords := os.Getenv(enums.NoRecord)
		if noRecords != "true" {
			evaluationResultResp, err := vs.sendEvaluationResult(vs.getEvaluationRequestData(policy.Name, startTime,
				policyCheckResults, namespace, resourceKind, resourceName))
			if err == nil {
				cliEvaluationId = evaluationResultResp.EvaluationId
			} else {
				cliEvaluationId = -2
				internalLogger.LogAndReportUnexpectedError("saving evaluation results failed")
				*warningMessages = append(*warningMessages, "saving evaluation results failed")
			}
		}

		// get results text
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
			internalLogger.LogAndReportUnexpectedError(fmt.Sprintf("GetResultsText err: %s", err.Error()))
		}

		didFailCurrentPolicyCheck := evaluationSummary.PassedPolicyCheckCount == 0
		if didFailCurrentPolicyCheck {
			didFailAtLeastOnePolicyCheck = true

			sb := strings.Builder{}
			sb.WriteString("\n---\n")
			sb.WriteString(resultStr)
			msg = sb.String()
		}
	}

	allowed := !didFailAtLeastOnePolicyCheck

	if !vs.State.GetIsEnforceMode() {
		allowed = true
		baseUrl := strings.Split(prerunData.RegistrationURL, "datree.io")[0] + "datree.io"
		invocationUrl := fmt.Sprintf("%s/cli/invocations/%d?webhook=true", baseUrl, cliEvaluationId)
		if didFailAtLeastOnePolicyCheck {
			*warningMessages = append([]string{
				fmt.Sprintf("🚩 Object with name \"%s\" and kind \"%s\" failed the policy check", resourceName, resourceKind),
				fmt.Sprintf("👉 Get the full report %s", invocationUrl),
			}, *warningMessages...)
		} else {
			*warningMessages = append([]string{
				fmt.Sprintf("✅ Object with name \"%s\" and kind \"%s\" passed Datree's policy check", resourceName, resourceKind),
				fmt.Sprintf("👉 Get the full report %s", invocationUrl),
			}, *warningMessages...)
		}
	}

	verifyVersionResponse, err := vs.CliServiceClient.GetVersionRelatedMessages(vs.State.GetServiceVersion())
	if err != nil {
		*warningMessages = append(*warningMessages, err.Error())
	} else {
		if verifyVersionResponse != nil {
			//nolint:all
			for i := range verifyVersionResponse.MessageTextArray {
				*warningMessages = append(*warningMessages, verifyVersionResponse.MessageTextArray[i])
			}
		}
	}

	clusterRequestMetadata := getClusterRequestMetadata(cliEvaluationId, token, false, allowed, resourceKind, resourceName, managers, clusterK8sVersion, "PLACEHOLDER_POLICY_NAME", namespace, server.ConfigMapScanningFilters) // TODO use the real policy name
	vs.saveRequestMetadataLogInAggregator(clusterRequestMetadata)
	return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, allowed, msg, *warningMessages), false
}

func (m *clusterRequestMetadataMap) Len() int {
	return m.entriesCount
}
func (m *clusterRequestMetadataMap) ShouldSendBatchToServer() bool {
	return m.entriesCount >= 500
}

func (m *clusterRequestMetadataMap) LoadOrStore(logJson string, clusterRequestMetadata *cliClient.ClusterRequestMetadata) {
	existingLog, loaded := m.clusterRequestMetadataAggregator.LoadOrStore(logJson, clusterRequestMetadata)
	if loaded {
		existingLog.(*cliClient.ClusterRequestMetadata).Occurrences++
	} else {
		m.entriesCount += 1
	}
}

func (m *clusterRequestMetadataMap) Clear() {
	m.clusterRequestMetadataAggregator = &sync.Map{}
	m.entriesCount = 0
}

type clusterRequestMetadataMap struct {
	entriesCount                     int
	clusterRequestMetadataAggregator *sync.Map
}

func clusterRequestMetadataMapNew() *clusterRequestMetadataMap {
	return &clusterRequestMetadataMap{
		entriesCount:                     0,
		clusterRequestMetadataAggregator: &sync.Map{},
	}
}

var clusterRequestMetadataAggregatorMap = clusterRequestMetadataMapNew()

func (vs *ValidationService) saveRequestMetadataLogInAggregator(clusterRequestMetadata *cliClient.ClusterRequestMetadata) {
	logJsonInBytes, err := json.Marshal(clusterRequestMetadata)
	if err != nil {
		vs.ErrorReporter.ReportUnexpectedError(err)
		logger.LogUtil(err.Error())
		return
	}
	logJson := string(logJsonInBytes)
	clusterRequestMetadataAggregatorMap.LoadOrStore(logJson, clusterRequestMetadata)

	if clusterRequestMetadataAggregatorMap.ShouldSendBatchToServer() {
		vs.SendMetadataInBatch()
	}
}

func (vs *ValidationService) SendMetadataInBatch() {
	clusterRequestMetadataArray := make([]*cliClient.ClusterRequestMetadata, 0, clusterRequestMetadataAggregatorMap.entriesCount)
	clusterRequestMetadataAggregatorMap.clusterRequestMetadataAggregator.Range(func(key, value interface{}) bool {
		clusterRequestMetadataArray = append(clusterRequestMetadataArray, value.(*cliClient.ClusterRequestMetadata))
		return true
	})

	go vs.CliServiceClient.SendRequestMetadataBatch(cliClient.ClusterRequestMetadataBatchReqBody{MetadataLogs: clusterRequestMetadataArray})

	clusterRequestMetadataAggregatorMap.Clear()
}

func (vs *ValidationService) sendEvaluationResult(evaluationRequestData cliClient.WebhookEvaluationRequestData) (*baseCliClient.SendEvaluationResultsResponse, error) {
	var OSInfoFn = utils.NewOSInfo
	osInfo := OSInfoFn()

	sendEvaluationResultsResponse, err := vs.CliServiceClient.SendWebhookEvaluationResult(&cliClient.EvaluationResultRequest{
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

func (vs *ValidationService) shouldPolicyRunForNamespace(policyName string, namespace string) bool {
	namespaceRestrictions := vs.getNamespaceRestrictionsByPolicyName(policyName)
	if namespaceRestrictions == nil {
		return true
	}
	for _, excludePattern := range namespaceRestrictions.ExcludePatterns {
		if match, _ := regexp.MatchString(excludePattern, namespace); match {
			return false
		}
	}
	for _, includePattern := range namespaceRestrictions.IncludePatterns {
		if match, _ := regexp.MatchString(includePattern, namespace); match {
			return true
		}
	}
	return false
}
func (vs *ValidationService) getNamespaceRestrictionsByPolicyName(policyName string) *servicestate.Namespaces {
	policies := vs.State.GetMultiplePolicies()
	if policies == nil {
		return nil
	}

	for _, policy := range *policies {
		if policy.Policy == policyName {
			return &policy.Namespaces
		}
	}
	return nil
}

func getFileConfiguration(admissionReviewReq *admission.AdmissionRequest) []*extractor.FileConfigurations {
	yamlSchema, _ := yaml.JSONToYAML(admissionReviewReq.Object.Raw)
	configs, _ := extractor.ParseYaml(string(yamlSchema))

	var filesConfigurations []*extractor.FileConfigurations
	filesConfigurations = append(filesConfigurations, &extractor.FileConfigurations{
		FileName:       fmt.Sprintf("webhook-%s-%s.tmp.yaml\n\n", admissionReviewReq.Name, admissionReviewReq.Kind.Kind),
		Configurations: *configs,
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

func (vs *ValidationService) getEvaluationRequestData(policyName string,
	startTime time.Time, policyCheckResults evaluation.PolicyCheckResultData, evaluationNamespace string, kind string, metadataName string) cliClient.WebhookEvaluationRequestData {
	//nolint:all
	evaluationDurationSeconds := time.Now().Sub(startTime).Seconds()
	evaluationRequestData := cliClient.WebhookEvaluationRequestData{
		EvaluationData: evaluation.EvaluationRequestData{
			Token:                     vs.State.GetToken(),
			ClientId:                  vs.State.GetClientId(),
			K8sVersion:                vs.State.GetK8sVersion(),
			PolicyName:                policyName,
			RulesData:                 policyCheckResults.RulesData,
			FilesData:                 policyCheckResults.FilesData,
			PolicyCheckResults:        policyCheckResults.RawResults,
			EvaluationDurationSeconds: evaluationDurationSeconds,
		},
		WebhookVersion: vs.State.GetServiceVersion(),
		ClusterUuid:    vs.State.GetClusterUuid(),
		IsEnforceMode:  vs.State.GetIsEnforceMode(),
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
