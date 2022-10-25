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

//go:embed resourceFilterService_testFixtures/notManagedByKubectl.json
var notManagedByKubectl string

//go:embed resourceFilterService_testFixtures/managedByKubectl.json
var managedByKubectl string

//go:embed resourceFilterService_testFixtures/metadataNameMissing.json
var metadataNameMissing string

//go:embed resourceFilterService_testFixtures/deletedResource.json
var deletedResource string

//go:embed resourceFilterService_testFixtures/kindEvent.json
var kindEvent string

//go:embed resourceFilterService_testFixtures/kindGitRepository.json
var kindGitRepository string

//go:embed resourceFilterService_testFixtures/managedByHelm.json
var managedByHelm string

//go:embed resourceFilterService_testFixtures/managedByTerraform.json
var managedByTerraform string

func TestShouldResourceBeValidated(t *testing.T) {
	testCases := []shouldResourceBeValidatedTestCases{
		{
			testName:           "resource should be skipped because it is not managed by kubectl",
			isSkipped:          true,
			admissionReviewReq: notManagedByKubectl,
		},
		{
			testName:           "resource should be validated because it is managed by kubectl",
			isSkipped:          false,
			admissionReviewReq: managedByKubectl,
		},
		{
			testName:           "resource should be validated because it is managed by helm",
			isSkipped:          false,
			admissionReviewReq: managedByHelm,
		},
		{
			testName:           "resource should be validated because it is managed by terraform",
			isSkipped:          false,
			admissionReviewReq: managedByTerraform,
		},
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

	for _, testCase := range testCases {
		var admissionReviewReq *admission.AdmissionReview
		if err := json.Unmarshal([]byte(testCase.admissionReviewReq), &admissionReviewReq); err != nil {
			panic(err)
		}

		rootObject := getResourceRootObject(admissionReviewReq)

		isResourceSkipped := !ShouldResourceBeValidated(admissionReviewReq, rootObject)
		t.Run(testCase.testName, func(t *testing.T) {
			assert.Equal(t, testCase.isSkipped, isResourceSkipped)
		})
	}
}

func TestConfigmapScanningFiltersValidation(t *testing.T) {
	skipList := []string{"(.*?);Scale;(.*?)", "namespace;kind;name"}
	server.ConfigmapScanningFilters.SkipList = skipList

	var admissionReviewReq *admission.AdmissionReview
	if err := json.Unmarshal([]byte(managedByKubectl), &admissionReviewReq); err != nil {
		panic(err)
	}

	rootObject := getResourceRootObject(admissionReviewReq)

	shouldResourceBeSkipByScanningFilters := shouldResourceBeSkipByScanningFilters(admissionReviewReq, rootObject)
	t.Run("resource should be skipped because kind Scale it is in the skip list", func(t *testing.T) {
		assert.Equal(t, true, shouldResourceBeSkipByScanningFilters)
	})

}
