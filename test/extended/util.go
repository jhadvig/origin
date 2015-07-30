package extended

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/kubernetes/test/e2e"
)

var testContext = e2e.TestContextType{}

func kubeConfigPath() string {
	return os.Getenv("KUBECONFIG")
}

func writeTempJSON(path, content string) error {
	return ioutil.WriteFile(path, []byte(content), 0644)
}

func getTempFilePath(name string) string {
	return filepath.Join(testContext.OutputDir, name+".json")
}
