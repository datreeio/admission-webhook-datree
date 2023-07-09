package openshiftService

import (
	"context"
	userClientV1Api "github.com/openshift/api/user/v1"
	userClientV1 "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type openshiftClient struct {
	userClientV1 userClientV1.UserV1Interface
}

func newOpenshiftClient() (*openshiftClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	userClientV1Instance, err := userClientV1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &openshiftClient{
		userClientV1: userClientV1Instance,
	}, nil
}

func (oc *openshiftClient) getGroups() (*userClientV1Api.GroupList, error) {
	return oc.userClientV1.Groups().List(context.TODO(), metav1.ListOptions{})
}
