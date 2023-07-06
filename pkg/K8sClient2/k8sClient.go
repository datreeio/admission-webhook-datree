package K8sClient

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"k8s.io/client-go/discovery"
	"os/exec"

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

type GroupResources = []GroupResource

type GroupResource = struct {
	Users    []string `json:"users"`
	Metadata struct {
		Name string `json:"name"`
	}
}

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

type kubectlGetAllGroupsResp struct {
	ApiVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Items      GroupResources `json:"items"`
}

func (kc *K8sClient) GetAllGroupsTheUserIsIn(username string) ([]string, error) {
	commandArgs := []string{"get", "groups", "-o", "json"}

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
		kc.errorReporter.ReportUnexpectedError(errors.New(errStr))
	}

	getAllGroupsResp := &kubectlGetAllGroupsResp{}
	err = json.Unmarshal(stdOutBuffer.Bytes(), &getAllGroupsResp)
	if err != nil {
		fmt.Printf("couldn't get the resources in cluster %s \n", err.Error())

		if e, ok := err.(*json.SyntaxError); ok {
			fmt.Printf("syntax error at byte offset %d \n", e.Offset)
		}

		return nil, err
	}

	allTheGroupsTheUserIsIn := extractAllTheGroupsTheUserIsIn(username, getAllGroupsResp)

	return allTheGroupsTheUserIsIn, nil
}

func extractAllTheGroupsTheUserIsIn(username string, getAllGroupsResp *kubectlGetAllGroupsResp) []string {
	allTheGroupsTheUserIsIn := make([]string, 0)
	for _, group := range getAllGroupsResp.Items {
		for _, user := range group.Users {
			if user == username {
				allTheGroupsTheUserIsIn = append(allTheGroupsTheUserIsIn, group.Metadata.Name)
			}
		}
	}
	return allTheGroupsTheUserIsIn
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
