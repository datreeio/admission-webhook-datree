package k8sMetadataUtil

import (
	"context"
	"fmt"
	K8sClient "github.com/datreeio/admission-webhook-datree/pkg/K8sClient2"
	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"os"
	"strings"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/leaderElection"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	servicestate "github.com/datreeio/admission-webhook-datree/pkg/serviceState"

	cliClient "github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/datreeio/datree/pkg/deploymentConfig"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/robfig/cron/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sMetadataUtil struct {
	ClientSet            kubernetes.Interface
	CreateClientSetError error
	leaderElection       *leaderElection.LeaderElection
	internalLogger       logger.Logger
	errorReporter        *errorReporter.ErrorReporter
}

type K8sMetadata struct {
	ClusterUuid     k8sTypes.UID
	NodesCount      int
	NodesCountErr   error
	K8sDistribution string
	ActionOnFailure enums.ActionOnFailure
}

var ClusterUuid k8sTypes.UID = ""
var ClusterK8sVersion string = ""

func NewK8sMetadataUtil(clientset *kubernetes.Clientset, createClientSetError error, leaderElection *leaderElection.LeaderElection, internalLogger logger.Logger) *K8sMetadataUtil {
	if createClientSetError != nil {
		internalLogger.LogAndReportUnexpectedError("NewK8sMetadataUtil: failed to create k8s clientset: " + createClientSetError.Error())
		return &K8sMetadataUtil{
			CreateClientSetError: createClientSetError,
			leaderElection:       leaderElection,
			internalLogger:       internalLogger,
		}
	}
	return &K8sMetadataUtil{
		ClientSet:      clientset,
		leaderElection: leaderElection,
		internalLogger: internalLogger,
	}
}

func (k8sMetadataUtil *K8sMetadataUtil) InitK8sMetadataUtil(state *servicestate.ServiceState) {
	validator := networkValidator.NewNetworkValidator()
	cliClient := cliClient.NewCliServiceClient(deploymentConfig.URL, validator, state)

	var clusterUuid k8sTypes.UID
	var actionOnFailure enums.ActionOnFailure

	if os.Getenv(enums.Enforce) == "true" {
		actionOnFailure = enums.EnforceActionOnFailure
	} else {
		actionOnFailure = enums.MonitorActionOnFailure
	}

	if k8sMetadataUtil.CreateClientSetError != nil {
		k8sMetadataClientSetErr := K8sMetadata{
			NodesCount:      -1,
			NodesCountErr:   k8sMetadataUtil.CreateClientSetError,
			ClusterUuid:     clusterUuid,
			ActionOnFailure: actionOnFailure,
		}
		k8sMetadataUtil.sendK8sMetadata(cliClient, k8sMetadataClientSetErr)
		return
	}

	clusterUuid, err := k8sMetadataUtil.GetClusterUuid()
	if err != nil {
		k8sMetadataGetClusterUuidErr := K8sMetadata{
			NodesCount:      -1,
			NodesCountErr:   err,
			ClusterUuid:     clusterUuid,
			ActionOnFailure: actionOnFailure,
		}
		k8sMetadataUtil.sendK8sMetadata(cliClient, k8sMetadataGetClusterUuidErr)
	}

	nodesCount, nodes, nodesCountErr := getNodesCount(k8sMetadataUtil.ClientSet)
	k8sDistribution := getLocalK8sDistribution(k8sMetadataUtil.ClientSet, nodes)
	k8sMetadataOnInit := K8sMetadata{
		NodesCount:      nodesCount,
		NodesCountErr:   nodesCountErr,
		ClusterUuid:     clusterUuid,
		ActionOnFailure: actionOnFailure,
		K8sDistribution: k8sDistribution,
	}
	k8sMetadataUtil.sendK8sMetadata(cliClient, k8sMetadataOnInit)

	cornJob := cron.New(cron.WithLocation(time.UTC))

	_, err = cornJob.AddFunc("@hourly", func() {
		if k8sMetadataUtil.leaderElection.IsLeader() {
			nodesCount, nodes, nodesCountErr := getNodesCount(k8sMetadataUtil.ClientSet)
			k8sDistribution := getLocalK8sDistribution(k8sMetadataUtil.ClientSet, nodes)
			k8sMetadataHourly := K8sMetadata{
				NodesCount:      nodesCount,
				NodesCountErr:   nodesCountErr,
				ClusterUuid:     clusterUuid,
				ActionOnFailure: actionOnFailure,
				K8sDistribution: k8sDistribution,
			}
			k8sMetadataUtil.sendK8sMetadata(cliClient, k8sMetadataHourly)
		}

	})
	if err != nil {
		fmt.Printf("Cronjon failed to be added, err: %s \n", err.Error())
	}
	cornJob.Start()
}

func getNodesCount(clientset kubernetes.Interface) (int, *v1.NodeList, error) {
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return -1, nil, err
	}

	return len(nodes.Items), nodes, nil
}

func getLocalK8sDistribution(clientset kubernetes.Interface, nodes *v1.NodeList) string {
	if nodes == nil {
		return "Error getting nodes"
	}

	localK8sDistributions := []string{"minikube", "microk8s", "docker-desktop", "k3s", "k3d", "kind"}

	for _, node := range nodes.Items {
		_, isControlPlaneLabelExist := node.Labels["node-role.kubernetes.io/control-plane"]
		if isControlPlaneLabelExist {
			for key, value := range node.Labels {
				for _, localK8sDistribution := range localK8sDistributions {
					if strings.Contains(key, localK8sDistribution) || strings.Contains(value, localK8sDistribution) {
						return localK8sDistribution
					}
				}
			}
		}
	}

	return "probably not a local distribution"
}

func (k8sMetadataUtil *K8sMetadataUtil) GetClusterUuid() (k8sTypes.UID, error) {
	if ClusterUuid != "" {
		return ClusterUuid, nil
	}

	if k8sMetadataUtil.CreateClientSetError != nil {
		return "", k8sMetadataUtil.CreateClientSetError
	} else {
		// get the equivalent of "kubectl get groups" command, for openshift crd
		client2, err := K8sClient.NewK8sClient(k8sMetadataUtil.errorReporter)

		res, err := client2.GetAllGroups([]string{}, k8sMetadataUtil.errorReporter)

		fmt.Println(res, err)

		clusterMetadata, err := k8sMetadataUtil.ClientSet.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		ClusterUuid = clusterMetadata.UID
		fmt.Println("cluster uuid: ", ClusterUuid)
	}

	return ClusterUuid, nil
}

func (k8sMetadataUtil *K8sMetadataUtil) GetClusterK8sVersion() (string, error) {
	if ClusterK8sVersion != "" {
		return ClusterK8sVersion, nil
	}

	unknownVersion := "unknown k8s version"

	config, err := rest.InClusterConfig()
	if err != nil {
		ClusterK8sVersion = unknownVersion
		return unknownVersion, err
	}
	discClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		ClusterK8sVersion = unknownVersion
		return unknownVersion, err
	}

	serverInfo, err := discClient.ServerVersion()
	if err != nil {
		ClusterK8sVersion = unknownVersion
		return unknownVersion, err
	}

	if serverInfo.GitVersion == "" {
		ClusterK8sVersion = unknownVersion
		return unknownVersion, nil
	}

	ClusterK8sVersion = serverInfo.GitVersion
	return serverInfo.GitVersion, nil
}

func (k8sMetadataUtil *K8sMetadataUtil) sendK8sMetadata(client *cliClient.CliClient, k8sMetadata K8sMetadata) {
	token := os.Getenv(enums.Token)

	var nodesCountErrString string
	if k8sMetadata.NodesCountErr != nil {
		nodesCountErrString = k8sMetadata.NodesCountErr.Error()
	}

	client.ReportK8sMetadata(&cliClient.ReportK8sMetadataRequest{
		ClusterUuid:     k8sMetadata.ClusterUuid,
		Token:           token,
		NodesCount:      k8sMetadata.NodesCount,
		NodesCountErr:   nodesCountErrString,
		ActionOnFailure: k8sMetadata.ActionOnFailure,
		K8sDistribution: k8sMetadata.K8sDistribution,
	})
}
