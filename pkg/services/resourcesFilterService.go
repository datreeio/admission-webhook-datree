package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/datreeio/admission-webhook-datree/pkg/server"
	"github.com/google/go-cmp/cmp"
	admission "k8s.io/api/admission/v1"
	"k8s.io/utils/strings/slices"
)

type OwnerReference struct {
	ApiVersion         string `json:"apiVersion"`
	Kind               string `json:"kind"`
	Name               string `json:"name"`
	Uid                string `json:"uid"`
	Controller         bool   `json:"controller"`
	BlockOwnerDeletion bool   `json:"blockOwnerDeletion"`
}

type RootObject struct {
	Metadata Metadata `json:"metadata"`
}

type ShouldValidatedResourceData struct {
	ShouldValidate     bool
	OpenShiftRequester string
}

func ShouldResourceBeValidated(admissionReviewReq *admission.AdmissionReview, rootObject RootObject) ShouldValidatedResourceData {

	if admissionReviewReq == nil {
		panic("admissionReviewReq is nil")
	}

	resourceKind := admissionReviewReq.Request.Kind.Kind
	managedFields := rootObject.Metadata.ManagedFields
	userInfo := admissionReviewReq.Request.UserInfo
	resourceAnnotations := rootObject.Metadata.Annotations

	// assigning to variables for easier debugging
	isMetadataNameExists := isMetadataNameExists(rootObject)
	isUnsupportedKind := isUnsupportedKind(resourceKind)
	isResourceDeleted := isResourceDeleted(rootObject)
	isNamespaceThatShouldBeSkipped := isNamespaceThatShouldBeSkipped(admissionReviewReq)
	arePrerequisitesMet := isMetadataNameExists && !isUnsupportedKind && !isResourceDeleted && !isNamespaceThatShouldBeSkipped

	if !arePrerequisitesMet {
		return ShouldValidatedResourceData{
			ShouldValidate: false,
		}
	}

	if hasOwnerReference(rootObject) {
		return ShouldValidatedResourceData{
			ShouldValidate: false,
		}
	}

	if !isUsernamePrefixSystem(userInfo.Username) {
		return ShouldValidatedResourceData{
			ShouldValidate: true,
		}
	}

	if isEqualObjectAndOldObject(admissionReviewReq) {
		return ShouldValidatedResourceData{
			ShouldValidate: false,
		}
	}

	isKubectl := isKubectl(managedFields)
	isHelm := isHelm(managedFields)
	isTerraform := isTerraform(managedFields)
	isFluxResourceThatShouldBeEvaluated := isFluxResourceThatShouldBeEvaluated(admissionReviewReq, rootObject, managedFields)
	isArgoResourceThatShouldBeEvaluated := isArgoResourceThatShouldBeEvaluated(admissionReviewReq, resourceKind, managedFields)
	isOpenshiftResourceThatShouldBeEvaluated, openShiftRequester := isOpenshiftResourceThatShouldBeEvaluated(managedFields, resourceAnnotations)
	isResourceWhiteListed := isKubectl || isHelm || isTerraform || isFluxResourceThatShouldBeEvaluated || isArgoResourceThatShouldBeEvaluated || isOpenshiftResourceThatShouldBeEvaluated

	return ShouldValidatedResourceData{
		ShouldValidate:     isResourceWhiteListed,
		OpenShiftRequester: openShiftRequester,
	}

}

func ShouldResourceBeSkippedByConfigMapScanningFilters(admissionReviewReq *admission.AdmissionReview, rootObject RootObject) bool {
	namespace := admissionReviewReq.Request.Namespace
	resourceKind := admissionReviewReq.Request.Kind.Kind
	resourceName := rootObject.Metadata.Name

	for _, skipListItem := range server.ConfigMapScanningFilters.SkipList {
		skipRuleItem := strings.Split(skipListItem, ";")

		if len(skipRuleItem) != 3 {
			continue
		}

		if doesRegexMatchString(skipRuleItem[0], namespace) &&
			doesRegexMatchString(skipRuleItem[1], resourceKind) &&
			doesRegexMatchString(skipRuleItem[2], resourceName) {
			return true
		}
	}

	return false
}

func isMetadataNameExists(rootObject RootObject) bool {
	return rootObject.Metadata.Name != ""
}

func isUnsupportedKind(resourceKind string) bool {
	unsupportedResourceKinds := []string{"Event", "GitRepository", "SubjectAccessReview", "SelfSubjectAccessReview"}
	return slices.Contains(unsupportedResourceKinds, resourceKind)
}

func isResourceDeleted(rootObject RootObject) bool {
	return rootObject.Metadata.DeletionTimestamp != ""
}

func isNamespaceThatShouldBeSkipped(admissionReviewReq *admission.AdmissionReview) bool {
	namespacesToSkip := []string{"kube-public", "kube-node-lease"}
	return slices.Contains(namespacesToSkip, admissionReviewReq.Request.Namespace)
}

func isEqualObjectAndOldObject(admissionReviewReq *admission.AdmissionReview) bool {
	fmt.Println("@@admissionReviewReq", admissionReviewReq)
	if admissionReviewReq.Request.OldObject.Raw == nil {
		return false
	}
	if admissionReviewReq.Request.Operation != admission.Update {
		return false
	}
	clonedObject := admissionReviewReq.Request.Object.DeepCopy()
	clonedOldObject := admissionReviewReq.Request.OldObject.DeepCopy()

	var objectMap map[string]interface{}
	var oldObjectMap map[string]interface{}
	_ = json.Unmarshal(clonedObject.Raw, &objectMap)
	_ = json.Unmarshal(clonedOldObject.Raw, &oldObjectMap)

	if objectMetadata, ok := objectMap["metadata"]; ok {
		delete(objectMetadata.(map[string]interface{}), "managedFields")
		delete(objectMetadata.(map[string]interface{}), "selfLink")
	}

	if oldObjectMetadata, ok := oldObjectMap["metadata"]; ok {
		delete(oldObjectMetadata.(map[string]interface{}), "managedFields")
		delete(oldObjectMetadata.(map[string]interface{}), "selfLink")
	}

	fmt.Println("@@clonedObject", string(clonedObject.Raw))
	fmt.Println("@@clonedOldObject", string(clonedOldObject.Raw))

	isEqual := cmp.Equal(objectMap, oldObjectMap)
	diff := cmp.Diff(objectMap, oldObjectMap)
	fmt.Println("@@isEqual", isEqual)
	fmt.Println("@@diff", diff)
	return isEqual
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

func isTerraform(managedFields []ManagedFields) bool {
	/**
	Also supports Terragrunt: https://github.com/gruntwork-io/terragrunt
	Default terraform field manager: "Terraform"
	https://github.com/hashicorp/terraform-provider-kubernetes/blob/aa76ff0f804cf52d98a0f2ac21f9d7e9c225c585/manifest/provider/plan.go#L68
	Some users also have a field manager similar to this one: "terraform-provider-helm_v2.6.0_x5"
	Therefore, we check if a field manager contains the case-insensitive "terraform"
	Some users have the field manager "HashiCorp", therefore we add it as well
	*/
	return doesAtLeastOneFieldManagerContainOneOfTheInputsNotCaseSensitive(managedFields, []string{"terraform", "hashicorp"})
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
	return !badFluxObject
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

func doesAtLeastOneFieldManagerContainOneOfTheInputsNotCaseSensitive(managedFields []ManagedFields, inputs []string) bool {
	for _, field := range managedFields {
		for _, input := range inputs {
			if strings.Contains(strings.ToLower(field.Manager), strings.ToLower(input)) {
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

func doesRegexMatchString(regex string, str string) bool {
	r, err := regexp.Compile(regex)
	if err != nil {
		return false
	}
	return r.MatchString(str)
}

func isOpenshiftResourceThatShouldBeEvaluated(managedFields []ManagedFields, annotations map[string]string) (bool, string) {
	if val, ok := annotations["openshift.io/requester"]; ok {
		if !strings.HasPrefix(val, "system:serviceaccount") {
			// If the value of "openshift.io/requester" does not start with "system:serviceaccount",
			// it indicates that the resource is created by a human and should be evaluated.
			return true, val
		}
	}

	return false, ""
}

func hasOwnerReference(resource RootObject) bool {
	if resource.Metadata.OwnerReferences == nil {
		return false
	}

	for _, owner := range resource.Metadata.OwnerReferences {
		if owner.Kind != "" && owner.Name != "" {
			return true
		}
	}
	return false
}

func isUsernamePrefixSystem(username string) bool {
	return strings.HasPrefix(username, "system:")
}
