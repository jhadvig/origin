package buildlog

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"io/ioutil"
	"encoding/json"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/pod"
	// "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/openshift/origin/pkg/build/registry/build"

	// kubeapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	// "github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	// "github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/openshift/origin/pkg/build/api"
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

// var logRegexp = regexp.MustCompile(`.*\[([A-Z][a-z]{2}\s+[0-3]{0,1}[0-9]T[0-2][0-9]:[0-5][0-9]:[0-5][0-9].[0-9]{0,}Z})\]\s+(.*)`)
var logRegexp = regexp.MustCompile(`.*([0-9]{4}-[0-1][0-9]-[0-3][0-9]T[0-2][0-9]:[0-5][0-9]:[0-5][0-9].[0-9]{0,}Z)\s+(.*)`)

func (storage *Storage) Get(id string) (interface{}, error) {

	build, err := storage.BuildRegistry.GetBuild(id)
	if err != nil {
		return nil,	fmt.Errorf("No such build")
	}

	pod, err := storage.PodRegistry.GetPod(build.PodID)
	if err != nil {
		return nil,	fmt.Errorf("No such pod")
	}

	buildPodID := build.PodID
	buildHost  := pod.CurrentState.Host
	// Build will take place only in one container 
	buildContainerName := pod.DesiredState.Manifest.Containers[0].Name

	// podInfoGetter := client.HTTPPodInfoGetter{
	// 	Client: http.DefaultClient,
	// 	Port:   10250,
	// }
	// // get hostname:hostport ?
	// _, err = podInfoGetter.GetPodInfo(buildHost, buildPodID)

	client := &http.DefaultClient

	req, err := http.NewRequest(
		"GET", 
		fmt.Sprintf("http://127.0.0.1:8080/proxy/minion/%s/containerLogs/%s/%s", buildHost, buildPodID, buildContainerName), 
		nil,
	)

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	buildLog := &api.BuildLog{}

	logLines, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil,	err
	}

	for _, line := range strings.Split(string(logLines), "\n") {
		if len(line) > 0 {
			matches := logRegexp.FindStringSubmatch(line)
			buildLog.LogItems = append(buildLog.LogItems, api.LogItem{Timestamp:matches[1], Log:matches[2]})
		}
	}

	buildLog.CreationTimestamp = util.Now()

	test, _ := json.Marshal(buildLog)
	err = ioutil.WriteFile("/tmp/buildLog", test, 0644)
	if err != nil {
		return &buildLog, nil
	}

	return buildLog, nil
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
	return nil,	fmt.Errorf("BuildLog can only be retrieved")
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
