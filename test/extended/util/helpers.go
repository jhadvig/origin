package util

import (
	"fmt"
	"math/rand"
	"os"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/watch"
	buildapi "github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/util/namer"
)

// From github.com/GoogleCloudPlatform/kubernetes/pkg/api/generator.go
var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// RequireServerVars verifies that all environment variables required to access
// the OpenShift server are set.
func RequireServerVars() {
	if len(GetMasterAddr()) == 0 {
		FatalErr("The 'MASTER_ADDR' environment variable must be set.")
	}
	if len(GetServerConfigDir()) == 0 {
		FatalErr("The 'SERVER_CONFIG_DIR' environment variable must be set.")
	}
}

// FatalErr exits the test in case a fatal error has occured.
func FatalErr(msg interface{}) {
	fmt.Printf("ERROR: %v\n", msg)
	os.Exit(1)
}

// GetServerConfigDir returns the path to OpenShift server config directory
func GetServerConfigDir() string {
	return os.Getenv("SERVER_CONFIG_DIR")
}

// GetMasterAddr returns the address of OpenShift API server.
func GetMasterAddr() string {
	return os.Getenv("MASTER_ADDR")
}

// WaitForPodRunning waits util a pod specified by name is running.
// If the pod state is Failed an error is returned.
func WaitForPodRunning(podName string, w watch.Interface) error {
	fmt.Printf("Waiting for pod %q ...\n", podName)
	for event := range w.ResultChan() {
		eventPod, ok := event.Object.(*kapi.Pod)
		if !ok {
			return fmt.Errorf("cannot convert input to pod object")
		}
		if podName != eventPod.Name {
			continue
		}
		switch eventPod.Status.Phase {
		case kapi.PodFailed:
			return fmt.Errorf("the pod %q failed: %+v", podName, eventPod.Status)
		case kapi.PodRunning:
			fmt.Printf("Pod %q status is now %q\n", podName, eventPod.Status.Phase)
			return nil
		default:
			fmt.Printf("Pod %q status is now %q\n", podName, eventPod.Status.Phase)
		}
	}
	return fmt.Errorf("unexpected closure of result channel for watcher")
}

// WaitForBuildComplete waits until a build specified by name complete.
// If the build status is Error or Failed an error is returned.
func WaitForBuildComplete(buildName string, w watch.Interface) error {
	fmt.Printf("Waiting for build %q ...\n", buildName)
	for event := range w.ResultChan() {
		eventBuild, ok := event.Object.(*buildapi.Build)
		if !ok {
			return fmt.Errorf("cannot convert input to build object")
		}
		if buildName != eventBuild.Name {
			continue
		}
		switch eventBuild.Status.Phase {
		case buildapi.BuildPhaseFailed, buildapi.BuildPhaseError:
			return fmt.Errorf("the build %q failed: %+v", buildName, eventBuild.Status)
		case buildapi.BuildPhaseComplete:
			fmt.Printf("Build %q status is now %q\n", buildName, eventBuild.Status.Phase)
			return nil
		default:
			fmt.Printf("Build %q status is now %q\n", buildName, eventBuild.Status.Phase)
		}
	}
	return fmt.Errorf("unexpected closure of result channel for watcher")
}

// WaitForEndpoint waits until an endpoint receives an address.
func WaitForEndpoint(endpointName string, w watch.Interface) error {
	fmt.Printf("Waiting for endpoint %q ...\n", endpointName)
	for event := range w.ResultChan() {
		eventEndpoint, ok := event.Object.(*kapi.Endpoints)
		if !ok {
			return fmt.Errorf("cannot covert input to endpoint object")
		}
		if endpointName != eventEndpoint.Name {
			continue
		}
		if len(eventEndpoint.Subsets) != 0 {
			for _, set := range eventEndpoint.Subsets {
				for _, address := range set.Addresses {
					for _, port := range set.Ports {
						fmt.Printf("Endpoint %q received address %s:%s\n", eventEndpoint.Name, address.IP, port)
						return nil
					}
				}
			}
		}
		fmt.Printf("Waiting for endpoint %q to receive address ...\n", eventEndpoint.Name)
	}
	return fmt.Errorf("unexpected closure of result channel for watcher")
}

// CreatePodForImageStream creates sample pod for given imageStream.
// It resolves the dockerImageReference from the image stream.
func CreatePodForImageStream(oc *CLI, imageStream string) (string, error) {
	imageName, err := oc.Run("get").Args("is", imageStream).
		Template("{{ with index .status.tags 0 }}{{ with index .items 0}}{{ .dockerImageReference }}{{ end }}{{ end }}").
		Output()
	if err != nil {
		return "", err
	}
	podName := namer.GetName("test-pod", randSeq(5), util.DNS1123SubdomainMaxLength)
	fmt.Printf("Creating pod %q using %q image ...\n", podName, imageName)
	pod := &kapi.Pod{
		ObjectMeta: kapi.ObjectMeta{
			Name:   podName,
			Labels: map[string]string{"name": podName},
		},
		Spec: kapi.PodSpec{
			ServiceAccountName: "builder",
			Containers: []kapi.Container{
				{
					Name:  "test",
					Image: imageName,
				},
			},
			RestartPolicy: kapi.RestartPolicyNever,
		},
	}
	newPod, err := oc.AdminKubeRESTClient().Pods(oc.Namespace()).Create(pod)
	if err != nil {
		return "", err
	}
	return newPod.Name, nil
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
