package buildlog

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/pod"
	"github.com/openshift/origin/pkg/build/registry/build"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
)

// Storage is an implementation of RESTStorage for the api server.
type Storage struct {
	BuildRegistry build.Registry
	PodRegistry   pod.Registry
}

// NewStorage creates a new Storage for BuildLog
func NewStorage(b build.Registry, p pod.Registry) apiserver.RESTStorage {
	return &Storage{
		BuildRegistry: b,
		PodRegistry: p,
	}
}

// Redirector implementation
func (storage *Storage) ResourceLocation(id string) (string, error) {
	build, err := storage.BuildRegistry.GetBuild(id)
	if err != nil {
		return "",	fmt.Errorf("No such build")
	}

	pod, err := storage.PodRegistry.GetPod(build.PodID)
	if err != nil {
		return "",	fmt.Errorf("No such pod")
	}

	buildPodID := build.PodID
	buildHost  := pod.CurrentState.Host
	// Build will take place only in one container 
	buildContainerName := pod.DesiredState.Manifest.Containers[0].Name
	location := fmt.Sprintf("http://127.0.0.1:8080/proxy/minion/%s/containerLogs/%s/%s", buildHost, buildPodID, buildContainerName)

	return location, nil
}

func (storage *Storage) Get(id string) (interface{}, error) {

	return nil, nil	
}

func (storage *Storage) New() interface{} {
	return nil
}

func (storage *Storage) List(selector labels.Selector) (interface{}, error) {
	return nil, nil	
}

func (storage *Storage) Delete(id string) (<-chan interface{}, error) {
	return apiserver.MakeAsync(func() (interface{}, error) {
		return nil,	nil
	}), nil
}

func (storage *Storage) Create(obj interface{}) (<-chan interface{}, error) {
	return apiserver.MakeAsync(func() (interface{}, error) {
		return nil,	nil
	}), nil
}

func (storage *Storage) Update(obj interface{}) (<-chan interface{}, error) {
	return apiserver.MakeAsync(func() (interface{}, error) {
		return nil,	nil
	}), nil
}