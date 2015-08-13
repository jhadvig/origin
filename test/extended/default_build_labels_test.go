// +build default

package extended

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	buildapi "github.com/openshift/origin/pkg/build/api"
	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = Describe("Check S2I and Docker build image for proper labels", func() {
	defer GinkgoRecover()
	var (
		imageStreamFixture = filepath.Join("..", "integration", "fixtures", "test-image-stream.json")
		stiBuildFixture = filepath.Join("fixtures", "test-sti-build.json")
		dockerBuildFixture = filepath.Join("fixtures", "test-docker-build.json")
		oc                 = exutil.NewCLI("build-sti-env", kubeConfigPath())
	)

	Describe("S2I build from a template", func() {
		It(fmt.Sprintf("should create a image from %q template with proper Docker labels", stiBuildFixture), func() {
			oc.SetOutputDir(testContext.OutputDir)

			By(fmt.Sprintf("calling oc create -f %q", imageStreamFixture))
			err := oc.Run("create").Args("-f", imageStreamFixture).Execute()
			Expect(err).NotTo(HaveOccurred())

			By(fmt.Sprintf("calling oc create -f %q", stiBuildFixture))
			err = oc.Run("create").Args("-f", stiBuildFixture).Execute()
			Expect(err).NotTo(HaveOccurred())

			By("starting a test build")
			buildName, err := oc.Run("start-build").Args("test").Output()
			Expect(err).NotTo(HaveOccurred())

			By("expecting the S2I build is in Complete phase")
			err = exutil.WaitForABuild(oc.REST().Builds(oc.Namespace()), buildName,
				// The build passed
				func(b *buildapi.Build) bool {
					return b.Name == buildName && b.Status.Phase == buildapi.BuildPhaseComplete
				},
				// The build failed
				func(b *buildapi.Build) bool {
					if b.Name != buildName {
						return false
					}
					return b.Status.Phase == buildapi.BuildPhaseFailed || b.Status.Phase == buildapi.BuildPhaseError
				},
			)
			Expect(err).NotTo(HaveOccurred())

			By("getting the Docker image reference from ImageStream")
			imageRef, err := exutil.GetDockerImageReference(oc.REST().ImageStreams(oc.Namespace()), "test", "latest")
			Expect(err).NotTo(HaveOccurred())

			imageLabels, err := exutil.GetImageLabels(oc.REST().ImageStreamImages(oc.Namespace()), "test", imageRef)
			Expect(err).NotTo(HaveOccurred())

			By("inspecting the new image for proper Docker labels")
			err = testLabels(imageLabels)
			Expect(err).NotTo(HaveOccurred())

		})
	})

	Describe("Docker build from a template", func() {
		It(fmt.Sprintf("should create a image from %q template with proper Docker labels", dockerBuildFixture), func() {
			oc.SetOutputDir(testContext.OutputDir)

			By(fmt.Sprintf("calling oc create -f %q", imageStreamFixture))
			err := oc.Run("create").Args("-f", imageStreamFixture).Execute()
			Expect(err).NotTo(HaveOccurred())

			By(fmt.Sprintf("calling oc create -f %q", dockerBuildFixture))
			err = oc.Run("create").Args("-f", dockerBuildFixture).Execute()
			Expect(err).NotTo(HaveOccurred())

			By("starting a test build")
			buildName, err := oc.Run("start-build").Args("test").Output()
			Expect(err).NotTo(HaveOccurred())

			By("expecting the Docker build is in Complete phase")
			err = exutil.WaitForABuild(oc.REST().Builds(oc.Namespace()), buildName,
				// The build passed
				func(b *buildapi.Build) bool {
					return b.Name == buildName && b.Status.Phase == buildapi.BuildPhaseComplete
				},
				// The build failed
				func(b *buildapi.Build) bool {
					if b.Name != buildName {
						return false
					}
					return b.Status.Phase == buildapi.BuildPhaseFailed || b.Status.Phase == buildapi.BuildPhaseError
				},
			)
			Expect(err).NotTo(HaveOccurred())

			By("getting the Docker image reference from ImageStream")
			imageRef, err := exutil.GetDockerImageReference(oc.REST().ImageStreams(oc.Namespace()), "test", "latest")
			Expect(err).NotTo(HaveOccurred())

			imageLabels, err := exutil.GetImageLabels(oc.REST().ImageStreamImages(oc.Namespace()), "test", imageRef)
			Expect(err).NotTo(HaveOccurred())

			By("inspecting the new image for proper Docker labels")
			err = testLabels(imageLabels)
			Expect(err).NotTo(HaveOccurred())

		})
	})


})

// testLabels tests if builded Docker image contains appropriate
// labels.
func testLabels(labels map[string]string) error {
	testedLabels := []string{
		"io.openshift.build.commit.author",
		"io.openshift.build.commit.date",
		"io.openshift.build.commit.id",
		"io.openshift.build.commit.ref",
		"io.openshift.build.commit.message",
		"io.openshift.build.source-location",
		"io.openshift.build.source-context-dir",
	}

	for _, label := range testedLabels {
		if labels[label] == "" {
			return fmt.Errorf("Builded image doesn't contain proper Docker image labels. Missing %q label", label)
		}
	}

	return nil
}
