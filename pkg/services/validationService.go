package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/openshiftClient"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"

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
	Name              string                     `json:"name"`
	DeletionTimestamp string                     `json:"deletionTimestamp"`
	ManagedFields     []ManagedFields            `json:"managedFields"`
	Labels            map[string]string          `json:"labels"`
	OwnerReferences   []cliClient.OwnerReference `json:"ownerReferences"`
	Annotations       map[string]string          `json:"annotations"`
}

type ValidationService struct {
	CliServiceClient *cliClient.CliClient
	K8sMetadataUtil  *k8sMetadataUtil.K8sMetadataUtil
	ErrorReporter    *errorReporter.ErrorReporter
	State            *servicestate.ServiceState
	OpenshiftClient  *openshiftClient.OpenshiftClient
	Logger           *logger.Logger
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
	resourceUserInfo := admissionReviewReq.Request.UserInfo

	saveMetadataAndReturnAResponseForSkippedResource := func() (admissionReview *admission.AdmissionReview, isSkipped bool) {
		clusterRequestMetadata := getClusterRequestMetadata(vs.State.GetClusterUuid(), vs.State.GetServiceVersion(), cliEvaluationId, token, true, true, resourceKind, resourceName, managers, clusterK8sVersion, "", namespace, server.ConfigMapScanningFilters, rootObject.Metadata.OwnerReferences)
		vs.saveRequestMetadataLogInAggregator(clusterRequestMetadata)
		*warningMessages = append([]string{
			fmt.Sprintf("⏩ Object with name \"%s\" was skipped by Datree's policy check.", resourceName),
			"👉 To avoid skipping this resource, contact support using the live chat: https://app.datree.io/",
		}, *warningMessages...)
		return ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, true, msg, *warningMessages), true
	}

	shouldValidatedResourceData := ShouldResourceBeValidated(admissionReviewReq, rootObject)

	if !shouldValidatedResourceData.ShouldValidate {
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
		vs.State.SetBypassPermissions(&prerunData.BypassPermissions)
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

	allowed := true

	sb := strings.Builder{}

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
		shouldBypassByPermissions := vs.shouldBypassByPermissions(resourceUserInfo, shouldValidatedResourceData.OpenShiftRequester)

		if didFailCurrentPolicyCheck && vs.State.GetIsEnforceMode() && !shouldBypassByPermissions {
			allowed = false

			sb.WriteString("\n---\n")
			sb.WriteString(resultStr)
		}

		if shouldBypassByPermissions && didFailCurrentPolicyCheck {
			*warningMessages = append([]string{
				"🚩 Your resource failed the policy check, but it has been applied due to your bypass privileges",
			}, *warningMessages...)
		} else if !vs.State.GetIsEnforceMode() {
			baseUrl := strings.Split(prerunData.RegistrationURL, "datree.io")[0] + "datree.io"
			invocationUrl := fmt.Sprintf("%s/cli/invocations/%d?webhook=true", baseUrl, cliEvaluationId)
			if didFailCurrentPolicyCheck {
				*warningMessages = append([]string{
					fmt.Sprintf("🚩 Object with name \"%s\" and kind \"%s\" failed the policy check for policy \"%s\"", resourceName, resourceKind, policyName),
					fmt.Sprintf("👉 Get the full report %s", invocationUrl),
				}, *warningMessages...)
			} else {
				*warningMessages = append([]string{
					fmt.Sprintf("✅  Object with name \"%s\" and kind \"%s\" passed Datree's policy check for policy \"%s\"", resourceName, resourceKind, policyName),
					fmt.Sprintf("👉 Get the full report %s", invocationUrl),
				}, *warningMessages...)
			}
		}
	}

	msg = sb.String()

	verifyVersionResponse, err := vs.CliServiceClient.GetVersionRelatedMessages(vs.State.GetServiceVersion())
	if err != nil {
		*warningMessages = append(*warningMessages, err.Error())
	} else {
		if verifyVersionResponse != nil {
			*warningMessages = append(*warningMessages, verifyVersionResponse.MessageTextArray...)
		}
	}

	clusterRequestMetadata := getClusterRequestMetadata(vs.State.GetClusterUuid(), vs.State.GetServiceVersion(), cliEvaluationId, token, false, allowed, resourceKind, resourceName, managers, clusterK8sVersion, vs.State.GetPolicyName(), namespace, server.ConfigMapScanningFilters, rootObject.Metadata.OwnerReferences)
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

func (vs *ValidationService) shouldBypassByPermissions(userInfo authenticationv1.UserInfo, openShiftRequester string) bool {
	bypassPermissions := vs.State.GetBypassPermissions()

	if bypassPermissions == nil {
		return false
	}

	userName := userInfo.Username
	groups := userInfo.Groups
	if openShiftRequester != "" {
		// override username
		userName = openShiftRequester

		// override groups
		groupsFromOpenshiftClient, err := vs.OpenshiftClient.GetGroupsUserBelongsTo(openShiftRequester)
		if err != nil {
			vs.Logger.LogError(fmt.Sprintf("Failed to get groups for user %s from openshift client: %s", openShiftRequester, err.Error()))
		} else {
			groups = groupsFromOpenshiftClient
		}
	}

	for _, userAccount := range bypassPermissions.UserAccounts {
		if match, _ := regexp.MatchString(userAccount, userName); match {
			return true
		}
	}

	for _, serviceAccount := range bypassPermissions.ServiceAccounts {
		if match, _ := regexp.MatchString(serviceAccount, userName); match {
			return true
		}
	}

	for _, bypassGroup := range bypassPermissions.Groups {
		for _, userInfoGroup := range groups {
			if match, _ := regexp.MatchString(bypassGroup, userInfoGroup); match {
				return true
			}
		}
	}
	return false
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

	evaluationDurationSeconds := time.Since(startTime).Seconds()
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

func getClusterRequestMetadata(clusterUuid k8sTypes.UID, webhookVersion string, cliEvaluationId int, token string, skipped bool, allowed bool, resourceKind string, resourceName string,
	managers []string, clusterK8sVersion string, policyName string, namespace string, configMapScanningFilters server.ConfigMapScanningFiltersType, ownerReferences []cliClient.OwnerReference) *cliClient.ClusterRequestMetadata {

	clusterRequestMetadata := &cliClient.ClusterRequestMetadata{
		ClusterUuid:              clusterUuid,
		WebhookVersion:           webhookVersion,
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
		OwnerReferences:          ownerReferences,
		Occurrences:              1,
	}

	return clusterRequestMetadata
}
