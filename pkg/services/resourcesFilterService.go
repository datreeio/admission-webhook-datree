package services

import (
	"encoding/json"
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
		panic(err)
	}

	resourceKind := admissionReviewReq.Request.Kind.Kind
	managedFields := rootObject.Metadata.ManagedFields

	// assigning to variables for easier debugging
	isMetadataNameExists := isMetadataNameExists(rootObject)
	isUnsupportedKind := isUnsupportedKind(resourceKind)
	isResourceDeleted := isResourceDeleted(rootObject)
	arePrerequisitesMet := isMetadataNameExists && !isUnsupportedKind && !isResourceDeleted

	if !arePrerequisitesMet {
		return false
	}

	isKubectl := isKubectl(managedFields)
	isHelm := isHelm(managedFields)
	isFluxResourceThatShouldBeEvaluated := isFluxResourceThatShouldBeEvaluated(admissionReviewReq, rootObject, managedFields)
	isArgoResourceThatShouldBeEvaluated := isArgoResourceThatShouldBeEvaluated(admissionReviewReq, resourceKind, managedFields)
	isResourceWhiteListed := isKubectl || isHelm || isFluxResourceThatShouldBeEvaluated || isArgoResourceThatShouldBeEvaluated

	if !isResourceWhiteListed {
		return false
	}

	return true
}

func isMetadataNameExists(rootObject RootObject) bool {
	return rootObject.Metadata.Name != ""
}

func isUnsupportedKind(resourceKind string) bool {
	unsupportedResourceKinds := []string{"Event", "GitRepository"}
	return slices.Contains(unsupportedResourceKinds, resourceKind)
}

func isResourceDeleted(rootObject RootObject) bool {
	return rootObject.Metadata.DeletionTimestamp != ""
}

func isKubectl(managedFields []ManagedFields) bool {
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
	return isAtLeastOneFieldManagerEqualToOneOfTheExpectedFieldManagers(managedFields, []string{"kubectl-client-side-apply", "kubectl-create", "kubectl-edit", "kubectl-patch"})
}

func isHelm(managedFields []ManagedFields) bool {
	return isAtLeastOneFieldManagerEqualToOneOfTheExpectedFieldManagers(managedFields, []string{"helm"})
}

func isFluxResourceThatShouldBeEvaluated(admissionReviewReq *admission.AdmissionReview, rootObject RootObject, managedFields []ManagedFields) bool {
	isDryRun := *admissionReviewReq.Request.DryRun
	labels := rootObject.Metadata.Labels
	namespace := admissionReviewReq.Request.Namespace

	if !doesAtLeastOneFieldManagerStartWithOneOfThePrefixes(managedFields, []string{"kustomize-controller"}) {
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

func isArgoResourceThatShouldBeEvaluated(admissionReviewReq *admission.AdmissionReview, resourceKind string, managedFields []ManagedFields) bool {
	operation := admissionReviewReq.Request.Operation

	if !doesAtLeastOneFieldManagerStartWithOneOfThePrefixes(managedFields, []string{"argocd", "argo"}) {
		return false
	}

	isKindInArgoCRDListThatShouldBeValidatedOnlyOnCreate := slices.Contains([]string{"Application", "Workflow", "Rollout"}, resourceKind)
	isOperationCreate := operation == admission.Create

	return isKindInArgoCRDListThatShouldBeValidatedOnlyOnCreate && isOperationCreate || !isKindInArgoCRDListThatShouldBeValidatedOnlyOnCreate
}

func isFluxObject(labels map[string]string, namespace string) bool {
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

func doesAtLeastOneFieldManagerStartWithOneOfThePrefixes(managedFields []ManagedFields, prefixes []string) bool {
	for _, field := range managedFields {
		for _, prefix := range prefixes {
			if strings.HasPrefix(field.Manager, prefix) {
				return true
			}
		}
	}
	return false
}

func isAtLeastOneFieldManagerEqualToOneOfTheExpectedFieldManagers(fields []ManagedFields, expectedFieldManagers []string) bool {
	for _, field := range fields {
		for _, expectedFieldManager := range expectedFieldManagers {
			if field.Manager == expectedFieldManager {
				return true
			}
		}
	}
	return false
}
