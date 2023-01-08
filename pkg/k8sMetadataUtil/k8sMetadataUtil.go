package k8sMetadataUtil

import (
	"context"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/leaderElection"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"os"
	"time"

	cliClient "github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/datreeio/datree/pkg/deploymentConfig"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type K8sMetadataUtil struct {
	ClientSet            kubernetes.Interface
	CreateClientSetError error
	leaderElection       *leaderElection.LeaderElection
	internalLogger       logger.Logger
}

var ClusterUuid k8sTypes.UID = ""

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

func (k8sMetadataUtil *K8sMetadataUtil) InitK8sMetadataUtil() {
	validator := networkValidator.NewNetworkValidator()
	cliClient := cliClient.NewCliServiceClient(deploymentConfig.URL, validator)

	var clusterUuid k8sTypes.UID

	if k8sMetadataUtil.CreateClientSetError != nil {
		k8sMetadataUtil.sendK8sMetadataIfLeader(-1, k8sMetadataUtil.CreateClientSetError, clusterUuid, cliClient)
		return
	}

	clusterUuid, err := k8sMetadataUtil.GetClusterUuid()
	if err != nil {
		k8sMetadataUtil.sendK8sMetadataIfLeader(-1, err, clusterUuid, cliClient)
	}

	nodesCount, nodesCountErr := getNodesCount(k8sMetadataUtil.ClientSet)
	k8sMetadataUtil.sendK8sMetadataIfLeader(nodesCount, nodesCountErr, clusterUuid, cliClient)

	cornJob := cron.New(cron.WithLocation(time.UTC))
	cornJob.AddFunc("* * * * *", func() {
		fmt.Println("running cron")
		nodesCount, nodesCountErr := getNodesCount(k8sMetadataUtil.ClientSet)
		k8sMetadataUtil.sendK8sMetadataIfLeader(nodesCount, nodesCountErr, clusterUuid, cliClient)
	})
	cornJob.Start()
}

func getNodesCount(clientset kubernetes.Interface) (int, error) {
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return -1, err
	}

	return len(nodes.Items), nil
}

func (k8sMetadataUtil *K8sMetadataUtil) GetClusterUuid() (k8sTypes.UID, error) {
	if ClusterUuid != "" {
		return ClusterUuid, nil
	}

	if k8sMetadataUtil.CreateClientSetError != nil {
		return "", k8sMetadataUtil.CreateClientSetError
	} else {
		clusterMetadata, err := k8sMetadataUtil.ClientSet.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		ClusterUuid = clusterMetadata.UID
	}

	return ClusterUuid, nil
}

func (k8sMetadataUtil *K8sMetadataUtil) sendK8sMetadataIfLeader(nodesCount int, nodesCountErr error, clusterUuid k8sTypes.UID, client *cliClient.CliClient) {
	if !k8sMetadataUtil.leaderElection.IsLeader() {
		fmt.Println("not leader")
		return
	}
	fmt.Println("sending k8s metadata")
	token := os.Getenv(enums.Token)

	var nodesCountErrString string
	if nodesCountErr != nil {
		nodesCountErrString = nodesCountErr.Error()
	}

	client.ReportK8sMetadata(&cliClient.ReportK8sMetadataRequest{
		ClusterUuid:   clusterUuid,
		Token:         token,
		NodesCount:    nodesCount,
		NodesCountErr: nodesCountErrString,
	})
}
