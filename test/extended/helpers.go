package extended

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/util/namer"
	testutil "github.com/openshift/origin/test/util"
)

var adminConfigPath = filepath.Join(GetServerConfigDir(), "master", "admin.kubeconfig")

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

// LoginAndCreateProject creates project with given name that will belong to the
// provided user. It returns the path to admin kubeconfig and the project name
func LoginAndCreateProject(user, project string) string {
	client, err := testutil.GetClusterAdminClient(adminConfigPath)
	if err != nil {
		FatalErr(err)
	}

	adminConfig, err := testutil.GetClusterAdminClientConfig(adminConfigPath)
	if err != nil {
		FatalErr(err)
	}

	projectName := namer.GetName(project, "test-"+randSeq(5), util.DNS1123SubdomainMaxLength)
	if _, err := testutil.CreateNewProject(
		client,
		*adminConfig,
		projectName,
		namer.GetName(user, "test-"+randSeq(5), util.DNS1123SubdomainMaxLength),
	); err != nil {
		FatalErr(err)
	}

	return projectName
}

func RunCLI(commandName, ns string, args []string, in io.Reader, out io.Writer) error {
	// TODO: Handle stdin
	cmd := cli.NewCommandCLI("oc", "openshift", in, out)
	for _, c := range cmd.Commands() {
		c.SetOutput(out)
	}
	cmd.SetOutput(out)
	cmdArgs := []string{commandName}
	cmdArgs = append(cmdArgs, args...)
	authArgs := []string{
		"-n", ns,
		"--config=" + adminConfigPath,
	}
	cmdArgs = append(cmdArgs, authArgs...)
	cmd.SetArgs(cmdArgs)
	fmt.Printf("command=%+v\n", cmdArgs)
	return cmd.Execute()
}

// FatalErr exits the test in case a fatal error occurred.
func FatalErr(msg interface{}) {
	fmt.Printf("ERROR: %v\n", msg)
	os.Exit(1)
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
