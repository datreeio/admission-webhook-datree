package k8sMetadataUtil

import (
	"context"
	"errors"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var nodesCount int
var nodesCountErr error

func InitK8sMetadataUtil() {
	k8sClient, err := getClientSet()
	if err != nil {
		errString := fmt.Sprintf("getClientSet error: %s", err.Error())
		fmt.Println(errString)
		nodesCount = -1
		nodesCountErr = errors.New(errString)
		return
	}
	go func() {
		for {
			setNodesCount(k8sClient)
			time.Sleep(time.Hour)
		}
	}()
}

func GetNodesCount() (int, error) {
	return nodesCount, nodesCountErr
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

func setNodesCount(clientset *kubernetes.Clientset) {

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		errString := fmt.Sprintf("List error: %s", err.Error())
		fmt.Println(errString)
		nodesCount = -1
		nodesCountErr = errors.New(errString)
		return
	}

	nodesCount = len(nodes.Items)
	nodesCountErr = nil
}
