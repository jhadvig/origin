package kubectl

import (
	"fmt"
	"text/tabwriter"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl"
	"github.com/openshift/origin/pkg/client"
)

func DescriberFor(kind string, c *client.Client) (kubectl.Describer, bool) {
	switch kind {
	case "Build":
		return &BuildDescriber{
			BuildClient: func(namespace string) (client.BuildInterface, error) {
				return c.Builds(namespace), nil
			},
		}, true
	}
	return nil, false
}

// BuildDescriber generates information about a build
type BuildDescriber struct {
	BuildClient func(namespace string) (client.BuildInterface, error)
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
		fmt.Fprintf(out, "Name:\t%s\n", build.Name)
		fmt.Fprintf(out, "Labels:\t%s\n", formatLabels(build.Labels))
		fmt.Fprintf(out, "Status:\t%s\n", string(build.Status))
		return nil
	})
}
