package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	kclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	. "github.com/GoogleCloudPlatform/kubernetes/test/e2e"
	. "github.com/onsi/ginkgo"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/admin/policy"
	"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/cmd/server/bootstrappolicy"
	testutil "github.com/openshift/origin/test/util"
	"github.com/spf13/cobra"
)

type CLI struct {
	namespace       string
	verb            string
	adminConfigPath string
	globalArgs      []string
	commandArgs     []string
	finalArgs       []string
	stdout          io.Writer
	verbose         bool
	cmd             *cobra.Command
	Framework       *Framework
}

// NewCLI initialize the Kubernetes E2E framework, sets namespaces and provide
// project role bindings.
// If runInGinko argument is set to true, this function will install BeforeEach
// hook to Gingo test case that grant project role bindings to a namespace created
// by the E2E test. This option must be set inside Describe() or Context()
func NewCLI(project, configPath string, runInGinko bool) *CLI {
	client := &CLI{}
	client.Framework = NewFramework(project)
	client.SetAdminKubeConfigPath(configPath)
	if runInGinko {
		func() { BeforeEach(client.beforeEach) }()
	}
	return client
}

func (c *CLI) beforeEach() {
	if ns := c.Framework.Namespace; ns != nil {
		Logf("Adding project role bindings to %q namespace", ns.Name)
		c.SetupProject(ns.Name).SetNamespaceFromFramework()
	} else {
		Failf("Framework does not have the namespace set")
	}
}

// SetAdminKubeConfigPath allows to override the path to admin kubeconfig file.
func (c *CLI) SetAdminKubeConfigPath(path string) {
	if len(path) == 0 {
		FatalErr(fmt.Errorf("The admin.kubeconfig path can not be empty"))
	}
	c.adminConfigPath = path
}

// SetupProject setups a project role binding for the Kubernetes namespace
func (c *CLI) SetupProject(ns string) *CLI {
	for _, binding := range bootstrappolicy.GetBootstrapServiceAccountProjectRoleBindings(ns) {
		addRole := &policy.RoleModificationOptions{
			RoleName:            binding.RoleRef.Name,
			RoleNamespace:       binding.RoleRef.Namespace,
			RoleBindingAccessor: policy.NewLocalRoleBindingAccessor(ns, c.AdminRESTClient()),
			Users:               binding.Users.List(),
			Groups:              binding.Groups.List(),
		}
		if err := addRole.AddRole(); err != nil {
			Failf("Unable to add role binding %+v to namespace %q: %v", addRole, ns, err)
		}
	}
	return c
}

// SetNamespaceFromFramework sets the current namespace to Kubernetes E2E Framework
// namespace.
func (c *CLI) SetNamespaceFromFramework() {
	c.namespace = c.Framework.Namespace.Name
	return
}

// Verbose turns on verbose messages
func (c *CLI) Verbose() *CLI {
	c.verbose = true
	return c
}

// AdminRESTClient return the current project namespace REST client
func (c *CLI) AdminRESTClient() *client.Client {
	if client, err := testutil.GetClusterAdminClient(c.adminConfigPath); err != nil {
		FatalErr(err)
		return nil
	} else {
		return client
	}
}

// AdminKubeRESTClient returns the current project namespace Kubernetes client
func (c *CLI) AdminKubeRESTClient() *kclient.Client {
	if c.Framework.Client != nil {
		return c.Framework.Client
	}
	if client, err := testutil.GetClusterAdminKubeClient(c.adminConfigPath); err != nil {
		FatalErr(err)
		return nil
	} else {
		return client
	}
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
		namespace: c.namespace,
		verb:      verb,
		cmd:       cli.NewCommandCLI("oc", "openshift", out),
		globalArgs: []string{
			verb,
			fmt.Sprintf("--namespace=%s", c.namespace),
			fmt.Sprintf("--config=%s", c.adminConfigPath),
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

// FatalErr exits the test in case a fatal error has occured.
func FatalErr(msg interface{}) {
	fmt.Printf("ERROR: %v\n", msg)
	os.Exit(1)
}
