package k8sMetadataUtil

import (
	"context"
	"fmt"
	"os"
	"time"

	cliClient "github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	licensemanagerclient "github.com/datreeio/admission-webhook-datree/pkg/licenseManagerClient"
	"github.com/datreeio/datree/pkg/deploymentConfig"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func InitK8sMetadataUtil() {

	validator := networkValidator.NewNetworkValidator()
	cliClient := cliClient.NewCliServiceClient(deploymentConfig.URL, validator)
	k8sClient, err := getClientSet()

	if os.Getenv(enums.AWSMarketplaceEnableCheckoutLicense) == "true" {
		runDailyAWSCheckoutLicenseCronJob(k8sClient)
	}

	var clusterUuid k8sTypes.UID
	if err != nil {
		sendK8sMetadata(-1, err, clusterUuid, cliClient)
		return
	}

	clusterUuid, err = getClusterUuid(k8sClient)
	if err != nil {
		sendK8sMetadata(-1, err, clusterUuid, cliClient)
	}

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

func getClusterUuid(clientset *kubernetes.Clientset) (k8sTypes.UID, error) {
	clusterMetadata, err := clientset.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
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

func runDailyAWSCheckoutLicenseCronJob(k8sClient *kubernetes.Clientset) {
	licenseManagerClient := licensemanagerclient.NewLicenseManagerClient()

	licenseCheckerCornJob := cron.New(cron.WithLocation(time.UTC))
	licenseCheckerCornJob.AddFunc("@daily", func() {
		nodesCount, err := getNodesCount(k8sClient)
		if err != nil {
			fmt.Println("failed counting nodes for checkout", err)
		}

		err = licenseManagerClient.CheckoutLicense(nodesCount)
		if err != nil {
			fmt.Println("checkout license failed: ", err)
		}
	})
	licenseCheckerCornJob.Start()
}
