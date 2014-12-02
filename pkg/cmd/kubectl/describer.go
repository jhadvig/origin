package kubectl

import (
	"fmt"
	"text/tabwriter"
	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/meta"
	kctl "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl"
	kclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	kubecmd "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"
	buildapi "github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/api/latest"
	"github.com/openshift/origin/pkg/client"
)

func DescriberFor(kind string, c *client.Client, cmd *cobra.Command) (kctl.Describer, bool) {
	switch kind {
	case "Build":
		return &BuildDescriber{
			BuildClient: func(namespace string) (client.BuildInterface, error) {
				return c.Builds(namespace), nil
			},
		}, true
	case "BuildConfig":
		return &BuildConfigDescriber{
			BuildConfigClient: func(namespace string) (client.BuildConfigInterface, error) {
				return c.BuildConfigs(namespace), nil
			},
			Command: cmd,
		}, true
	case "Deployment":
		return &DeploymentDescriber{
			DeploymentClient: func(namespace string) (client.DeploymentInterface, error) {
				return c.Deployments(namespace), nil
			},
		}, true
	case "DeploymentConfig":
		return &DeploymentConfigDescriber{
			DeploymentConfigClient: func(namespace string) (client.DeploymentConfigInterface, error) {
				return c.DeploymentConfigs(namespace), nil
			},
		}, true
	}
	return nil, false
}

// BuildDescriber generates information about a build
type BuildDescriber struct {
	BuildClient func(namespace string) (client.BuildInterface, error)
}

func (d *BuildDescriber) DescribeParameters(p buildapi.BuildParameters, out *tabwriter.Writer) {
	fmt.Fprintf(out, "Strategy:\t%s\n", string(p.Strategy.Type))
	fmt.Fprintf(out, "Source Type:\t%s\n", string(p.Source.Type))
	if p.Source.Git != nil {
		fmt.Fprintf(out, "URL:\t%s\n", string(p.Source.Git.URI))
		if len(p.Source.Git.Ref) > 0 {
			fmt.Fprintf(out, "Ref:\t%s\n", string(p.Source.Git.Ref))
		}
	}
	fmt.Fprintf(out, "Output Image:\t%s\n", string(p.Output.ImageTag))
	fmt.Fprintf(out, "Output Registry:\t%s\n", string(p.Output.Registry))
}

func (d *BuildDescriber) Describe(namespace, name string) (string, error) {
	bc, err := d.BuildClient(namespace)
	if err != nil {
		return "", err
	}
	build, err := bc.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, build.ObjectMeta)
		fmt.Fprintf(out, "Status:\t%s\n", string(build.Status))
		fmt.Fprintf(out, "Build Pod:\t%s\n", string(build.PodName))
		d.DescribeParameters(build.Parameters, out)
		return nil
	})
}

// DeploymentDescriber generates information about a deployment
type DeploymentDescriber struct {
	DeploymentClient func(namespace string) (client.DeploymentInterface, error)
}

func (d *DeploymentDescriber) Describe(namespace, name string) (string, error) {
	bc, err := d.DeploymentClient(namespace)
	if err != nil {
		return "", err
	}
	deployment, err := bc.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, deployment.ObjectMeta)
		fmt.Fprintf(out, "Status:\t%s\n", string(deployment.Status))
		fmt.Fprintf(out, "Strategy:\t%s\n", string(deployment.Strategy.Type))
		fmt.Fprintf(out, "Causes:\n")
		for _, c := range deployment.Details.Causes {
			fmt.Fprintf(out, "\t\t%s\n", string(c.Type))
		}
		// TODO: Add description for controllerTemplate
		return nil
	})
}

// DeploymentConfigDescriber generates information about a DeploymentConfig
type DeploymentConfigDescriber struct {
	DeploymentConfigClient func(namespace string) (client.DeploymentConfigInterface, error)
}

func (d *DeploymentConfigDescriber) Describe(namespace, name string) (string, error) {
	bc, err := d.DeploymentConfigClient(namespace)
	if err != nil {
		return "", err
	}
	deploymentConfig, err := bc.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, deploymentConfig.ObjectMeta)
		fmt.Fprintf(out, "Latest Version:\t%s\n", string(deploymentConfig.LatestVersion))
		fmt.Fprintf(out, "Triggers:\t\n")
		for _, t := range deploymentConfig.Triggers {
			fmt.Fprintf(out, "Type:\t%s\n", t.Type)
		}
		return nil
	})
}

// BuildConfigDescriber generates information about a buildConfig
type BuildConfigDescriber struct {
	BuildConfigClient func(namespace string) (client.BuildConfigInterface, error)
	Command *cobra.Command
}

func (d *BuildConfigDescriber) Describe(namespace, name string) (string, error) {
	bc, err := d.BuildConfigClient(namespace)
	if err != nil {
		return "", err
	}
	buildConfig, err := bc.Get(name)
	if err != nil {
		return "", err
	}

	kubeConfig := kubecmd.GetKubeConfig(d.Command)
	webhooks := GetWebhookUrl(buildConfig, kubeConfig)

	// fmt.Printf("\n\n&v\n\n", webhooks)


	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, buildConfig.ObjectMeta)
		fmt.Fprintf(out, "Source:\t%s\n", string(buildConfig.Parameters.Source.Git.URI))
		fmt.Fprintf(out, "Strategy:\t%s\n", string(buildConfig.Parameters.Strategy.Type))
		fmt.Fprintf(out, "Image:\t%s\n", string(buildConfig.Parameters.Output.ImageTag))
		for whType, whURL := range webhooks {
			fmt.Fprintf(out, "Webhook-%s:\t%s\n", string(whType), string(whURL))
		}
		return nil
	})
}

// GetWebhookUrl assembles array of webhook urls which can trigger given buildConfig
func GetWebhookUrl(bc *buildapi.BuildConfig, config *kclient.Config) map[string]string {
	// triggers := make([]string, len(bc.Triggers))
	triggers := make(map[string]string)
	for i, trigger := range bc.Triggers {
		var whTrigger string
		switch trigger.Type {
		case "github":
			whTrigger = trigger.GithubWebHook.Secret
		case "generic":
			whTrigger = trigger.GenericWebHook.Secret
		}
		apiVersion := latest.Version
		if accessor, err := meta.Accessor(bc); err == nil && len(accessor.APIVersion()) > 0 {
			apiVersion = accessor.APIVersion()
		}
		url := fmt.Sprintf("%s/osapi/%s/buildConfigHooks/%s/%s/%s", config.Host, apiVersion, bc.Name, whTrigger, bc.Triggers[i].Type)
		triggers[string(trigger.Type)] = url
	}
	return triggers
}