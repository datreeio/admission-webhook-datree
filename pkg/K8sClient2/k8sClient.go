package K8sClient

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"os/exec"
	"strings"

	"k8s.io/client-go/discovery"

	"github.com/datreeio/datree/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sClient struct {
	clientSet       *kubernetes.Clientset
	discoveryClient *discovery.DiscoveryClient
	errorReporter   *errorReporter.ErrorReporter
	config          *rest.Config
}

func NewK8sClient(errorReporter *errorReporter.ErrorReporter) (*K8sClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		clientSet:       clientSet,
		discoveryClient: discoveryClient,
		errorReporter:   errorReporter,
		config:          config,
	}, nil
}

func (kc *K8sClient) GetAllGroups(namespacesToSkip []string, errorReporter *errorReporter.ErrorReporter) (*RawK8sResources, error) {
	resources, err := kc.getResources(namespacesToSkip, errorReporter, []string{"groups"})
	if err != nil {
		return nil, err
	}
	return &resources, nil
}

type RawK8sResources = []RawK8sResource

type RawK8sResource = map[string]interface{}

type K8sResource struct {
	ApiVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Metadata   K8sResourceMetadata `json:"metadata"`
}

type OwnerReference struct {
	Kind string `json:"kind"`
}

type K8sResourceMetadata struct {
	Name            string           `json:"name"`
	Namespace       string           `json:"namespace"`
	OwnerReferences []OwnerReference `json:"ownerReferences"`
}

type kubectlGetAllResourcesResp struct {
	ApiVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Items      RawK8sResources `json:"items"`
}

func (kc *K8sClient) getResources(namespacesToSkip []string, errorReporter *errorReporter.ErrorReporter, resourcesKindsToGet []string) (RawK8sResources, error) {
	commandArgs := []string{"get", strings.Join(resourcesKindsToGet, ","), "--all-namespaces", "-o", "json", "--selector", "kubernetes.io/bootstrapping!=rbac-defaults,app.kubernetes.io/part-of!=datree"}

	if len(namespacesToSkip) > 0 {
		// command should look like this: kubectl get all --all-namespaces -o json --field-selector metadata.namespace!=<SKIPPED-NAMESPACE-NAME>,metadata.namespace!=kube-system,metadata.namespace!=<DATREE_NAMESPACE>
		fieldSelectorParams := "metadata.namespace!="
		fieldSelectorParams += strings.Join(namespacesToSkip, fmt.Sprintf(",%s", fieldSelectorParams))

		commandArgs = append(commandArgs, "--field-selector", fieldSelectorParams)
	}

	stdOutBuffer, stdErrBuffer, err := kc.kubectlExec(commandArgs)

	if stdOutBuffer.Len() == 0 {
		// this is a real error that causes the scan to fail
		if err != nil {
			return nil, errors.New(fmt.Sprint(err) + ": " + stdErrBuffer.String())
		} else {
			return nil, errors.New("stdOut is empty: " + stdErrBuffer.String())
		}
	}

	if err != nil {
		// this is a minor error that should be reported but not fail the scan
		errStr := "error in getResources: " + stdErrBuffer.String()
		fmt.Println(errStr)
		errorReporter.ReportUnexpectedError(errors.New(errStr))
	}

	getAllResourcesResp := &kubectlGetAllResourcesResp{}
	err = json.Unmarshal(stdOutBuffer.Bytes(), &getAllResourcesResp)
	if err != nil {
		fmt.Printf("couldn't get the resources in cluster %s \n", err.Error())

		if e, ok := err.(*json.SyntaxError); ok {
			fmt.Printf("syntax error at byte offset %d \n", e.Offset)
		}

		return nil, err
	}

	return getAllResourcesResp.Items, nil
}

func (kc *K8sClient) getApiResourcesKinds(errorReporter *errorReporter.ErrorReporter) ([]string, error) {
	stdOutBuffer, stdErrBuffer, err := kc.kubectlExec([]string{"api-resources", "-o", "name"})

	if stdOutBuffer.Len() == 0 {
		// this is a real error that causes the scan to fail
		if err != nil {
			return nil, errors.New(fmt.Sprint(err) + ": " + stdErrBuffer.String())
		} else {
			return nil, errors.New("stdOut is empty: " + stdErrBuffer.String())
		}
	}

	if err != nil {
		// this is a minor error that should be reported but not fail the scan
		errStr := "error in fetch api-resources: " + stdErrBuffer.String()
		fmt.Println(errStr)
		errorReporter.ReportUnexpectedError(errors.New(errStr))
	}

	apiResourcesKinds := rawApiResourcesToApiResourcesKinds(stdOutBuffer.String())
	return apiResourcesKinds, nil
}

func rawApiResourcesToApiResourcesKinds(rawApiResources string) []string {
	apiResourcesKinds := utils.MapSlice(strings.Split(rawApiResources, "\n"), func(resource string) string {
		return strings.Split(resource, ".")[0]
	})

	var apiResourcesKindsNotEmpty []string
	for _, resourceKind := range apiResourcesKinds {
		if resourceKind != "" {
			apiResourcesKindsNotEmpty = append(apiResourcesKindsNotEmpty, resourceKind)
		}
	}

	return apiResourcesKindsNotEmpty
}

func (kc *K8sClient) kubectlExec(args []string) (bytes.Buffer, bytes.Buffer, error) {
	cmd := exec.Command("kubectl", args...)
	// we want to separate stdErr and stdOut, therefore we use cmd.Run() instead of cmd.CombinedOutput()
	var stdOutBuffer, stdErrBuffer bytes.Buffer
	cmd.Stdout = &stdOutBuffer
	cmd.Stderr = &stdErrBuffer
	err := cmd.Run()
	return stdOutBuffer, stdErrBuffer, err
}
