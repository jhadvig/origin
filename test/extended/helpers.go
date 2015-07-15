package extended

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	kclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/watch"
	buildapi "github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/util/namer"
	testutil "github.com/openshift/origin/test/util"
	"github.com/spf13/cobra"
)

var adminConfigPath = filepath.Join(GetServerConfigDir(), "master", "admin.kubeconfig")

type CLI struct {
	namespace         string
	verb              string
	globalArgs        []string
	commandArgs       []string
	finalArgs         []string
	stdout            io.Writer
	verbose           bool
	cmd               *cobra.Command
	adminClient       *client.Client
	adminKubeClient   *kclient.Client
	adminClientConfig *kclient.Config
}

// NewCLI returns a new OpenShift CLI client for testing
func NewCLI(project string) *CLI {
	client := CLI{}
	return client.SetupProject(project)
}

// SetupProject creates a random project (and namespace) used for executing tests
func (c *CLI) SetupProject(name string) *CLI {
	var (
		err error
	)

	if c.adminClient, err = testutil.GetClusterAdminClient(adminConfigPath); err != nil {
		FatalErr(err)
	}

	if c.adminKubeClient, err = testutil.GetClusterAdminKubeClient(adminConfigPath); err != nil {
		FatalErr(err)
	}

	if c.adminClientConfig, err = testutil.GetClusterAdminClientConfig(adminConfigPath); err != nil {
		FatalErr(err)
	}

	username := namer.GetName("test-user", randSeq(5), util.DNS1123SubdomainMaxLength)
	project := namer.GetName(name, "test-"+randSeq(5), util.DNS1123SubdomainMaxLength)
	if _, err := testutil.CreateNewProject(c.adminClient, *c.adminClientConfig, project, username); err != nil {
		FatalErr(err)
	}
	c.namespace = project

	return c
}

// Verbose turns on verbose messages
func (c *CLI) Verbose() *CLI {
	c.verbose = true
	return c
}

// AdminRESTClient return the current project namespace REST client
func (c *CLI) AdminRESTClient() *client.Client {
	return c.adminClient
}

// AdminKubeRESTClient returns the current project namespace Kubernetes client
func (c *CLI) AdminKubeRESTClient() *kclient.Client {
	return c.adminKubeClient
}

// Namespace returns the current project namespace
func (c *CLI) Namespace() string {
	return c.namespace
}

// SetOutput sets the default output for the command
func (c *CLI) SetOutput(out io.Writer) *CLI {
	c.stdout = out
	for _, subCmd := range c.cmd.Commands() {
		subCmd.SetOutput(c.stdout)
	}
	c.cmd.SetOutput(c.stdout)
	return c
}

// Run executes given OpenShift CLI command verb (iow. "oc <verb>").
// This function also override the default 'stdout' to redirect all output
// to a buffer and prepare the global flags such as namespace and config path.
func (c *CLI) Run(verb string) *CLI {
	out := new(bytes.Buffer)
	if len(c.namespace) == 0 {
		FatalErr(fmt.Errorf("You must setup project first before running a command."))
	}
	nc := &CLI{
		namespace:         c.namespace,
		adminClient:       c.adminClient,
		adminClientConfig: c.adminClientConfig,
		adminKubeClient:   c.adminKubeClient,
		verb:              verb,
		cmd:               cli.NewCommandCLI("oc", "openshift", out),
		globalArgs: []string{
			verb,
			fmt.Sprintf("--namespace=%s", c.namespace),
			fmt.Sprintf("--config=%s", adminConfigPath),
		},
	}
	return nc.SetOutput(out)
}

// Template sets a Go template for the OpenShift CLI command.
// This is equivalent of running "oc get foo -o template -t '{{ .spec }}'"
func (c *CLI) Template(t string) *CLI {
	if c.verb != "get" {
		FatalErr("Cannot use Template() for non-get verbs.")
		return c
	}
	templateArgs := []string{"--output=template", fmt.Sprintf("--template=%s", t)}
	commandArgs := append(c.commandArgs, templateArgs...)
	c.finalArgs = append(c.globalArgs, commandArgs...)
	c.cmd.SetArgs(c.finalArgs)
	return c
}

// Args sets the additional arguments for the OpenShift CLI command
func (c *CLI) Args(args ...string) *CLI {
	c.commandArgs = args
	c.finalArgs = append(c.globalArgs, c.commandArgs...)
	c.cmd.SetArgs(c.finalArgs)
	return c
}

func (c *CLI) printCmd() string {
	return strings.Join(c.finalArgs, " ")
}

// Output executes the command and return the output as string
func (c *CLI) Output() (string, error) {
	if c.verbose {
		fmt.Printf("DEBUG: oc %s\n", c.printCmd())
	}
	err := c.cmd.Execute()
	out := c.stdout.(*bytes.Buffer)
	return strings.TrimSpace(out.String()), err
}

// Execute executes the current command and return error if the execution failed
// This function will set the default output to stdout.
func (c *CLI) Execute() error {
	out, err := c.Output()
	if _, err := io.Copy(os.Stdout, strings.NewReader(out+"\n")); err != nil {
		fmt.Printf("ERROR: Unable to copy the output to stdout")
	}
	os.Stdout.Sync()
	return err
}

// WaitForResource will wait until the resource will be available.
// GO template returns '<no value>' if the demanded resource is not available.
// Returns resource as a string
func (c *CLI) WaitForResource() (string, error) {
	timeout := time.After(120 * time.Second)
	retry := time.Tick(500 * time.Millisecond)
	for {
		select {
		case <- timeout:
			return "", fmt.Errorf("ERROR: Waiting for %s %s has timeouted", c.verb, c.globalArgs)
		case <- retry:
			result, err := c.Output();
			if  err != nil {
				return "", err
			} else if result != "<no value>" {
				return result, nil
			}
		}
	}
}

// PingEndpoint will check if the socket is open
func PingEndpoint(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		FatalErr(fmt.Errorf("Error while reaching %s endpoint: %v\n", address, err))
	}
	defer conn.Close()
	fmt.Printf("Endpoint %s is reachable\n", address)
}

// FatalErr exits the test in case a fatal error occurred.
func FatalErr(msg interface{}) {
	fmt.Printf("ERROR: %v\n", msg)
	os.Exit(1)
}

// RequireServerVars verifies that all environment variables required to access
// the OpenShift server are set.
func RequireServerVars() {
	if len(GetMasterAddr()) == 0 {
		FatalErr("The 'MASTER_ADDR' environment variable must be set.")
	}
	if len(GetServerConfigDir()) == 0 {
		FatalErr("The 'SERVER_CONFIG_DIR' environment variable must be set.")
	}
}

// GetServerConfigDir returns the path to OpenShift server config directory
func GetServerConfigDir() string {
	return os.Getenv("SERVER_CONFIG_DIR")
}

// GetMasterAddr returns the address of OpenShift API server.
func GetMasterAddr() string {
	return os.Getenv("MASTER_ADDR")
}

// From github.com/GoogleCloudPlatform/kubernetes/pkg/api/generator.go
var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789-")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func waitForPodRunning(podName string, w watch.Interface) error {
	fmt.Printf("Waiting for pod %q ...\n", podName)
	for event := range w.ResultChan() {
		eventPod, ok := event.Object.(*kapi.Pod)
		if !ok {
			return fmt.Errorf("cannot convert input to pod object")
		}
		if podName != eventPod.Name {
			continue
		}
		switch eventPod.Status.Phase {
		case kapi.PodFailed:
			return fmt.Errorf("the pod %q failed: %+v", podName, eventPod.Status)
		case kapi.PodRunning:
			fmt.Printf("Pod %q status is now %q\n", podName, eventPod.Status.Phase)
			return nil
		default:
			fmt.Printf("Pod %q status is now %q\n", podName, eventPod.Status.Phase)
		}
	}
	return fmt.Errorf("unexpected closure of result channel for watcher")
}

// waitForComplete waits for the Build to finish
func waitForBuildComplete(buildName string, w watch.Interface) error {
	fmt.Printf("Waiting for build %q ...\n", buildName)
	for event := range w.ResultChan() {
		eventBuild, ok := event.Object.(*buildapi.Build)
		if !ok {
			return fmt.Errorf("cannot convert input to build object")
		}
		if buildName != eventBuild.Name {
			continue
		}
		switch eventBuild.Status.Phase {
		case buildapi.BuildPhaseFailed, buildapi.BuildPhaseError:
			return fmt.Errorf("the build %q failed: %+v", buildName, eventBuild.Status)
		case buildapi.BuildPhaseComplete:
			fmt.Printf("Build %q status is now %q\n", buildName, eventBuild.Status.Phase)
			return nil
		default:
			fmt.Printf("Build %q status is now %q\n", buildName, eventBuild.Status.Phase)
		}
	}
	return fmt.Errorf("unexpected closure of result channel for watcher")
}

// createPodForImageStream creates sample pod for given imageStream
// It resolves the dockerImageReference from the image stream.
func createPodForImageStream(oc *CLI, imageStream string) (string, error) {
	imageName, err := oc.Run("get").Args("is", imageStream).
		Template("{{ with index .status.tags 0 }}{{ with index .items 0}}{{ .dockerImageReference }}{{ end }}{{ end }}").
		Output()
	if err != nil {
		return "", err
	}
	podName := namer.GetName("test-pod", randSeq(5), util.DNS1123SubdomainMaxLength)
	fmt.Printf("Creating pod %q using %q image ...\n", podName, imageName)
	pod := &kapi.Pod{
		ObjectMeta: kapi.ObjectMeta{
			Name:   podName,
			Labels: map[string]string{"name": podName},
		},
		Spec: kapi.PodSpec{
			ServiceAccountName: "builder",
			Containers: []kapi.Container{
				{
					Name:  "test",
					Image: imageName,
				},
			},
			RestartPolicy: kapi.RestartPolicyNever,
		},
	}
	newPod, err := oc.AdminKubeRESTClient().Pods(oc.Namespace()).Create(pod)
	if err != nil {
		return "", err
	}
	return newPod.Name, nil
}
