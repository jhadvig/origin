package kubectl

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
)

// TODO: Make this function exported in upstream
func tabbedString(f func(*tabwriter.Writer) error) (string, error) {
	out := new(tabwriter.Writer)
	b := make([]byte, 1024)
	buf := bytes.NewBuffer(b)
	out.Init(buf, 0, 8, 1, '\t', 0)

	err := f(out)
	if err != nil {
		return "", err
	}

	out.Flush()
	str := string(buf.String())
	return str, nil
}

// TODO Make this function exported in upstream
// TODO Move to labels package.
func formatLabels(labelMap map[string]string) string {
	l := labels.Set(labelMap).String()
	if l == "" {
		l = "<none>"
	}
	return l
}

func formatMeta(out *tabwriter.Writer, m api.ObjectMeta) {
	fmt.Fprintf(out, "Name:\t%s\n", m.Name)
	fmt.Fprintf(out, "Annotations:\t%s\n", formatLabels(m.Annotations))
	fmt.Fprintf(out, "Labels:\t%s\n", formatLabels(m.Labels))
	fmt.Fprintf(out, "Created:\t%s\n", m.CreationTimestamp)
}
