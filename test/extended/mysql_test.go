package extended

import (
	"bytes"
	"fmt"
	"testing"
)

func init() {
	fmt.Printf("Checking server ...\n")
	RequireServerVars()
}

func TestMysqlCreateFromTemplate(t *testing.T) {
	project := LoginAndCreateProject("mysql", "mysql-create")

	out := new(bytes.Buffer)

	RunCLI("process", project, []string{
		"-f", "../../examples/db-templates/mysql-ephemeral-template.json",
	}, out)

	fmt.Printf("output=%v\n", out.String())

	// oc create -f ^^
	//err := client.Create(objects)
	// [[ wait for mysql ]]

	//client.Endpoints().Get(objects[0].Name)
	// test connection
	//. ..
}
