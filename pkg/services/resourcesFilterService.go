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

func shouldResourceBeValidated(admissionReviewReq *admission.AdmissionReview) bool {
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
	return isMetadataNameExists(rootObject.Metadata) &&
		shouldEvaluateResourceByKind(resourceKind) &&
		shouldEvaluateArgoCRDResources(resourceKind, admissionReviewReq.Request.Operation) &&
		shouldEvaluateFluxCDResources(*admissionReviewReq.Request.DryRun, rootObject.Metadata.Labels, admissionReviewReq.Request.Namespace) &&
		shouldEvaluateResourceByManager(rootObject.Metadata.ManagedFields) &&
		rootObject.Metadata.DeletionTimestamp == ""
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
