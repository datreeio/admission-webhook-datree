package main

import (
	"context"
	userClientV1 "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type OpenshiftClient struct {
	userClientV1 userClientV1.UserV1Interface
}

func NewOpenshiftClient() (*OpenshiftClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// get all groups
	userClientV1Instance, err := userClientV1.NewForConfig(config)

	if err != nil {
		return nil, err
	}

	return &OpenshiftClient{
		userClientV1: userClientV1Instance,
	}, nil
}

func (oc *OpenshiftClient) GetGroupsUserBelongsTo(username string) ([]string, error) {
	groups, err := oc.userClientV1.Groups().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var groupsUserBelongsTo []string
	for _, group := range groups.Items {
		for _, user := range group.Users {
			if user == username {
				groupsUserBelongsTo = append(groupsUserBelongsTo, group.Name)
			}
		}
	}
	return groupsUserBelongsTo, nil
}
