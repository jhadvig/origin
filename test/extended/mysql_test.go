package extended

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
)

func init() {
	RequireServerVars()
}

var templatePath = filepath.Join("..", "..", "examples", "db-templates", "mysql-ephemeral-template.json")

func TestMysqlCreateFromTemplate(t *testing.T) {
	// FIXME: Remove the Verbose()
	oc := NewCLI("mysql-create").Verbose()

	// Process the template and store the output in temporary file
	listOutput, err := oc.Run("process").Args("-f", templatePath).Output()
	if err != nil {
		t.Fatalf("Unexpected error while processing %s: %v", templatePath, err)
	}

	listPath := filepath.Join(os.TempDir(), oc.Namespace()+".json")
	defer os.Remove(listPath)
	if err := ioutil.WriteFile(listPath, []byte(listOutput), 0644); err != nil {
		t.Fatalf("Unexpected error while writing list file: %v", err)
	}

	if err := oc.Run("create").Args("-f", listPath).Execute(); err != nil {
		t.Fatalf("Unexpected error while creating: %v", err)
	}

	endpointWatcher, err := oc.AdminKubeRESTClient().Endpoints(oc.Namespace()).
		Watch(labels.Everything(), fields.Everything(), "0")
	if err != nil {
		t.Fatalf("Unable to create watcher for endpoints: %v", err)
	}
	defer endpointWatcher.Stop()

	list, err := oc.AdminKubeRESTClient().Endpoints(oc.Namespace()).List(labels.Everything())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	for _, endpoint := range list.Items {
		if err := waitForEndpoint(endpoint.Name,endpointWatcher); err != nil {
			t.Fatalf("Endpoint error: %v\n", err)
		}
	}
}
