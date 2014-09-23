package buildlog

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/pod"
	"github.com/openshift/origin/pkg/build/registry/build"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
)

// REST is an implementation of RESTStorage for the api server.
type REST struct {
	BuildRegistry build.Registry
	PodRegistry   pod.Registry
}

// NewREST creates a new REST for BuildLog
func NewREST(b build.Registry, p pod.Registry) apiserver.RESTStorage {
	return &REST{
		BuildRegistry: b,
		PodRegistry: p,
	}
}

// Redirector implementation
func (r *REST) ResourceLocation(id string) (string, error) {
	build, err := r.BuildRegistry.GetBuild(id)
	if err != nil {
		return "",	fmt.Errorf("No such build")
	}

	pod, err := r.PodRegistry.GetPod(build.PodID)
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

func (r *REST) Get(id string) (runtime.Object, error) {

	return nil, nil	
}

func (r *REST) New() runtime.Object {
	return nil
}

func (r *REST) List(selector, fields labels.Selector) (runtime.Object, error) {
	return nil, nil	
}

func (r *REST) Delete(id string) (<-chan runtime.Object, error) {
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return nil,	nil
	}), nil
}

func (r *REST) Create(obj runtime.Object) (<-chan runtime.Object, error) {
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return nil,	nil
	}), nil
}

func (r *REST) Update(obj runtime.Object) (<-chan runtime.Object, error) {
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return nil,	nil
	}), nil
}
