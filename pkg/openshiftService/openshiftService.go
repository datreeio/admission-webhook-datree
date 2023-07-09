package openshiftService

import (
	userClientV1Api "github.com/openshift/api/user/v1"
	"github.com/patrickmn/go-cache"
	"time"
)

type IOpenshiftClient interface {
	getGroups() (*userClientV1Api.GroupList, error)
}

type OpenshiftService struct {
	openshiftClient IOpenshiftClient
	cache           *cache.Cache
}

const groupsCacheKey = "groups"

func NewOpenshiftService() (*OpenshiftService, error) {
	openshiftClient, err := NewOpenshiftClient()
	if err != nil {
		return nil, err
	}
	return &OpenshiftService{
		openshiftClient: openshiftClient,
		cache:           cache.New(1*time.Minute, 10*time.Minute),
	}, nil
}

func (oc *OpenshiftService) GetGroupsUserBelongsTo(username string) ([]string, error) {
	groups, err := oc.getGroupsWithCache()
	if err != nil {
		return nil, err
	}

	groupsUserBelongsTo := make([]string, 0)
	for _, group := range groups.Items {
		for _, user := range group.Users {
			if user == username {
				groupsUserBelongsTo = append(groupsUserBelongsTo, group.Name)
				break
			}
		}
	}

	return groupsUserBelongsTo, nil
}

func (oc *OpenshiftService) getGroupsWithCache() (*userClientV1Api.GroupList, error) {
	valueFromCache, found := oc.cache.Get(groupsCacheKey)
	if found {
		return valueFromCache.(*userClientV1Api.GroupList), nil
	}
	groups, err := oc.openshiftClient.getGroups()
	if err != nil {
		return nil, err
	}
	oc.cache.Set(groupsCacheKey, groups, 1*time.Minute)
	return groups, nil
}
