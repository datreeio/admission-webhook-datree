package services

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/datreeio/admission-webhook-datree/pkg/server"
	"github.com/stretchr/testify/assert"
	admission "k8s.io/api/admission/v1"
)

//go:embed resourceFilterService_testFixtures/deploymentWithVariableFieldManager.json
var deploymentWithVariableFieldManager string

func TestPrerequisitesFilters(t *testing.T) {
	server.ConfigMapScanningFilters.SkipList = []string{}

	t.Run("resource should be skipped because resource is deleted", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"
		rootObject.Metadata.DeletionTimestamp = "2021-01-01T00:00:00Z"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because metadata name is missing", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"
		rootObject.Metadata.Name = ""
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because kind is Event", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"
		admissionReviewReq.Request.Kind.Kind = "Event"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because kind is GitRepository", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"
		admissionReviewReq.Request.Kind.Kind = "GitRepository"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
}

func TestConfigMapScanningFiltersValidation(t *testing.T) {
	server.ConfigMapScanningFilters.SkipList = []string{"(.*?);Deployment+;(.*?)", "namespace;kind;name"}

	t.Run("resource should be skipped because kind Deployment is in the skip list", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"
		admissionReviewReq.Request.Kind.Kind = "Deployment"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because kind Deploymenttt matches the regex in the skip list", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"
		admissionReviewReq.Request.Kind.Kind = "Deploymenttt"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be validated because kind non-skipped-kind is not in the skip list", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(deploymentWithVariableFieldManager)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"
		admissionReviewReq.Request.Kind.Kind = "non-skipped-kind"
		assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
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

func extractAdmissionReviewReqAndRootObject(resource string) (*admission.AdmissionReview, RootObject) {
	var admissionReviewReq *admission.AdmissionReview
	if err := json.Unmarshal([]byte(resource), &admissionReviewReq); err != nil {
		panic(err)
	}
	rootObject := getResourceRootObject(admissionReviewReq)
	return admissionReviewReq, rootObject
}
