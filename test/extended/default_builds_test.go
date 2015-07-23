// +build default

package extended

import (
	"fmt"
	"path/filepath"
	"os"

	. "github.com/GoogleCloudPlatform/kubernetes/test/e2e"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = Describe("STI environment Build", func() {
	defer GinkgoRecover()

	var imageStreamFixture = filepath.Join("..", "integration", "fixtures", "test-image-stream.json")
	var stiEnvBuildFixture = filepath.Join("fixtures", "test-env-build.json")
	// var stiEnvPodFixture   = filepath.Join("fixtures", "test-env-pod.json")
	var oc = exutil.NewCLI("mysql-create", adminKubeConfigPath(), true)

	// var kc = oc.AdminKubeRESTClient()

	if _, err := os.Stat(imageStreamFixture); os.IsNotExist(err) {
	    fmt.Printf("no such file or directory: %s", imageStreamFixture)
	    return
	}

	Describe("Creating from a build", func(){

		It(fmt.Sprintf("should create image-streams from %q template", imageStreamFixture), func(){

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
		})
	})
})