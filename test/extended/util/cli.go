package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	kclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
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
