package k8sMetadataUtil

import (
	"context"
	"fmt"
	"os"
	"time"

	cliclient "github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	licensemanagerclient "github.com/datreeio/admission-webhook-datree/pkg/licenseManagerClient"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
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
	cliClient := cliclient.NewCliServiceClient(deploymentConfig.URL, validator)

	k8sClient, err := getClientSet()
	if err != nil {
		cliClient.ReportK8sMetadata(&cliclient.ReportK8sMetadataRequest{
			ClusterUuid:   "",
			Token:         os.Getenv(enums.Token),
			NodesCount:    -1,
			NodesCountErr: err.Error(),
		})
		return
	}

	clusterUuid, err := getClusterUuid(k8sClient)
	if err != nil {
		cliClient.ReportK8sMetadata(&cliclient.ReportK8sMetadataRequest{
			ClusterUuid:   clusterUuid,
			Token:         os.Getenv(enums.Token),
			NodesCount:    -1,
			NodesCountErr: err.Error(),
		})
	}

	runHourlyNodesCountCronJob(k8sClient, cliClient, clusterUuid)

	if os.Getenv(enums.AWSMarketplaceEnableCheckEntitlement) == "true" {
		runDailyAWSCheckoutLicenseCronJob(k8sClient, cliClient, clusterUuid)
	}

}

func runHourlyNodesCountCronJob(k8sClient *kubernetes.Clientset, cliClient *cliclient.CliClient, clusterUuid k8sTypes.UID) {
	cornJob := cron.New(cron.WithLocation(time.UTC))
	cornJob.AddFunc("@hourly", func() {
		nodesCount, nodesCountErr := getNodesCount(k8sClient)
		cliClient.ReportK8sMetadata(&cliclient.ReportK8sMetadataRequest{
			ClusterUuid:   clusterUuid,
			Token:         os.Getenv(enums.Token),
			NodesCount:    nodesCount,
			NodesCountErr: nodesCountErr.Error(),
		})
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

func getClusterUuid(clientset *kubernetes.Clientset) (k8sTypes.UID, error) {
	clusterMetadata, err := clientset.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	ClusterUuid = clusterMetadata.UID
	return clusterMetadata.UID, nil
}

// run chckout license cron job daily to check if aws marketplace license is valid with the nodes number
func runDailyAWSCheckoutLicenseCronJob(k8sClient *kubernetes.Clientset, cliClient *cliclient.CliClient, clusterUuid k8sTypes.UID) {
	licenseManagerClient := licensemanagerclient.NewLicenseManagerClient()

	licenseCheckerCornJob := cron.New(cron.WithLocation(time.UTC))
	// @daily means run once a day, midnight. On debug mode it's very nice to run it once a minute i.e '@every 1min'
	licenseCheckerCornJob.AddFunc("@daily", func() {
		nodesCount, err := getNodesCount(k8sClient)
		if err != nil {
			// should be on debug
			logger.LogUtil(fmt.Sprint("failed counting nodes for checkout", err))
			cliClient.ReportK8sMetadata(&cliclient.ReportK8sMetadataRequest{
				ClusterUuid:   clusterUuid,
				Token:         os.Getenv(enums.Token),
				NodesCount:    -1,
				NodesCountErr: err.Error(),
			})
			return
		}

		// should be on debug
		logger.LogUtil(fmt.Sprint("checking aws marketplace license with nodes count", nodesCount))
		err = licenseManagerClient.CheckoutLicense(nodesCount)
		if err != nil {
			// should be on debug
			logger.LogUtil(fmt.Sprint("checkout license failed: ", err))
			cliClient.ReportK8sMetadata(&cliclient.ReportK8sMetadataRequest{
				ClusterUuid:   clusterUuid,
				Token:         os.Getenv(enums.Token),
				NodesCount:    -1,
				NodesCountErr: err.Error(),
			})
		}
	})
	licenseCheckerCornJob.Start()
}
