package builder

import (
	buildapi "github.com/openshift/origin/pkg/build/api"
	stiapi "github.com/openshift/source-to-image/pkg/api"
)

// getBuildEnvVars returns a map with the environment variables that should be added
// to the built image
func getBuildEnvVars(build *buildapi.Build) map[string]string {
	envVars := map[string]string{
		"OPENSHIFT_BUILD_NAME":      build.Name,
		"OPENSHIFT_BUILD_NAMESPACE": build.Namespace,
		"OPENSHIFT_BUILD_SOURCE":    build.Spec.Source.Git.URI,
	}
	if build.Spec.Source.Git.Ref != "" {
		envVars["OPENSHIFT_BUILD_REFERENCE"] = build.Spec.Source.Git.Ref
	}
	if build.Spec.Revision != nil &&
		build.Spec.Revision.Git != nil &&
		build.Spec.Revision.Git.Commit != "" {
		envVars["OPENSHIFT_BUILD_COMMIT"] = build.Spec.Revision.Git.Commit
	}
	if build.Spec.Strategy.Type == buildapi.SourceBuildStrategyType {
		userEnv := build.Spec.Strategy.SourceStrategy.Env
		for _, v := range userEnv {
			envVars[v.Name] = v.Value
		}
	}
	return envVars
}

// setBuildLabels ...
func setBuildLabels(info *stiapi.SourceInfo) map[string]string {

	labels := make(map[string]string)

	if info.CommitID != "" {
		labels["io.openshift.build.commit.id"] = info.CommitID
	}
	if info.Author != "" {
		labels["io.openshift.build.commit.author"] = info.Author
	}
	if info.Date != "" {
		labels["io.openshift.build.commit.date"] = info.Date
	}
	if info.Message != "" {
		labels["io.openshift.build.commit.message"] = info.Message
	}
	if info.Ref != "" {
		labels["io.openshift.build.commit.ref"] = info.Ref
	}
	if info.Location != "" {
		labels["io.openshift.build.commit.location"] = info.Location
	}
	if info.ContextDir != "" {
		labels["io.openshift.build.commit.source-context-dir"] = info.ContextDir
	}
	return labels
}
	