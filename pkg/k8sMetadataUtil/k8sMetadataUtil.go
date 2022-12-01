package k8sMetadataUtil

import (
	"context"
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
	"k8s.io/client-go/rest"
)

type K8sMetadataUtil struct {
	ClientSet            kubernetes.Interface
	CreateClientSetError error
}

var ClusterUuid k8sTypes.UID = ""

func NewK8sMetadataUtil() *K8sMetadataUtil {
	clientset, err := getClientSet()
	if err != nil {
		return &K8sMetadataUtil{
			CreateClientSetError: err,
		}
	}
	return &K8sMetadataUtil{
		ClientSet: clientset,
	}
}

func (k8sMDU *K8sMetadataUtil) InitK8sMetadataUtil() {

	validator := networkValidator.NewNetworkValidator()
	cliClient := cliClient.NewCliServiceClient(deploymentConfig.URL, validator)

	var clusterUuid k8sTypes.UID

	if k8sMDU.CreateClientSetError != nil {
		sendK8sMetadata(-1, k8sMDU.CreateClientSetError, clusterUuid, cliClient)
		return
	}

	clusterUuid, err := k8sMDU.GetClusterUuid()
	if err != nil {
		sendK8sMetadata(-1, err, clusterUuid, cliClient)
	}

	nodesCount, nodesCountErr := getNodesCount(k8sMDU.ClientSet)
	sendK8sMetadata(nodesCount, nodesCountErr, clusterUuid, cliClient)

	cornJob := cron.New(cron.WithLocation(time.UTC))
	cornJob.AddFunc("@hourly", func() {
		nodesCount, nodesCountErr := getNodesCount(k8sMDU.ClientSet)
		sendK8sMetadata(nodesCount, nodesCountErr, clusterUuid, cliClient)
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

func getClientSet() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func (k8sMDU *K8sMetadataUtil) GetClusterUuid() (k8sTypes.UID, error) {
	if ClusterUuid != "" {
		return ClusterUuid, nil
	}

	if k8sMDU.CreateClientSetError != nil {
		return "", k8sMDU.CreateClientSetError
	} else {
		clusterMetadata, err := k8sMDU.ClientSet.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		ClusterUuid = clusterMetadata.UID
	}

	return ClusterUuid, nil
}

func sendK8sMetadata(nodesCount int, nodesCountErr error, clusterUuid k8sTypes.UID, client *cliClient.CliClient) {
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
