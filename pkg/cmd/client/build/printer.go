package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"encoding/json"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubecfg"
	"github.com/openshift/origin/pkg/build/api"
)

var buildColumns = []string{"ID", "Status", "Pod ID"}
var buildConfigColumns = []string{"ID", "Type", "SourceURI"}
var buildLogColumns = []string{"Timestamp", "Log"}

// RegisterPrintHandlers registers HumanReadablePrinter handlers
// for build and buildConfig resources.
func RegisterPrintHandlers(printer *kubecfg.HumanReadablePrinter) {
	printer.Handler(buildColumns, printBuild)
	printer.Handler(buildColumns, printBuildList)
	printer.Handler(buildConfigColumns, printBuildConfig)
	printer.Handler(buildConfigColumns, printBuildConfigList)
	printer.Handler(buildLogColumns, printBuildLog)
}

func printBuild(build *api.Build, w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s\t%s\t%s\n", build.ID, build.Status, build.PodID)
	return err
}
func printBuildList(buildList *api.BuildList, w io.Writer) error {
	for _, build := range buildList.Items {
		if err := printBuild(&build, w); err != nil {
			return err
		}
	}
	return nil
}

func printBuildConfig(bc *api.BuildConfig, w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s\t%s\t%s\n", bc.ID, bc.DesiredInput.Type, bc.DesiredInput.SourceURI)
	return err
}
func printBuildConfigList(buildList *api.BuildConfigList, w io.Writer) error {
	for _, buildConfig := range buildList.Items {
		if err := printBuildConfig(&buildConfig, w); err != nil {
			return err
		}
	}
	return nil
}

func printBuildLog(bl *api.BuildLog, w io.Writer) error {

	test, _ := json.Marshal(bl)
	err := ioutil.WriteFile("/tmp/printer", test, 0644)
	if err != nil {
		return nil
	}

	for _, log := range bl.LogItems {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", log.Timestamp, log.Log); err != nil {
			return err
		}
	}
	return nil
}