// +build default

package extended

import (
	"fmt"
	"path/filepath"

	. "github.com/GoogleCloudPlatform/kubernetes/test/e2e"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = Describe("MySQL ephemeral template", func() {
	defer GinkgoRecover()

	var templatePath = filepath.Join("..", "..", "examples", "db-templates", "mysql-ephemeral-template.json")
	var oc = exutil.NewGinkoCLI("mysql-create", kubeConfigPath())

	Describe("Creating from a template", func() {

		It(fmt.Sprintf("should process and create the %q template", templatePath), func() {
			outputPath := getTempFilePath(oc.Namespace())

			By(fmt.Sprintf("calling oc process -f %q", templatePath))
			templateOutput, err := oc.Run("process").Args("-f", templatePath).Output()
			if err != nil {
				Failf("Couldn't process template %q: %v", templatePath, err)
			}

			By(fmt.Sprintf("by writing the output to %q", outputPath))
			err = writeTempJSON(outputPath, templateOutput)
			if err != nil {
				Failf("Couldn't write to %q: %v", outputPath, err)
			}

			By(fmt.Sprintf("calling oc create -f %q", outputPath))
			if err := oc.Run("create").Args("-f", outputPath).Verbose().Execute(); err != nil {
				Failf("Unable to create from list: %v", err)
			}

			By("expecting the mysql service get endpoints")
			Expect(oc.KubeFramework().WaitForAnEndpoint("mysql")).NotTo(HaveOccurred())
		})
	})

})
