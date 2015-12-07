package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"net/url"

	"github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/generate/app"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
	"github.com/openshift/origin/pkg/generate/git"
	"k8s.io/kubernetes/pkg/util/sets"
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

	fmt.Fprint(out, "We're going to create an OpenShift application for you!\n\n")
	fmt.Fprintf(out, "(1) Create an application based on a git repository\n")
	fmt.Fprintf(out, "(2) Deploy an application based on a pre-existing imagestream\n")
	fmt.Fprintf(out, "(3) Instantiate a template\n\n")

	validChoices := sets.NewString("1", "2", "3")
	userChoice := util.PromptForString(reader, out, "What type of application would you like to create: ")
	// Keep prompting until they give us a valid choice:
	for !validChoices.Has(userChoice) {
		fmt.Fprintf(out, "Invalid input: %s\n", userChoice)
		userChoice = util.PromptForString(reader, out, "What type of application would you like to create: ")
	}

	if userChoice == "1" {
		// TODO: clone git repository and try to determine type of application here
		fmt.Fprintf(out, "You chose to create an application from an existing git repository, where is it?\n")
		constructFromGitRepo(reader, out)
	} else if userChoice == "2" {
		fmt.Fprintf(out, "You chose to create an application from an existing imagesteam?\n")
	} else if userChoice == "3" {
		fmt.Fprintf(out, "You chose to instantiate a template. We don't support that yet.\n")
	}

	return nil
}

func constructFromGitRepo(reader io.Reader, out io.Writer) error {
	fmt.Fprintf(out, "Please specify your git repository URL.")
	fmt.Fprintf(out, "\nex. https://github.com/openshift/origin.git\n\n")
	gitRepoLoc := util.PromptForString(reader, out, "Git repository: ")

	// Try to determine what type of repository this might be:
	// Clone git repository into a local directory:
	// TODO: Support references to commit/branch equivalent to new-app
	var err error
	srcRef := &app.SourceRef{}
	if srcRef.Dir, err = ioutil.TempDir("", "gen"); err != nil {
		return err
	}
	fmt.Fprintf(out, "Created temp dir for git clone: %s\n", srcRef.Dir)
	gitRepoUrl, err := url.Parse(gitRepoLoc)
	if err != nil {
		return fmt.Errorf("Invalid repository URL: %s", gitRepoLoc)
	}
	srcRef.URL = gitRepoUrl
	gitRepo := git.NewRepository()
	fmt.Fprintf(out, "Cloning %s to %s for source detection...", srcRef.URL.String(), srcRef.Dir)
	if err = gitRepo.Clone(srcRef.Dir, srcRef.URL.String()); err != nil {
		fmt.Fprintf(out, "Something went pretty wrong: %s\n", err)
		return fmt.Errorf("Error cloning git repository at: %s", srcRef.URL.String())
	}
	return nil
}
