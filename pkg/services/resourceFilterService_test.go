package services

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/datreeio/admission-webhook-datree/pkg/server"
	"github.com/stretchr/testify/assert"
	admission "k8s.io/api/admission/v1"
)

type shouldResourceBeValidatedTestCases struct {
	testName           string
	isSkipped          bool
	admissionReviewReq string
}

//go:embed resourceFilterService_testFixtures/metadataNameMissing.json
var metadataNameMissing string

//go:embed resourceFilterService_testFixtures/deletedResource.json
var deletedResource string

//go:embed resourceFilterService_testFixtures/kindEvent.json
var kindEvent string

//go:embed resourceFilterService_testFixtures/kindGitRepository.json
var kindGitRepository string

//go:embed resourceFilterService_testFixtures/deploymentWithVariableFieldManager.json
var deploymentWithVariableFieldManager string

func TestShouldResourceBeValidated(t *testing.T) {
	testCases := []shouldResourceBeValidatedTestCases{
		{
			testName:           "resource should be skipped because metadata name is missing",
			isSkipped:          true,
			admissionReviewReq: metadataNameMissing,
		},
		{
			testName:           "resource should be skipped because it is deleted",
			isSkipped:          true,
			admissionReviewReq: deletedResource,
		},
		{
			testName:           "resource should be skipped because kind is Event",
			isSkipped:          true,
			admissionReviewReq: kindEvent,
		},
		{
			testName:           "resource should be skipped because kind is GitRepository",
			isSkipped:          true,
			admissionReviewReq: kindGitRepository,
		},
	}

	server.ConfigMapScanningFilters.SkipList = []string{}
	for _, testCase := range testCases {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(testCase.admissionReviewReq)

		isResourceSkipped := !ShouldResourceBeValidated(admissionReviewReq, rootObject)
		t.Run(testCase.testName, func(t *testing.T) {
			assert.Equal(t, testCase.isSkipped, isResourceSkipped)
		})
	}
}

func TestConfigMapScanningFiltersValidation(t *testing.T) {
	skipList := []string{"(.*?);Deployment;(.*?)", "namespace;kind;name"}
	server.ConfigMapScanningFilters.SkipList = skipList

	admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)

	shouldResourceBeSkipByScanningFilters := shouldResourceBeSkipByScanningFilters(admissionReviewReq, rootObject)
	t.Run("resource should be skipped because kind Scale is in the skip list", func(t *testing.T) {
		assert.Equal(t, true, shouldResourceBeSkipByScanningFilters)
	})
}

func extractAdmissionReviewReqAndRootObject(resource string) (*admission.AdmissionReview, RootObject) {
	var admissionReviewReq *admission.AdmissionReview
	if err := json.Unmarshal([]byte(resource), &admissionReviewReq); err != nil {
		panic(err)
	}
	rootObject := getResourceRootObject(admissionReviewReq)
	return admissionReviewReq, rootObject
}

func TestFieldManagersFilters(t *testing.T) {
	server.ConfigMapScanningFilters.SkipList = []string{}

	t.Run("resource should be validated because it is managed by kubectl", func(t *testing.T) {
		t.Run("kubectl-client-side-apply", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
			rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("kubectl-create", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
			rootObject.Metadata.ManagedFields[0].Manager = "kubectl-create"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("kubectl-edit", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
			rootObject.Metadata.ManagedFields[0].Manager = "kubectl-edit"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("kubectl-patch", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
			rootObject.Metadata.ManagedFields[0].Manager = "kubectl-patch"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
	})

	t.Run("resource should be validated because it is managed by helm", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
		rootObject.Metadata.ManagedFields[0].Manager = "helm"
		assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})

	t.Run("resource should be validated because it is managed by terraform", func(t *testing.T) {
		t.Run("Terraform", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
			rootObject.Metadata.ManagedFields[0].Manager = "Terraform"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("HashiCorp", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
			rootObject.Metadata.ManagedFields[0].Manager = "HashiCorp"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("some-prefix-terraform-provider-kubernetes", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
			rootObject.Metadata.ManagedFields[0].Manager = "some-prefix-terraform-provider-kubernetes"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
	})

	t.Run("special cases", func(t *testing.T) {
		t.Run("resource should be validated because 1 fields manager is matching", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
			rootObject.Metadata.ManagedFields[0].Manager = "non-matching-manager"
			rootObject.Metadata.ManagedFields = append(rootObject.Metadata.ManagedFields, ManagedFields{Manager: "kubectl-client-side-apply"})
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})

		t.Run("resource should be skipped because it is managed by non-allowed manager", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
			rootObject.Metadata.ManagedFields[0].Manager = "non-matching-manager"
			assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
	})
}
