package k8sMetadataUtil

import (
	"context"
	"os"
	"time"

	"k8s.io/client-go/kubernetes/fake"

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
	ClientSet kubernetes.Interface
}

func NewK8sMetadataUtil() *K8sMetadataUtil {
	clientset, err := getClientSet()
	if err != nil {
		return &K8sMetadataUtil{}
	}
	return &K8sMetadataUtil{
		ClientSet: clientset,
	}
}

func NewK8sMetadataUtilMock() *K8sMetadataUtil {
	return &K8sMetadataUtil{
		ClientSet: fake.NewSimpleClientset(),
	}
}

//func NewK8sMetadataUtilMock() *K8sMetadataUtilMock {
//	return &K8sMetadataUtilMock{
//		//ClientSet: ,
//		ClientSet: fakeclientset.NewSimpleClientset(&v1.Pod{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:        "influxdb-v2",
//				Namespace:   "default",
//				Annotations: map[string]string{},
//			},
//		}, &v1.Pod{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:        "chronograf",
//				Namespace:   "default",
//				Annotations: map[string]string{},
//			},
//		}),
//	}
//}

func (k8sMDU *K8sMetadataUtil) InitK8sMetadataUtil() {

	validator := networkValidator.NewNetworkValidator()
	cliClient := cliClient.NewCliServiceClient(deploymentConfig.URL, validator)

	var clusterUuid k8sTypes.UID

	k8sClient, err := getClientSet()

	if err != nil {
		sendK8sMetadata(-1, err, clusterUuid, cliClient)
		return
	}

	clusterUuid, err = k8sMDU.GetClusterUuid()
	if err != nil {
		sendK8sMetadata(-1, err, clusterUuid, cliClient)
	}

	nodesCount, nodesCountErr := getNodesCount(k8sClient)
	sendK8sMetadata(nodesCount, nodesCountErr, clusterUuid, cliClient)

	cornJob := cron.New(cron.WithLocation(time.UTC))
	cornJob.AddFunc("@hourly", func() {
		nodesCount, nodesCountErr := getNodesCount(k8sClient)
		sendK8sMetadata(nodesCount, nodesCountErr, clusterUuid, cliClient)
	})
	cornJob.Start()
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

var ClusterUuid k8sTypes.UID = ""

func (k8sMDU *K8sMetadataUtil) GetClusterUuid() (k8sTypes.UID, error) {
	clusterMetadata, err := k8sMDU.ClientSet.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	ClusterUuid = clusterMetadata.UID
	return clusterMetadata.UID, nil
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
