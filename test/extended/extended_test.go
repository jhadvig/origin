package extended

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/clientcmd"
	. "github.com/GoogleCloudPlatform/kubernetes/test/e2e"
	"github.com/golang/glog"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	exutil "github.com/openshift/origin/test/extended/util"
)

var (
	reportDir = flag.String("report-dir", "", "Path to the directory where the JUnit XML reports should be saved. Default is empty, which doesn't generate these reports.")
)

func init() {
	exutil.RequireServerVars()

	var testKubeConfig, testKubeCert string

	kubeConfigPath := os.Getenv("SERVER_KUBECONFIG_PATH")
	if len(kubeConfigPath) == 0 {
		testKubeConfig = filepath.Join(os.Getenv("SERVER_CONFIG_DIR"), "master", "admin.kubeconfig")
		testKubeCert = filepath.Join(os.Getenv("SERVER_CONFIG_DIR"), "master")
	} else {
		testKubeConfig = filepath.Join(os.Getenv("SERVER_CONFIG_DIR"), kubeConfigPath)
		testKubeCert = filepath.Join(os.Getenv("SERVER_CERT_DIR"))
	}

	testContext := TestContextType{}

	// Turn on verbose by default to get spec names
	config.DefaultReporterConfig.Verbose = true

	// Turn on EmitSpecProgress to get spec progress (especially on interrupt)
	config.GinkgoConfig.EmitSpecProgress = true

	// Randomize specs as well as suites
	config.GinkgoConfig.RandomizeAllSpecs = false

	flag.StringVar(&testContext.KubeConfig, clientcmd.RecommendedConfigPathFlag, testKubeConfig, "Path to kubeconfig containing embeded authinfo.")
	flag.StringVar(&testContext.KubeContext, clientcmd.FlagContext, "", "kubeconfig context to use/override. If unset, will use value from 'current-context'")
	flag.StringVar(&testContext.CertDir, "cert-dir", testKubeCert, "Path to the directory containing the certs. Default is empty, which doesn't use certs.")
	flag.StringVar(&testContext.Host, "host", os.Getenv("MASTER_ADDR"), "The host, or apiserver, to connect to")
	flag.StringVar(&testContext.KubectlPath, "kubectl-path", "kubectl", "The kubectl binary to use. For development, you might use 'cluster/kubectl.sh' here.")
	flag.StringVar(&testContext.OutputDir, "e2e-output-dir", "/tmp", "Output directory for interesting/useful test data, like performance data, benchmarks, and other metrics.")

	SetTestContext(testContext)
}

func TestExtended(t *testing.T) {
	var r []ginkgo.Reporter

	if *reportDir != "" {
		if err := os.MkdirAll(*reportDir, 0755); err != nil {
			glog.Errorf("Failed creating report directory: %v", err)
		}
		defer CoreDump(*reportDir)
	}

	// Disable density test unless it's explicitly requested.
	if config.GinkgoConfig.FocusString == "" && config.GinkgoConfig.SkipString == "" {
		config.GinkgoConfig.SkipString = "Skipped"
	}
	gomega.RegisterFailHandler(ginkgo.Fail)

	if *reportDir != "" {
		r = append(r, reporters.NewJUnitReporter(path.Join(*reportDir, fmt.Sprintf("junit_%02d.xml", config.GinkgoConfig.ParallelNode))))
	}
  
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "OpenShift extended tests suite", r)
}
