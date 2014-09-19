package buildlog

import (
	"testing"

	kubeapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	_ "github.com/GoogleCloudPlatform/kubernetes/pkg/api/v1beta1"
	"github.com/openshift/origin/pkg/build/api"
	_ "github.com/openshift/origin/pkg/build/api/v1beta1"
	"github.com/openshift/origin/pkg/build/registry/test"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/etcd"
	// "github.com/GoogleCloudPlatform/kubernetes/pkg/registry/pod"
)

func mockBuild() *api.Build {
	return &api.Build{
		JSONBase: kubeapi.JSONBase{
			ID: "build-id",
		},
		Input: api.BuildInput{
			Type:      api.DockerBuildType,
			SourceURI: "http://my.build.com/the/build/Dockerfile",
			ImageTag:  "repository/dataBuild",
		},
		Status: api.BuildPending,
		PodID:  "pod-id",
		Labels: map[string]string{"name": "dataBuild"},
	}
}

func mockPod() *kubeapi.Pod {
	return &kubeapi.Pod{
		JSONBase: kubeapi.JSONBase{
			ID: "pod-id",
		},
		Labels:   map[string]string{"a": "b"},
	}
}

func TestGetBuildLog(t *testing.T) {

	expectedPod   := mockPod()
	expectedBuild := mockBuild()

	buildId := "test-build-id"

	etcd.NewRegistry()

	mockBuildRegistry := test.BuildRegistry{Build: expectedBuild}
	mockPodRegistry := test.PodRegistry{Build: expectedBuild}

	
	// mockRegistry := test.BuildLogRegistry{
	// 	Build: expectedBuild,
	// 	Pod: expectedPod,
	// }

	storage := Storage{
		BuildRegistry: &mockBuildRegistry,
		PodRegistry:   &mockPodRegistry,
	}


	buildLogObj, err := storage.Get(buildId)
	if err != nil {
		t.Errorf("Unexpected error returned: %v", err)
	}
	buildLog, ok := buildLogObj.(*api.BuildLog)
	if !ok {
		t.Errorf("A build log was not returned: %v", buildLogObj)
	}
}

// func TestGetBuildError(t *testing.T) {
// 	mockRegistry := test.BuildRegistry{Err: fmt.Errorf("get error")}
// 	storage := Storage{registry: &mockRegistry}
// 	buildObj, err := storage.Get("foo")
// 	if err != mockRegistry.Err {
// 		t.Errorf("Expected %#v, Got %#v", mockRegistry.Err, err)
// 	}
// 	if buildObj != nil {
// 		t.Errorf("Unexpected non-nil build: %#v", buildObj)
// 	}
// }