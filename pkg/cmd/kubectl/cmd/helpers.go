package cmd

import (
	"os"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl"
	kubecmd "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// TODO: Remove this when this func will be public in upstream
func usageError(cmd *cobra.Command, format string, args ...interface{}) {
	glog.Errorf(format, args...)
	glog.Errorf("See '%s -h' for help.", cmd.CommandPath())
	os.Exit(1)
}

// TODO: Remove this when this func will be public in upstream
func checkErr(err error) {
	if err != nil {
		glog.Fatalf("%v", err)
	}
}

// TODO: Make this public in upstream
func getOriginNamespace(cmd *cobra.Command) string {
	result := kapi.NamespaceDefault
	if ns := kubecmd.GetFlagString(cmd, "namespace"); len(ns) > 0 {
		result = ns
		glog.V(2).Infof("Using namespace from -ns flag")
	} else {
		nsPath := kubecmd.GetFlagString(cmd, "ns-path")
		nsInfo, err := kubectl.LoadNamespaceInfo(nsPath)
		if err != nil {
			glog.Fatalf("Error loading current namespace: %v", err)
		}
		result = nsInfo.Namespace
	}
	glog.V(2).Infof("Using namespace %s", result)
	return result

}
