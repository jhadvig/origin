// +build default

package extended

import (
	"fmt"
	"path/filepath"
	"encoding/json"

	. "github.com/GoogleCloudPlatform/kubernetes/test/e2e"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = Describe("STI environment Build", func() {
	defer GinkgoRecover()

	var imageStreamFixture = filepath.Join("..", "integration", "fixtures", "test-image-stream.json")
	var stiEnvBuildFixture = filepath.Join("fixtures", "test-env-build.json")
	var oc = exutil.NewCLI("build-sti-env", adminKubeConfigPath(), true)

	Describe("Creating from a build", func(){
		var outputPath string

		It(fmt.Sprintf("should create a running pod from %q template", stiEnvBuildFixture), func(){

			By(fmt.Sprintf("calling oc create -f %q", imageStreamFixture))
			if err := oc.Run("create").Args("-f", imageStreamFixture).Verbose().Execute(); err != nil {
				Failf("Could not create image-streams %q: %v", imageStreamFixture, err)
			}

			By(fmt.Sprintf("calling oc create -f %q", stiEnvBuildFixture))
			if err := oc.Run("create").Args("-f", stiEnvBuildFixture).Verbose().Execute(); err != nil {
				Failf("Could not create build %q: %v", stiEnvBuildFixture, err)
			}

			By("starting a test build")
			buildName, err := oc.Run("start-build").Args("test").Output()
			if err != nil {
				Failf("Unable to start build: %v", err)
			}

			By("expecting the build is in Complete phase")
			Expect(oc.OsFramework().WaitForABuild(buildName)).NotTo(HaveOccurred())

			By("creating and writing pod object")
			pod, err := oc.OsFramework().CreatePodObjectForImageStream("test")
			if err != nil {
				Failf("Unable to create and write pod spec: %v", err)
			}

			By(fmt.Sprintf("writing the pod object to %q", outputPath))
			podJSON, err := json.Marshal(pod)
			outputPath, err := writeTempJSON(oc.Namespace(), string(podJSON))
			if err != nil {
				Failf("Couldn't write to %q: %v", outputPath, err)
			}

			By(fmt.Sprintf("calling oc create -f %q", outputPath))
			if err := oc.Run("create").Args("-f", outputPath).Verbose().Execute(); err != nil {
				Failf("Unable to create pod: %v", err)
			}

			By("expecting the pod to be running")
			Expect(oc.KubeFramework().WaitForPodRunning(pod.Name)).NotTo(HaveOccurred())
		})
	})
})