package openshiftService

import (
	userClientV1Api "github.com/openshift/api/user/v1"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

type mockOpenshiftClient struct {
	mock.Mock
}

const getGroupsMethodName = "getGroups"

func (m *mockOpenshiftClient) getGroups() (*userClientV1Api.GroupList, error) {
	args := m.Called()
	return args.Get(0).(*userClientV1Api.GroupList), args.Error(1)
}

func createMockOpenshiftClient(getGroupsResponse *userClientV1Api.GroupList) *mockOpenshiftClient {
	mockOpenshiftClient := &mockOpenshiftClient{}
	mockOpenshiftClient.On(getGroupsMethodName).Return(getGroupsResponse, nil)
	return mockOpenshiftClient
}

func Test_GetGroupsUserBelongsTo(t *testing.T) {
	t.Run("should return the group when the user is in 1 group", func(t *testing.T) {
		// arrange
		openshiftService := OpenshiftService{
			openshiftClient: createMockOpenshiftClient(&userClientV1Api.GroupList{
				Items: []userClientV1Api.Group{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "group1",
						},
						Users: []string{
							"user1",
							"user2",
						},
					},
				},
			}),
			cache: cache.New(1*time.Minute, 10*time.Minute),
		}

		// act
		groupsUserBelongsTo, _ := openshiftService.GetGroupsUserBelongsTo("user1")

		// assert
		assert.Equal(t, []string{"group1"}, groupsUserBelongsTo)
	})

	t.Run("should return an emtpy array when the user is in 0 groups", func(t *testing.T) {
		// arrange
		openshiftService := OpenshiftService{
			openshiftClient: createMockOpenshiftClient(&userClientV1Api.GroupList{
				Items: []userClientV1Api.Group{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "group1",
						},
						Users: []string{
							"user1",
							"user2",
						},
					},
				},
			}),
			cache: cache.New(1*time.Minute, 10*time.Minute),
		}

		// act
		groupsUserBelongsTo, _ := openshiftService.GetGroupsUserBelongsTo("user3")

		// assert
		assert.Equal(t, []string{}, groupsUserBelongsTo)
	})

	t.Run("should use value from cache when called twice in the same minute, avoid spamming the api-server", func(t *testing.T) {
		// arrange
		mockOpenshiftClient := createMockOpenshiftClient(&userClientV1Api.GroupList{
			Items: []userClientV1Api.Group{},
		})
		openshiftService := OpenshiftService{
			openshiftClient: mockOpenshiftClient,
			cache:           cache.New(1*time.Minute, 10*time.Minute),
		}

		// act
		groupsUserBelongsTo1, _ := openshiftService.GetGroupsUserBelongsTo("user1")
		groupsUserBelongsTo2, _ := openshiftService.GetGroupsUserBelongsTo("user1")

		// assert
		mockOpenshiftClient.AssertNumberOfCalls(t, getGroupsMethodName, 1)
		assert.Equal(t, groupsUserBelongsTo1, groupsUserBelongsTo2)
	})

	t.Run("should return all the groups the user belongs to, complex case", func(t *testing.T) {
		// arrange
		openshiftService := OpenshiftService{
			openshiftClient: createMockOpenshiftClient(&userClientV1Api.GroupList{
				Items: []userClientV1Api.Group{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "group1",
						},
						Users: []string{
							"user1",
							"user2",
							"user1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "group2",
						},
						Users: []string{
							"user2",
							"user3",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "group3",
						},
						Users: []string{
							"user2",
							"user1",
						},
					},
				},
			}),
			cache: cache.New(1*time.Minute, 10*time.Minute),
		}

		// act
		groupsUserBelongsTo, _ := openshiftService.GetGroupsUserBelongsTo("user1")

		// assert
		assert.Equal(t, []string{"group1", "group3"}, groupsUserBelongsTo)
	})
}
