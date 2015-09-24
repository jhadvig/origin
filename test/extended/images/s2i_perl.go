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
	modifySedScript = `s/data => \$data\[0\]/data => "1337"/`
	modifyFile      = "lib/default.pm"
)

// modifySourceCode will modify source code in the pod so the application
// according to the sed script.
func modifySourceCode(oc *exutil.CLI) {
	pods, err := exutil.WaitForPods(oc.KubeREST().Pods(oc.Namespace()), exutil.ParseLabelsOrDie(fmt.Sprintf("deployment=dancer-mysql-example-1")), 1, 120*time.Second)
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(pods).Should(o.HaveLen(1))

	pod, err := oc.KubeREST().Pods(oc.Namespace()).Get(pods[0])
	o.Expect(err).NotTo(o.HaveOccurred())
	oc.Run("exec").Args(pod.Name, "-c", pod.Spec.Containers[0].Name, "--", "sed", "-ie", modifySedScript, modifyFile).Execute()
}

// Make a ginkgo assertions about page count returned via a http request from
// the application frontend.
func assertPageCountIs(oc *exutil.CLI, count int) {
	address, err := exutil.GetEndpointAddress(oc, "dancer-mysql-example")
	o.Expect(err).NotTo(o.HaveOccurred())

	// Make the request
	resp, err := http.Get(fmt.Sprintf("http://%s/", address))
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(resp.StatusCode).Should(o.Equal(200))

	// Read response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	o.Expect(err).NotTo(o.HaveOccurred())

	// Verify returned count
	cond := strings.Contains(string(body), fmt.Sprintf("<span class=\"code\" id=\"count-value\">%d</span>", count))
	o.Expect(cond).Should(o.BeTrue())
}

var _ = g.Describe("images: s2i: Perl", func() {
	defer g.GinkgoRecover()
	var (
		dancerTemplate = "https://raw.githubusercontent.com/openshift/dancer-ex/master/openshift/templates/dancer-mysql.json"
		oc             = exutil.NewCLI("s2i-perl", exutil.KubeConfigPath())
	)
	g.Describe("Dancer example", func() {
		g.It(fmt.Sprintf("should work with hot deploy"), func() {
			oc.SetOutputDir(exutil.TestContext.OutputDir)

			g.By(fmt.Sprintf("calling oc new-app -f %q", dancerTemplate))
			err := oc.Run("new-app").Args("-f", dancerTemplate).Execute()
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("waiting for endpoint")
			err = oc.KubeFramework().WaitForAnEndpoint("dancer-mysql-example")
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("checking page count")
			assertPageCountIs(oc, 1)
			assertPageCountIs(oc, 2)

			g.By("modifying the source code with disabled hot deploy")
			modifySourceCode(oc)
			assertPageCountIs(oc, 3)

			g.By("turning on hot-deploy")
			err = oc.Run("env").Args("rc", "dancer-mysql-example-1", "PERL_APACHE2_RELOAD=true").Execute()
			o.Expect(err).NotTo(o.HaveOccurred())
			err = oc.Run("scale").Args("rc", "dancer-mysql-example-1", "--replicas=0").Execute()
			o.Expect(err).NotTo(o.HaveOccurred())
			err = oc.Run("scale").Args("rc", "dancer-mysql-example-1", "--replicas=1").Execute()
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("modifying the source code with enabled hot deploy")
			modifySourceCode(oc)
			assertPageCountIs(oc, 1337)
		})
	})
})
