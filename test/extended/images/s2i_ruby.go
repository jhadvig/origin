package images

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/origin/test/extended/util"
)

const (
	modifyRailsSedScript    = `s%public/index.html%app/views/welcome/index.html.erb%`
	modifyRailsFile         = "app/controllers/welcome_controller.rb"
	removeFile              = "public/index.html"
)

func modifyRailsSourceCode(oc *exutil.CLI) {
	pods, err := exutil.WaitForPods(oc.KubeREST().Pods(oc.Namespace()), exutil.ParseLabelsOrDie(fmt.Sprintf("deployment=rails-postgresql-example-1")), 1, 120*time.Second)
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(pods).Should(o.HaveLen(1))

	pod, err := oc.KubeREST().Pods(oc.Namespace()).Get(pods[0])
	o.Expect(err).NotTo(o.HaveOccurred())
	oc.Run("exec").Args(pod.Name, "-c", pod.Spec.Containers[0].Name, "--", "sed", "-ie", modifyRailsSedScript, modifyRailsFile).Execute()
	oc.Run("exec").Args(pod.Name, "-c", pod.Spec.Containers[0].Name, "--", "rm", removeFile).Execute()
}

func assertPageContent(oc *exutil.CLI, content string) {
	address, err := exutil.GetEndpointAddress(oc, "rails-postgresql-example")
	o.Expect(err).NotTo(o.HaveOccurred())

	// Make the request
	resp, err := http.Get(fmt.Sprintf("http://%s/", address))
	o.Expect(err).NotTo(o.HaveOccurred())

	// Read response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	o.Expect(err).NotTo(o.HaveOccurred())

	// Verify returned count
	cond := strings.Contains(string(body), content)
	o.Expect(cond).Should(o.BeTrue())
}

func assertPageStatusCode(oc *exutil.CLI, code int) {
	address, err := exutil.GetEndpointAddress(oc, "rails-postgresql-example")
	resp, err := http.Get(fmt.Sprintf("http://%s/", address))
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(resp.StatusCode).Should(o.Equal(code))
	defer resp.Body.Close()
}

var _ = g.Describe("images: s2i: Ruby", func() {
	defer g.GinkgoRecover()
	var (
		railsTemplate = "https://raw.githubusercontent.com/openshift/rails-ex/master/openshift/templates/rails-postgresql.json"
		oc             = exutil.NewCLI("s2i-ruby", exutil.KubeConfigPath())
	)
	g.Describe("Rails example", func() {
		g.It(fmt.Sprintf("should work with hot deploy"), func() {
			oc.SetOutputDir(exutil.TestContext.OutputDir)

			g.By(fmt.Sprintf("calling oc new-app -f %q", railsTemplate))
			err := oc.Run("new-app").Args("-f", railsTemplate).Execute()
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("waiting for endpoint")
			err = oc.KubeFramework().WaitForAnEndpoint("rails-postgresql-example")
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("modifying the source code with disabled hot deploy")
			modifyRailsSourceCode(oc)
			assertPageStatusCode(oc, 500)

			g.By("turning on hot-deploy")
			err = oc.Run("env").Args("rc", "rails-postgresql-example-1", "RAILS_ENV=development").Execute()
			o.Expect(err).NotTo(o.HaveOccurred())
			err = oc.Run("scale").Args("rc", "rails-postgresql-example-1", "--replicas=0").Execute()
			o.Expect(err).NotTo(o.HaveOccurred())
			err = oc.Run("scale").Args("rc", "rails-postgresql-example-1", "--replicas=1").Execute()
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("modifying the source code with enabled hot deploy")
			modifyRailsSourceCode(oc)
			assertPageStatusCode(oc, 200)
			assertPageContent(oc, "Hello, Rails!")
		})
	})
})
