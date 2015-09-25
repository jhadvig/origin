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
	modifyCakePHPSedScript = `s/\$result\['c'\]/1337/`
	modifyCakePHPFile      = "app/View/Layouts/default.ctp"
)

func modifyCakePHPSourceCode(oc *exutil.CLI) {
	pods, err := exutil.WaitForPods(oc.KubeREST().Pods(oc.Namespace()), exutil.ParseLabelsOrDie(fmt.Sprintf("deployment=cakephp-mysql-example-1")), 1, 120*time.Second)
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(pods).Should(o.HaveLen(1))

	pod, err := oc.KubeREST().Pods(oc.Namespace()).Get(pods[0])
	o.Expect(err).NotTo(o.HaveOccurred())
	oc.Run("exec").Args(pod.Name, "-c", pod.Spec.Containers[0].Name, "--", "sed", "-ie", modifyCakePHPSedScript, modifyCakePHPFile).Execute()
}

func assertCakePHPPageCountIs(oc *exutil.CLI, count int) {
	address, err := exutil.GetEndpointAddress(oc, "cakephp-mysql-example")
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

var _ = g.Describe("images: s2i: PHP", func() {
	defer g.GinkgoRecover()
	var (
		cakephpTemplate = "https://raw.githubusercontent.com/openshift/cakephp-ex/master/openshift/templates/cakephp-mysql.json"
		oc             = exutil.NewCLI("s2i-php", exutil.KubeConfigPath())
	)
	g.Describe("CakePHP example", func() {
		g.It(fmt.Sprintf("should work with hot deploy"), func() {
			oc.SetOutputDir(exutil.TestContext.OutputDir)

			g.By(fmt.Sprintf("calling oc new-app -f %q", cakephpTemplate))
			err := oc.Run("new-app").Args("-f", cakephpTemplate).Execute()
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("waiting for endpoint")
			err = oc.KubeFramework().WaitForAnEndpoint("cakephp-mysql-example")
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("checking page count")
			assertCakePHPPageCountIs(oc, 1)
			assertCakePHPPageCountIs(oc, 2)

			g.By("modifying the source code to check if the hot deploy work out of box")
			modifyCakePHPSourceCode(oc)
			assertCakePHPPageCountIs(oc, 1337)
		})
	})
})
