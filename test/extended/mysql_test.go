package extended

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
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

	oc.Run("get").Args("service", "mysql").Template("{{ .spec.ClusterIP }}").Execute()
}
