package k8sMetadataUtil

import (
	"context"
	"fmt"
	"os"
	"time"

	cliClient "github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"github.com/datreeio/datree/pkg/deploymentConfig"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var ClusterUuid k8sTypes.UID = ""

func InitK8sMetadataUtil() {

	validator := networkValidator.NewNetworkValidator()
	cliClient := cliClient.NewCliServiceClient(deploymentConfig.URL, validator)

	k8sClient, err := getClientSet()

	if err != nil {
		sendK8sMetadata(-1, err, ClusterUuid, cliClient)
		return
	}

	err = setClusterUuid(k8sClient)
	if err != nil {
		sendK8sMetadata(-1, err, ClusterUuid, cliClient)
	}

	nodesCount, nodesCountErr := getNodesCount(k8sClient)
	sendK8sMetadata(nodesCount, nodesCountErr, ClusterUuid, cliClient)

	cronJob := cron.New(cron.WithLocation(time.UTC))
	cronJob.AddFunc("@hourly", func() {
		nodesCount, nodesCountErr := getNodesCount(k8sClient)
		sendK8sMetadata(nodesCount, nodesCountErr, ClusterUuid, cliClient)
	})

	var setClusterUuidJobEntryID cron.EntryID

	setClusterUuidJobEntryID, err = cronJob.AddFunc("@every 1m", func() {
		setClusterUuid(k8sClient)
		if ClusterUuid != "" && setClusterUuidJobEntryID != 0 {
			cronJob.Remove(setClusterUuidJobEntryID)
		}
		logger.LogUtil("ClusterUuid wasn't set yet, trying again in a minute")
	})
	if err != nil {
		logger.LogUtil(fmt.Sprintf("Could not create setClusterUuid cronjob err: %s", err.Error()))
	}

	cronJob.Start()
}

func getNodesCount(clientset *kubernetes.Clientset) (int, error) {
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

func setClusterUuid(clientset *kubernetes.Clientset) error {
	if ClusterUuid != "" {
		return nil
	}

	clusterMetadata, err := clientset.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	if err != nil {
		return err
	}
	ClusterUuid = clusterMetadata.UID
	return nil
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
