package services

import (
	"encoding/json"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/loggerUtil"
	admission "k8s.io/api/admission/v1"
	"k8s.io/utils/strings/slices"
	"strings"
)

type RootObject struct {
	Metadata Metadata `json:"metadata"`
}

func ShouldResourceBeValidated(admissionReviewReq *admission.AdmissionReview) bool {
	if admissionReviewReq == nil {
		panic("admissionReviewReq is nil")
	}

	var rootObject RootObject
	if err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &rootObject); err != nil {
		fmt.Println(err)
		panic(err)
	}

	resourceKind := admissionReviewReq.Request.Kind.Kind
	loggerUtil.Log(fmt.Sprintf("resource kind: %s", resourceKind))

	loggerUtil.Log("Starting filtering process")

	// assigning to variables for easier debugging
	isMetadataNameExists := isMetadataNameExists(rootObject.Metadata)
	isUnsupportedKind := isUnsupportedKind(resourceKind)
	isResourceDeleted := isResourceDeleted(rootObject)
	arePrerequisitesMet := isMetadataNameExists && !isUnsupportedKind && !isResourceDeleted

	if !arePrerequisitesMet {
		return false
	}

	isKubectl := isKubectl(rootObject.Metadata.ManagedFields)
	isFluxResourceThatShouldBeEvaluated := isFluxResourceThatShouldBeEvaluated(*admissionReviewReq.Request.DryRun, rootObject.Metadata.Labels, admissionReviewReq.Request.Namespace, rootObject.Metadata.ManagedFields)
	isArgoResourceThatShouldBeEvaluated := isArgoResourceThatShouldBeEvaluated(resourceKind, admissionReviewReq.Request.Operation, rootObject.Metadata.ManagedFields)
	isResourceWhiteListed := isKubectl || isFluxResourceThatShouldBeEvaluated || isArgoResourceThatShouldBeEvaluated

	if !isResourceWhiteListed {
		return false
	}

	return true
}

func isMetadataNameExists(metadata Metadata) bool {
	loggerUtil.Log("Filtering - isMetadataNameExists")
	return metadata.Name != ""
}

func isUnsupportedKind(resourceKind string) bool {
	loggerUtil.Log("Filtering - isUnsupportedKind")
	unsupportedResourceKinds := []string{"Event", "GitRepository"}
	return slices.Contains(unsupportedResourceKinds, resourceKind)
}

func isResourceDeleted(rootObject RootObject) bool {
	return rootObject.Metadata.DeletionTimestamp != ""
}

func isKubectl(fields []ManagedFields) bool {
	loggerUtil.Log("Filtering - isKubectl")

	/*
		This is a strict check for only those field managers to make sure the request was sent via kubectl.
		all values were taken from these pages under the default value of the flag "field-manager"
		https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply
		https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#create
		https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#edit
		https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#patch

		if the user overrides the default value of the flag "field-manager" then the request will not be considered a kubectl request
		and therefore will likely not be evaluated
	*/
	return isAtLeastOneFieldManagerEqualToOneOfThePrefixes(fields, []string{"kubectl-client-side-apply", "kubectl-create", "kubectl-edit", "kubectl-patch"})
}

func isFluxResourceThatShouldBeEvaluated(isDryRun bool, labels map[string]string, namespace string, fields []ManagedFields) bool {
	loggerUtil.Log("Filtering - isFluxResourceThatShouldBeEvaluated")

	if !doesAtLeastOneFieldManagerStartWithOneOfThePrefixes(fields, []string{"kustomize-controller"}) {
		return false
	}

	if !isFluxObject(labels, namespace) {
		return false
	}

	badFluxObject := (len(labels) == 0) || (isDryRun)
	if badFluxObject {
		return false
	}

	return true
}

func isArgoResourceThatShouldBeEvaluated(resourceKind string, operation admission.Operation, fields []ManagedFields) bool {
	loggerUtil.Log("Filtering - isArgoResourceThatShouldBeEvaluated")

	if !doesAtLeastOneFieldManagerStartWithOneOfThePrefixes(fields, []string{"argocd", "argo"}) {
		return false
	}

	isKindInArgoCRDListThatShouldBeValidatedOnlyOnCreate := slices.Contains([]string{"Application", "Workflow", "Rollout"}, resourceKind)
	isOperationCreate := operation == admission.Create

	return isKindInArgoCRDListThatShouldBeValidatedOnlyOnCreate && isOperationCreate || !isKindInArgoCRDListThatShouldBeValidatedOnlyOnCreate
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

func doesAtLeastOneFieldManagerStartWithOneOfThePrefixes(fields []ManagedFields, prefixes []string) bool {
	for _, field := range fields {
		for _, prefix := range prefixes {
			if strings.HasPrefix(field.Manager, prefix) {
				return true
			}
		}
	}
	return false
}

func isAtLeastOneFieldManagerEqualToOneOfThePrefixes(fields []ManagedFields, prefixes []string) bool {
	for _, field := range fields {
		for _, prefix := range prefixes {
			if field.Manager == prefix {
				return true
			}
		}
	}
	return false
}
