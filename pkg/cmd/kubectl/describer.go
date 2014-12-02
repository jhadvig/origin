package kubectl

import (
	"fmt"
	"text/tabwriter"

	kctl "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl"
	buildapi "github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/client"
)

func DescriberFor(kind string, c *client.Client) (kctl.Describer, bool) {
	switch kind {
	case "Build":
		return &BuildDescriber{
			BuildClient: func(namespace string) (client.BuildInterface, error) {
				return c.Builds(namespace), nil
			},
		}, true
	case "Deployment":
		return &DeploymentDescriber{
			DeploymentClient: func(namespace string) (client.DeploymentInterface, error) {
				return c.Deployments(namespace), nil
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
