package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
)

const (
	constructAppLong = `
Interactively construct a new application.
`
)

// NewCmdConstructApplication implements the OpenShift cli construct-app command
func NewCmdConstructApplication(fullName string, f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	config := newcmd.NewAppConfig()
	config.Deploy = true

	cmd := &cobra.Command{
		Use:        "construct-app",
		Short:      "Construct a new application interactively",
		Long:       fmt.Sprintf(constructAppLong, fullName),
		SuggestFor: []string{"app", "application"},
		Run: func(c *cobra.Command, args []string) {
			mapper, typer := f.Object()
			config.SetMapper(mapper)
			config.SetTyper(typer)
			config.SetClientMapper(f.ClientMapperForCommand())

			err := RunConstructApplication(fullName, f, reader, out, c, args, config)
			if err == errExit {
				os.Exit(1)
			}
			cmdutil.CheckErr(err)
		},
	}

	return cmd
}

// RunConstructApplication contains all the necessary functionality for the OpenShift cli new-app command
func RunConstructApplication(fullName string, f *clientcmd.Factory, reader io.Reader, out io.Writer, c *cobra.Command, args []string, config *newcmd.AppConfig) error {

	fmt.Fprintf(out, "Hey look we're running construct-app!!!\n")

	fmt.Fprint(out, "We're going to create an OpenShift application for you!\n")
	fmt.Fprintf(out, "What would you like to do?\n")
	fmt.Fprintf(out, "(1) Create an application based on a git repository\n")
	fmt.Fprintf(out, "(2) Deploy an application based on a pre-existing imagestream\n")
	fmt.Fprintf(out, "(3) Instantiate a template\n")
	// TODO: recieve input 1-3 and validate it's correct
	userChoice := util.PromptForString(reader, out, "What type of application would you like to create: ")

	if userChoice == "1" {
		// TODO: clone git repository and try to determine type of application here
		fmt.Fprintf(out, "You chose to create an application from an existing git repository, where is it?\n")
	} else if userChoice == "2" {
		fmt.Fprintf(out, "You chose to create an application from an existing imagesteam?\n")
	} else if userChoice == "3" {
		fmt.Fprintf(out, "You chose to instantiate a template. We don't support that yet.\n")
	}

	return nil
}
