// +build default

package extended

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/GoogleCloudPlatform/kubernetes/test/e2e"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = Describe("MySQL ephemeral template", func() {
	defer GinkgoRecover()

	var templatePath = filepath.Join("..", "..", "examples", "db-templates", "mysql-ephemeral-template.json")
	var oc = exutil.NewCLI("mysql-create")

	Describe("Creating from a template", func() {
		var outputPath string

		It(fmt.Sprintf("should process and create the %q template", templatePath), func() {
			By(fmt.Sprintf("calling oc process -f %q", templatePath))
			templateOutput, err := oc.Run("process").Args("-f", templatePath).Output()
			if err != nil {
				Failf("Couldn't process template %q: %v", templatePath, err)
			}

			By("writing output from process to a file")
			outputPath = filepath.Join(os.TempDir(), oc.Namespace()+".json")
			err = ioutil.WriteFile(outputPath, []byte(templateOutput), 0644)
			if err != nil {
				Failf("Couldn't write to %q: %v", outputPath, err)
			}

			By(fmt.Sprintf("calling oc create -f %q", outputPath))
			if err := oc.Run("create").Args("-f", outputPath).Verbose().Execute(); err != nil {
				Failf("Unable to create from list: %v", err)
			}

			By("waiting for an mysql endpoint")
			Expect(oc.Framework.WaitForAnEndpoint("mysql")).NotTo(HaveOccurred())
		})
	})

})
