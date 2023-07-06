package K8sClient

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func createKubectlGetGroupsResponse(groupResource []GroupResource) *kubectlGetAllGroupsResp {
	return &kubectlGetAllGroupsResp{
		ApiVersion: "",
		Kind:       "",
		Items:      groupResource,
	}
}

func Test_extractAllTheGroupsTheUserIsIn(t *testing.T) {
	t.Run("should return group1 when the user is in group 1", func(t *testing.T) {
		kubectlGetGroupsResponse := createKubectlGetGroupsResponse([]GroupResource{
			{
				Users: []string{"user1", "user2"},
				Metadata: struct {
					Name string `json:"name"`
				}{
					Name: "group1",
				},
			},
			{
				Users: []string{"user3", "user2"},
				Metadata: struct {
					Name string `json:"name"`
				}{
					Name: "group2",
				},
			},
		})

		assert.Equal(t, []string{"group1"}, extractAllTheGroupsTheUserIsIn("user1", kubectlGetGroupsResponse))
	})

	t.Run("should return group1 and group2 when the user is in group 1 and group 2", func(t *testing.T) {
		kubectlGetGroupsResponse := createKubectlGetGroupsResponse([]GroupResource{
			{
				Users: []string{"user1", "user2"},
				Metadata: struct {
					Name string `json:"name"`
				}{
					Name: "group1",
				},
			},
			{
				Users: []string{"user1", "user2"},
				Metadata: struct {
					Name string `json:"name"`
				}{
					Name: "group2",
				},
			},
		})

		assert.Equal(t, []string{"group1", "group2"}, extractAllTheGroupsTheUserIsIn("user1", kubectlGetGroupsResponse))
	})

	t.Run("should return an empty array when the user is in non of the groups", func(t *testing.T) {
		kubectlGetGroupsResponse := createKubectlGetGroupsResponse([]GroupResource{
			{
				Users: []string{},
				Metadata: struct {
					Name string `json:"name"`
				}{
					Name: "group1",
				},
			},
			{
				Users: []string{"user3", "user2"},
				Metadata: struct {
					Name string `json:"name"`
				}{
					Name: "group2",
				},
			},
		})

		assert.Equal(t, []string{}, extractAllTheGroupsTheUserIsIn("user1", kubectlGetGroupsResponse))

	})

	t.Run("should return an empty array when the api returns 0 groups", func(t *testing.T) {
		kubectlGetGroupsResponse := createKubectlGetGroupsResponse([]GroupResource{})
		assert.Equal(t, []string{}, extractAllTheGroupsTheUserIsIn("user1", kubectlGetGroupsResponse))
	})
}
