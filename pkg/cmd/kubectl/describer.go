package kubectl

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"github.com/spf13/cobra"

	kctl "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl"
	kubecmd "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"
	buildapi "github.com/openshift/origin/pkg/build/api"
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
	case "Image":
		return &ImageDescriber{
			ImageClient: func(namespace string) (client.ImageInterface, error) {
				return c.Images(namespace), nil
			},
		}, true
	case "ImageRepository":
		return &ImageRepositoryDescriber{
			ImageRepositoryClient: func(namespace string) (client.ImageRepositoryInterface, error) {
				return c.ImageRepositories(namespace), nil
			},
		}, true
	case "Route":
		return &RouteDescriber{
			RouteClient: func(namespace string) (client.RouteInterface, error) {
				return c.Routes(namespace), nil
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
	webhooks := WebhookUrl(buildConfig, kubecmd.GetKubeConfig(d.Command))
	buildDescriber := &BuildDescriber{}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, buildConfig.ObjectMeta)
		buildDescriber.DescribeParameters(buildConfig.Parameters, out)
		for whType, whURL := range webhooks {
			fmt.Fprintf(out, "Webhook %s:\t%s\n", strings.Title(string(whType)), string(whURL))
		}
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

// ImageDescriber generates information about a Image
type ImageDescriber struct {
	ImageClient func(namespace string) (client.ImageInterface, error)
}

func (d *ImageDescriber) Describe(namespace, name string) (string, error) {
	bc, err := d.ImageClient(namespace)
	if err != nil {
		return "", err
	}
	image, err := bc.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, image.ObjectMeta)
		fmt.Fprintf(out, "Docker Image:\t%s\n", string(image.DockerImageReference))
		return nil
	})
}

// ImageRepositoryDescriber generates information about a ImageRepository
type ImageRepositoryDescriber struct {
	ImageRepositoryClient func(namespace string) (client.ImageRepositoryInterface, error)
}

func (d *ImageRepositoryDescriber) Describe(namespace, name string) (string, error) {
	bc, err := d.ImageRepositoryClient(namespace)
	if err != nil {
		return "", err
	}
	imageRepository, err := bc.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, imageRepository.ObjectMeta)
		fmt.Fprintf(out, "Tags:\t%s\n", formatLabels(imageRepository.Tags))
		fmt.Fprintf(out, "Docker Repository:\t%s\n", string(imageRepository.DockerImageRepository))
		return nil
	})
}

// RouteDescriber generates information about a Route
type RouteDescriber struct {
	RouteClient func(namespace string) (client.RouteInterface, error)
}

func (d *RouteDescriber) Describe(namespace, name string) (string, error) {
	bc, err := d.RouteClient(namespace)
	if err != nil {
		return "", err
	}
	route, err := bc.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, route.ObjectMeta)
		fmt.Fprintf(out, "Host:\t%s\n", string(route.Host))
		fmt.Fprintf(out, "Path:\t%s\n", string(route.Path))
		fmt.Fprintf(out, "Service Name:\t%s\n", string(route.ServiceName))
		return nil
	})
}
