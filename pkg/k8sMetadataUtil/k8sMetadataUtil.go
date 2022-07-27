package k8sMetadataUtil

import (
	"context"
	"fmt"
	"time"

	cliClient "github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/datree/pkg/deploymentConfig"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func InitK8sMetadataUtil() {
	k8sClient, err := getClientSet()

	if err != nil {
		fmt.Println("Error getting k8s client", err)
		return
	}

	validator := networkValidator.NewNetworkValidator()
	cliClient := cliClient.NewCliServiceClient(deploymentConfig.URL, validator)
	cornJob := cron.New(cron.WithLocation(time.UTC))
	cornJob.AddFunc("@hourly", func() { sendK8sMetadata(k8sClient, cliClient) })
	cornJob.Start()
}

func getNodesCount(clientset *kubernetes.Clientset) (int, error) {
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0, err
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

func getClusterUuid(clientset *kubernetes.Clientset) (types.UID, error) {
	clusterMetadata, err := clientset.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return clusterMetadata.UID, nil
}

func sendK8sMetadata(clientset *kubernetes.Clientset, client *cliClient.CliClient) {
	nodesCount, nodesCountErr := getNodesCount(clientset)
	clusterUuid, _ := getClusterUuid(clientset)

	client.ReportK8sMetadata(&cliClient.ReportK8sMetadataRequest{
		ClusterUuid:   clusterUuid,
		NodesCount:    nodesCount,
		NodesCountErr: nodesCountErr,
	})
}
