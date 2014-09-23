package buildlog

import (
	"fmt"
	// "net/http"
	// "regexp"
	// "strings"
	"io/ioutil"
	// "encoding/json"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/pod"
	// "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/openshift/origin/pkg/build/registry/build"

	// kubeapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	// "github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	// "github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	// "github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	// "github.com/openshift/origin/pkg/build/api"
	// "github.com/openshift/origin/pkg/build/api/validation"
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

// Implement Redirector.
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
	err = ioutil.WriteFile("/tmp/redirector",[]byte(location), 0644)
	if err != nil {
		return "",	err
	}
	return location, nil
}

func (storage *Storage) Get(id string) (interface{}, error) {

	// redirector := storage.(apiserver.Redirector)
	// location, err := redirector.ResourceLocation(id)

	// // redirector, ok := storage.(Redirector)

	redirector := apiserver.Redirector(storage)

	logLocation, _ := redirector.ResourceLocation(id)

	return logLocation, nil
	// return fmt.Errorf("BuildLog can only be retrieved"), nil
}

func (storage *Storage) New() interface{} {
	return nil
}

func (storage *Storage) List(selector labels.Selector) (interface{}, error) {
	return fmt.Errorf("BuildLog can only be retrieved"), nil	
}

func (storage *Storage) Delete(id string) (<-chan interface{}, error) {
	return apiserver.MakeAsync(func() (interface{}, error) {
		return nil,	fmt.Errorf("BuildLog can only be retrieved")
	}), nil
}

func (storage *Storage) Extract(body []byte) (interface{}, error) {
	return fmt.Errorf("BuildLog can only be retrieved"), nil
}

func (storage *Storage) Create(obj interface{}) (<-chan interface{}, error) {
	return apiserver.MakeAsync(func() (interface{}, error) {
		return nil,	fmt.Errorf("BuildLog can only be retrieved")
	}), nil
}

func (storage *Storage) Update(obj interface{}) (<-chan interface{}, error) {
	return apiserver.MakeAsync(func() (interface{}, error) {
		return nil,	fmt.Errorf("BuildLog can only be retrieved")
	}), nil
}
