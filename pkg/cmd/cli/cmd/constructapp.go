package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	kapi "k8s.io/kubernetes/pkg/api"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/openshift/origin/pkg/api/latest"
	"github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/generate/app"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
	"github.com/openshift/origin/pkg/generate/dockerfile"
	"github.com/openshift/origin/pkg/generate/git"
	"github.com/openshift/origin/pkg/generate/source"
	"k8s.io/kubernetes/pkg/util/sets"
)

const (
	constructAppLong = `
Interactively construct a new application.
`

	constructAppMsg = `We're going to create an OpenShift application for you!
(1) Create an application based on a git repository
(2) Deploy an application based on a pre-existing imagestream
(3) Instantiate a template
`
	setAdditionMetaMsg   = `Want to set additional %[1]s to your project ?`
	setEnvForResourceMsg = `Select the resource to which you would like to add the specified environment variables:
(1) BuildConfig
(2) DeploymentConfig
(3) BuildConfig and DeploymentConfig
(4) Globally to all resources created during application construction
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

	fmt.Fprintf(out, constructAppMsg)

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
	fmt.Fprintf(out, "\nex. https://github.com/openshift/ruby-hello-world.git\n\n")
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

	// Now we try to detect what kind of repo this might be:
	srcRepoEnumerator := app.SourceRepositoryEnumerator{
		Detectors: source.DefaultDetectors,
		Tester:    dockerfile.NewTester(),
	}
	info, err := srcRepoEnumerator.Detect(srcRef.Dir)
	if err != nil {
		return err
	}
	for i := range info.Types {
		t := info.Types[i]
		fmt.Fprintf(out, "This appears to be a %s project.\n\n", t.Platform)
	}

	// Prompt the user for a namespace, imagestream, and tag
	// TODO: in the future support showing the user lists to choose from.
	fmt.Fprintf(out, "You now must specify an image stream for the resulting application.\n\n")

	namespace := util.PromptForString(reader, out, "Namespace: ")
	imageStream := util.PromptForString(reader, out, "Image Stream: ")
	tag := util.PromptForString(reader, out, "Tag: ")

	// Create a BuildConfig:
	objRef := kapi.ObjectReference{
		Kind:      "ImageStreamTag",
		Namespace: namespace,
		Name:      fmt.Sprintf("%s:%s", imageStream, tag),
	}

	bc := &api.BuildConfig{
		Spec: api.BuildConfigSpec{
			BuildSpec: api.BuildSpec{
				Source: api.BuildSource{
					Type: api.BuildSourceGit,
					Git: &api.GitBuildSource{
						URI: gitRepoUrl.String(),
					},
				},
				Strategy: api.BuildStrategy{
					Type: api.SourceBuildStrategyType,
					SourceStrategy: &api.SourceBuildStrategy{
						From: objRef,
					},
				},
			},
		},
	}
	data, err := latest.Codec.Encode(bc)
	fmt.Fprint(out, string(data))

	return nil
}

type ConstructedResourceEnvs struct {
	BuildConfigEnvs      []kapi.EnvVar
	DeploymentConfigEnvs []kapi.EnvVar
	SetOnAll             bool
}

// addEnvVars will prompt user for adding environment variables to desired resource, either bc, dc,
// both or on each bc and dc that the app construction creates.
func addEnvVars(reader io.Reader, out io.Writer) (ConstructedResourceEnvs, error) {
	addEnv := true
	validResourceChoices := sets.NewString("1", "2", "3", "4")
	resourceEnvVars := ConstructedResourceEnvs{}

	for addEnv {
		addEnv := util.PromptForBool(reader, out, fmt.Sprintf(setAdditionMetaMsg, "environment variable"))
		if addEnv {
			userChoice := util.PromptForString(reader, out, setEnvForResourceMsg)
			for validResourceChoices.Has(userChoice) {
				userChoice = util.PromptForString(reader, out, setEnvForResourceMsg)
			}

			envString := util.PromptForString(reader, out, "Write down environment variables you would like to add in following form: 'KEY_1=VAL_1 ... KEY_N=VAL_N'")
			envVars, _, err := ParseEnv(strings.Split(envString, " "), reader)
			if err != nil {
				return ConstructedResourceEnvs{}, err
			}

			switch userChoice {
			case "4":
				resourceEnvVars.SetOnAll = true
				fallthrough
			case "3":
				fallthrough
			case "2":
				resourceEnvVars.DeploymentConfigEnvs = append(resourceEnvVars.DeploymentConfigEnvs, envVars...)
				if userChoice != "3" {
					continue
				}
				fallthrough
			case "1":
				resourceEnvVars.BuildConfigEnvs = append(resourceEnvVars.BuildConfigEnvs, envVars...)
			}
		}

	}
	return resourceEnvVars, nil
}
