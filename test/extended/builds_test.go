package extended

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
)

func init() {
	RequireServerVars()
}

var (
	imageStreamFixture = filepath.Join("..", "integration", "fixtures", "test-image-stream.json")
	stiEnvBuildFixture = filepath.Join("fixtures", "test-env-build.json")
	stiEnvPodFixture   = filepath.Join("fixtures", "test-env-pod.json")
)

// TestSTIEnvironmentBuild exercises the scenario where you have .sti/environment
// file in your source code repository and you use STI build strategy. In that
// case the STI build should read that file and set all environment variables
// from that file to output image.
func TestSTIEnvironmentBuild(t *testing.T) {
	oc := NewCLI("build-sti-env").Verbose()

	// Create imageStream used in this test
	if err := oc.Run("create").Args("-f", imageStreamFixture).Execute(); err != nil {
		t.Fatalf("Error creating imageStream: %v", err)
	}
	defer oc.Run("delete").Args("imageStream", "test").Execute()

	// Create watcher to watch the build
	buildWatcher, err := oc.AdminRESTClient().Builds(oc.Namespace()).
		Watch(labels.Everything(), fields.Everything(), "0")
	if err != nil {
		t.Fatalf("Unable to create watcher for builds: %v", err)
	}
	defer buildWatcher.Stop()

	// Create buildConfig and start the build manually
	if err := oc.Run("create").Args("-f", stiEnvBuildFixture).Execute(); err != nil {
		t.Fatalf("Error creating build: %v", err)
	}
	defer oc.Run("delete").Args("buildConfig", "test").Execute()

	buildName, err := oc.Run("start-build").Args("test").Verbose().Output()
	if err != nil {
		t.Fatalf("Unable to start build: %v", err)
	}

	if err := waitForBuildComplete(buildName, buildWatcher); err != nil {
		logs, _ := oc.Run("build-logs").Args(buildName, "--nowait").Output()
		t.Fatalf("Build error: %v\n%s\n", err, logs)
	}

	// Verification:

	podWatcher, err := oc.AdminKubeRESTClient().Pods(oc.Namespace()).
		Watch(labels.Everything(), fields.Everything(), "0")
	if err != nil {
		t.Fatalf("Unable to create watcher for pods: %v", err)
	}
	defer podWatcher.Stop()

	// Run the pod with the built image and verify it's content
	podName, err := createPodForImageStream(oc, "test")
	if err != nil {
		t.Fatalf("Unable to create pod for verification: %v", err)
	}
	defer oc.Run("delete").Args("pod", podName).Execute()

	if err := waitForPodRunning(podName, podWatcher); err != nil {
		logs, _ := oc.Run("logs").Args("-p", podName).Output()
		t.Fatalf("Pod error: %v\n%s\n", err, logs)
	}

	result, err := oc.Run("exec").
		Args(podName, "--", "curl", "http://localhost:8080").
		Verbose().
		Output()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "success") {
		t.Errorf("Expected TEST_ENV contains 'success', got: %q", result)
	}
}
