package services

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/datreeio/admission-webhook-datree/pkg/server"
	"github.com/stretchr/testify/assert"
	admission "k8s.io/api/admission/v1"
)

// Template taken from here: https://github.com/datreeio/webhook-requests-invetigation/blob/main/kubectl-apply-requests/Deployment-b8a93de7-e0ef-4aaf-8a42-20df3811000e.json
//
//go:embed resourceFilterService_testFixtures/templateResource.json
var templateResource string

func TestConfigMapScanningFiltersValidation(t *testing.T) {
	server.ConfigMapScanningFilters.SkipList = []string{"test-namespace+;CronJob+;test-name+", "namespace;kind;name"}

	t.Run("resource should be skipped because properties match the skip list", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"

		admissionReviewReq.Request.Kind.Kind = "CronJob"
		admissionReviewReq.Request.Namespace = "test-namespace"
		rootObject.Metadata.Name = "test-name"
		assert.Equal(t, true, ShouldResourceBeSkippedByConfigMapScanningFilters(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because properties match the regexes in the skip list", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"

		admissionReviewReq.Request.Kind.Kind = "CronJobbb"
		admissionReviewReq.Request.Namespace = "test-namespaceee"
		rootObject.Metadata.Name = "test-nameee"
		assert.Equal(t, true, ShouldResourceBeSkippedByConfigMapScanningFilters(admissionReviewReq, rootObject))
	})
	t.Run("resource should be validated because kind non-skipped is not in the skip list", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"

		admissionReviewReq.Request.Kind.Kind = "non-skipped"
		admissionReviewReq.Request.Namespace = "test-namespace"
		rootObject.Metadata.Name = "test-name"
		assert.Equal(t, false, ShouldResourceBeSkippedByConfigMapScanningFilters(admissionReviewReq, rootObject))
	})
}

func TestPrerequisitesFilters(t *testing.T) {
	server.ConfigMapScanningFilters.SkipList = []string{}

	t.Run("resource should be skipped because resource is deleted", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		rootObject.Metadata.DeletionTimestamp = "2021-01-01T00:00:00Z"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because metadata name is missing", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		rootObject.Metadata.Name = ""
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because kind is Event", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		admissionReviewReq.Request.Kind.Kind = "Event"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because kind is GitRepository", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		admissionReviewReq.Request.Kind.Kind = "GitRepository"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because namespace is kube-public", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		admissionReviewReq.Request.Namespace = "kube-public"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
	t.Run("resource should be skipped because namespace is kube-node-lease", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		admissionReviewReq.Request.Namespace = "kube-node-lease"
		assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})
}

func TestWhiteListFilters(t *testing.T) {
	server.ConfigMapScanningFilters.SkipList = []string{}

	t.Run("resource should be validated because it is managed by kubectl", func(t *testing.T) {
		t.Run("kubectl-client-side-apply", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "kubectl-client-side-apply"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("kubectl-create", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "kubectl-create"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("kubectl-edit", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "kubectl-edit"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("kubectl-patch", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "kubectl-patch"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
	})

	t.Run("resource should be validated because it is managed by helm", func(t *testing.T) {
		admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
		rootObject.Metadata.ManagedFields[0].Manager = "helm"
		assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
	})

	t.Run("resource should be validated because it is managed by terraform", func(t *testing.T) {
		t.Run("Terraform", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "Terraform"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("HashiCorp", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "HashiCorp"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("some-prefix-terraform-provider-kubernetes", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "some-prefix-terraform-provider-kubernetes"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
	})

	t.Run("resource should be validated because it is managed by openShift", func(t *testing.T) {
		t.Run("openshift-controller-manager-some-postfix", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "openshift-controller-manager-some-postfix"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("openshift-apiserver-some-postfix", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "openshift-apiserver-some-postfix"
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
	})
	t.Run("resource should not be validated because it is not managed by openShift", func(t *testing.T) {
		t.Run("oc", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "oc"
			assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("oc-postfix", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "oc-postfix"
			assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("prefix-openshift-controller-manager-some-postfix", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "prefix-openshift-controller-manager-some-postfix"
			assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("prefix-openshift-apiserver-some-postfix", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "prefix-openshift-apiserver-some-postfix"
			assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("Mozilla should be skipped because we removed it", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "Mozilla"
			assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("Mozilla-postfix", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "Mozilla-postfix"
			assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
		t.Run("mozilla", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "Mozilla-postfix"
			assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
	})

	t.Run("special cases", func(t *testing.T) {
		t.Run("resource should be validated because 1 fields manager is matching", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "non-matching-manager"
			rootObject.Metadata.ManagedFields = append(rootObject.Metadata.ManagedFields, ManagedFields{Manager: "kubectl-client-side-apply"})
			assert.Equal(t, true, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})

		t.Run("resource should be skipped because it is managed by non-allowed manager", func(t *testing.T) {
			admissionReviewReq, rootObject := extractAdmissionReviewReqAndRootObject(templateResource)
			rootObject.Metadata.ManagedFields[0].Manager = "non-matching-manager"
			assert.Equal(t, false, ShouldResourceBeValidated(admissionReviewReq, rootObject))
		})
	})

	// TODO add test cases for Argo and Flux
}

func extractAdmissionReviewReqAndRootObject(resource string) (*admission.AdmissionReview, RootObject) {
	var admissionReviewReq *admission.AdmissionReview
	if err := json.Unmarshal([]byte(resource), &admissionReviewReq); err != nil {
		panic(err)
	}
	rootObject := getResourceRootObject(admissionReviewReq)
	return admissionReviewReq, rootObject
}
