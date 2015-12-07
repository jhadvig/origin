package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
)

const (
	constructAppLong = `
Interactively construct a new application.
`
)

// NewCmdConstructApplication implements the OpenShift cli construct-app command
func NewCmdConstructApplication(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
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

			err := RunConstructApplication(fullName, f, out, c, args, config)
			if err == errExit {
				os.Exit(1)
			}
			cmdutil.CheckErr(err)
		},
	}

	cmd.Flags().StringSliceVar(&config.SourceRepositories, "code", config.SourceRepositories, "Source code to use to build this application.")
	cmd.Flags().StringVar(&config.ContextDir, "context-dir", "", "Context directory to be used for the build.")
	cmd.Flags().StringSliceVarP(&config.ImageStreams, "image", "", config.ImageStreams, "Name of an image stream to use in the app. (deprecated)")
	cmd.Flags().StringSliceVarP(&config.ImageStreams, "image-stream", "i", config.ImageStreams, "Name of an image stream to use in the app.")
	cmd.Flags().StringSliceVar(&config.DockerImages, "docker-image", config.DockerImages, "Name of a Docker image to include in the app.")
	cmd.Flags().StringSliceVar(&config.Templates, "template", config.Templates, "Name of a stored template to use in the app.")
	cmd.Flags().StringSliceVarP(&config.TemplateFiles, "file", "f", config.TemplateFiles, "Path to a template file to use for the app.")
	cmd.Flags().StringSliceVarP(&config.TemplateParameters, "param", "p", config.TemplateParameters, "Specify a list of key value pairs (e.g., -p FOO=BAR,BAR=FOO) to set/override parameter values in the template.")
	cmd.Flags().StringSliceVar(&config.Groups, "group", config.Groups, "Indicate components that should be grouped together as <comp1>+<comp2>.")
	cmd.Flags().StringSliceVarP(&config.Environment, "env", "e", config.Environment, "Specify key value pairs of environment variables to set into each container.")
	cmd.Flags().StringVar(&config.Name, "name", "", "Set name to use for generated application artifacts")
	cmd.Flags().StringVar(&config.Strategy, "strategy", "", "Specify the build strategy to use if you don't want to detect (docker|source).")
	cmd.Flags().StringP("labels", "l", "", "Label to set in all resources for this application.")
	cmd.Flags().BoolVar(&config.InsecureRegistry, "insecure-registry", false, "If true, indicates that the referenced Docker images are on insecure registries and should bypass certificate checking")
	cmd.Flags().BoolVarP(&config.AsList, "list", "L", false, "List all local templates and image streams that can be used to create.")
	cmd.Flags().BoolVarP(&config.AsSearch, "search", "S", false, "Search all templates, image streams, and Docker images that match the arguments provided.")
	cmd.Flags().BoolVar(&config.AllowMissingImages, "allow-missing-images", false, "If true, indicates that referenced Docker images that cannot be found locally or in a registry should still be used.")
	cmd.Flags().BoolVar(&config.AllowSecretUse, "grant-install-rights", false, "If true, a component that requires access to your account may use your token to install software into your project. Only grant images you trust the right to run with your token.")
	cmd.Flags().BoolVar(&config.SkipGeneration, "no-install", false, "Do not attempt to run images that describe themselves as being installable")
	cmd.Flags().BoolVar(&config.DryRun, "dry-run", false, "If true, do not actually create resources.")

	// TODO AddPrinterFlags disabled so that it doesn't conflict with our own "template" flag.
	// Need a better solution.
	// cmdutil.AddPrinterFlags(cmd)
	cmd.Flags().StringP("output", "o", "", "Output format. One of: json|yaml|template|templatefile.")
	cmd.Flags().String("output-version", "", "Output the formatted object with the given version (default api-version).")
	cmd.Flags().Bool("no-headers", false, "When using the default output, don't print headers.")
	cmd.Flags().String("output-template", "", "Template string or path to template file to use when -o=template or -o=templatefile.  The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview]")

	return cmd
}

// RunConstructApplication contains all the necessary functionality for the OpenShift cli new-app command
func RunConstructApplication(fullName string, f *clientcmd.Factory, out io.Writer, c *cobra.Command, args []string, config *newcmd.AppConfig) error {

	fmt.Fprintf(out, "Hey look we're running construct-app!!!")
	return nil
}
