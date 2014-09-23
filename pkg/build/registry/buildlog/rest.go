package buildlog

import (
	"fmt"

	"github.com/openshift/origin/pkg/build/registry/build"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
)

// REST is an implementation of RESTStorage for the api server.
type REST struct {
	BuildRegistry build.Registry
	PodClient     client.PodInterface
	MasterAddr    string
}

// NewREST creates a new REST for BuildLog
func NewREST(b build.Registry, c client.PodInterface, a string) apiserver.RESTStorage {
	return &REST{
		BuildRegistry: b,
		PodClient: c,
		MasterAddr: a,
	}
}

// Redirector implementation
func (r *REST) ResourceLocation(id string) (string, error) {
	build, err := r.BuildRegistry.GetBuild(id)
	if err != nil {
		return "", fmt.Errorf("No such build")
	}

	pod, err := r.PodClient.GetPod(build.PodID)
	if err != nil {
		return "", fmt.Errorf("No such pod")
	}

	buildPodID := build.PodID
	buildHost  := pod.CurrentState.Host
	// Build will take place only in one container 
	buildContainerName := pod.DesiredState.Manifest.Containers[0].Name
	location := fmt.Sprintf("http://%s/proxy/minion/%s/containerLogs/%s/%s",r.MasterAddr, buildHost, buildPodID, buildContainerName)

	return location, nil
}

func (r *REST) Get(id string) (runtime.Object, error) {
	return nil, fmt.Errorf("BuildLog can't be retrieved")
}

func (r *REST) New() runtime.Object {
	return nil
}

func (r *REST) List(selector, fields labels.Selector) (runtime.Object, error) {
	return nil, fmt.Errorf("BuildLog can't be listed")
}

func (r *REST) Delete(id string) (<-chan runtime.Object, error) {
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return nil,	nil
	}), fmt.Errorf("BuildLog can't be deleted")
}

func (r *REST) Create(obj runtime.Object) (<-chan runtime.Object, error) {
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return nil,	nil
	}), fmt.Errorf("BuildLog can't be created")
}

func (r *REST) Update(obj runtime.Object) (<-chan runtime.Object, error) {
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return nil,	nil
	}), fmt.Errorf("BuildLog can't be updated")
}
