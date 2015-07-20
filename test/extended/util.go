package extended

import (
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/kubernetes/test/e2e"
)

var testContext = e2e.TestContextType{}

func adminKubeConfigPath() string {
	if kubeConfigPath := os.Getenv("SERVER_KUBECONFIG_PATH"); len(kubeConfigPath) != 0 {
		return filepath.Join(os.Getenv("SERVER_CONFIG_DIR"), kubeConfigPath)
	}
	return filepath.Join(os.Getenv("SERVER_CONFIG_DIR"), "master", "admin.kubeconfig")
}

func kubeCertPath() string {
	if p := os.Getenv("SERVER_CERT_DIR"); len(p) > 0 {
		return p
	}
	return filepath.Join(os.Getenv("SERVER_CONFIG_DIR"), "master")
}

func tempJSON(name string) string {
	return filepath.Join(testContext.OutputDir, name+".json")
}
