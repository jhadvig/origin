package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
	kapi "k8s.io/kubernetes/pkg/api"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	kutil "k8s.io/kubernetes/pkg/util"

	"github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	configcmd "github.com/openshift/origin/pkg/config/cmd"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	"github.com/openshift/origin/pkg/generate/app"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
	"github.com/openshift/origin/pkg/generate/dockerfile"
	"github.com/openshift/origin/pkg/generate/git"
	"github.com/openshift/origin/pkg/generate/source"
	imageapi "github.com/openshift/origin/pkg/image/api"
	"k8s.io/kubernetes/pkg/util/sets"
)

const (
	constructAppLong = `
Interactively construct a new application.
`

	constructAppMsg = `We're going to create an OpenShift application for you!
(1) Create an application based on a git repository
(2) Deploy an application based on a pre-existing imagestream
`
	setAdditionMetaMsg   = `Want to set additional %[1]s to your project ?`
	setEnvForResourceMsg = `Select the resource to which you would like to add the specified environment variables:
(1) BuildConfig
(2) DeploymentConfig
(3) BuildConfig and DeploymentConfig
(4) Done adding env variables
`
)

// NewCmdConstructApplication implements the OpenShift cli construct-app command
func NewCmdConstructApplication(fullName string, f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	config := newcmd.NewAppConfig()
	config.Deploy = true

	cmd := &cobra.Command{
		Use:        "construct-app",
		Short:      "Construct a new application interactively",
		Long:       fmt.Sprintf(constructAppLong, fullName),
		SuggestFor: []string{"app", "application"},
		Run: func(c *cobra.Command, args []string) {
			mapper, typer := f.Object()
			config.SetMapper(mapper)
			config.SetTyper(typer)
			config.SetClientMapper(f.ClientMapperForCommand())

			err := RunConstructApplication(fullName, f, reader, out, c, args, config)
			if err == errExit {
				os.Exit(1)
			}
			cmdutil.CheckErr(err)
		},
	}

	cmd.Flags().StringP("output", "o", "json", "Output format. One of: json|yaml|template|templatefile.")
	cmd.Flags().Bool("no-headers", false, "When using the default output, don't print headers.")
	//cmd.Flags().Bool("show-all", true, "")
	cmd.Flags().StringP("output-version", "", "v1", "")

	return cmd
}

// RunConstructApplication contains all the necessary functionality for the OpenShift cli new-app command
func RunConstructApplication(fullName string, f *clientcmd.Factory, reader io.Reader, out io.Writer, c *cobra.Command, args []string, config *newcmd.AppConfig) error {

	items := &kapi.List{}

	appName := util.PromptForString(reader, out, "What would you like to name the thing you are creating: ")
	is := &imageapi.ImageStream{
		ObjectMeta: kapi.ObjectMeta{
			Name: appName,
		},
	}
	items.Items = append(items.Items, is)

	fmt.Fprintf(out, constructAppMsg)
	validChoices := sets.NewString("1", "2")
	userChoice := util.PromptForString(reader, out, "What type of application would you like to create: ")
	// Keep prompting until they give us a valid choice:
	for !validChoices.Has(userChoice) {
		fmt.Fprintf(out, "Invalid input: %s\n", userChoice)
		userChoice = util.PromptForString(reader, out, "What type of application would you like to create: ")
	}

	if userChoice == "1" {
		userEnv, _ := addEnvVars(reader, out)

		// TODO: clone git repository and try to determine type of application here
		fmt.Fprintf(out, "You chose to create an application from an existing git repository, where is it?\n")
		bc, ist := constructFromGitRepo(appName, reader, out, userEnv.BuildConfigEnvs)
		ports := collectPorts(f, reader, out, ist.namespace, ist.name, ist.tag)
		containerPorts := createContainerPorts(ports)
		triggerIST := &imageStreamTag{
			name: appName,
			tag:  "latest",
		}
		dc := defineDeploymentConfig(appName, triggerIST, userEnv.DeploymentConfigEnvs, containerPorts)
		items.Items = append(items.Items, dc)
		items.Items = append(items.Items, bc)
		for i, port := range ports {
			if port.asService {
				s := defineService(fmt.Sprintf("%s-service-%d", appName, i), port, appName)
				items.Items = append(items.Items, s)
			}
		}

	} else if userChoice == "2" {
		userEnv, _ := addEnvVars(reader, out)
		// TODO: clone git repository and try to determine type of application here
		fmt.Fprintf(out, "You chose to create an application from an existing image stream\n")
		ist := getImageStream(reader, out)
		ports := collectPorts(f, reader, out, ist.namespace, ist.name, ist.tag)
		containerPorts := createContainerPorts(ports)
		dc := defineDeploymentConfig(appName, ist, userEnv.DeploymentConfigEnvs, containerPorts)
		items.Items = append(items.Items, dc)
		for i, port := range ports {
			if port.asService {
				s := defineService(fmt.Sprintf("%s-service-%d", appName, i), port, appName)
				items.Items = append(items.Items, s)
			}
		}
	} else if userChoice == "3" {
		fmt.Fprintf(out, "You chose to instantiate a template. We don't support that yet.\n")
	}

	createObjects2(f, items)

	/*
		items := &kapi.List{}
		ports := collectPorts(f, reader, out, "openshift", "jenkins", "latest")
		containerPorts := createContainerPorts(ports)
		ist := imageStreamTag{
			namespace: "openshift",
			name:      "jenkins",
			tag:       "latest",
		}
		envs := []kapi.EnvVar{}
		name := "myapp"
		dc := defineDeploymentConfig(name, ist, envs, containerPorts)
		for i, port := range ports {
			if port.asService {
				s := defineService(fmt.Sprintf("%s-service-%d", name, i), port, name)
				items.Items = append(items.Items, s)
			}
		}
		items.Items = append(items.Items, dc)

		createObjects2(f, items)
	*/

	return nil
}

func getImageStream(reader io.Reader, out io.Writer) *imageStreamTag {
	fmt.Fprintf(out, "You now must specify an image stream for the resulting application.\n\n")
	namespace := util.PromptForString(reader, out, "Namespace: ")
	imageStream := util.PromptForString(reader, out, "Image Stream: ")
	tag := util.PromptForString(reader, out, "Tag: ")

	return &imageStreamTag{name: imageStream, namespace: namespace, tag: tag}
}

func constructFromGitRepo(appName string, reader io.Reader, out io.Writer, envs []kapi.EnvVar) (*api.BuildConfig, *imageStreamTag) {
	fmt.Fprintf(out, "Please specify your git repository URL.")
	fmt.Fprintf(out, "\nex. https://github.com/openshift/ruby-hello-world.git\n\n")
	gitRepoLoc := util.PromptForString(reader, out, "Git repository: ")

	// Try to determine what type of repository this might be:
	// Clone git repository into a local directory:
	// TODO: Support references to commit/branch equivalent to new-app
	var err error
	srcRef := &app.SourceRef{}
	if srcRef.Dir, err = ioutil.TempDir("", "gen"); err != nil {
		fmt.Printf("Error %v", err)
		return nil, nil
	}
	fmt.Fprintf(out, "Created temp dir for git clone: %s\n", srcRef.Dir)
	gitRepoUrl, err := url.Parse(gitRepoLoc)
	if err != nil {
		fmt.Printf("Invalid repository URL: %s", gitRepoLoc)
		return nil, nil
	}
	srcRef.URL = gitRepoUrl
	gitRepo := git.NewRepository()
	fmt.Fprintf(out, "Cloning %s to %s for source detection...", srcRef.URL.String(), srcRef.Dir)
	if err = gitRepo.Clone(srcRef.Dir, srcRef.URL.String()); err != nil {
		fmt.Fprintf(out, "Something went pretty wrong: %s\n", err)
		fmt.Printf("Error cloning git repository at: %s", srcRef.URL.String())
		return nil, nil
	}

	// Now we try to detect what kind of repo this might be:
	srcRepoEnumerator := app.SourceRepositoryEnumerator{
		Detectors: source.DefaultDetectors,
		Tester:    dockerfile.NewTester(),
	}
	info, err := srcRepoEnumerator.Detect(srcRef.Dir)
	if err != nil {
		fmt.Printf("Error %v", err)
		return nil, nil
	}
	for i := range info.Types {
		t := info.Types[i]
		fmt.Fprintf(out, "This appears to be a %s project.\n\n", t.Platform)
	}

	// Prompt the user for a namespace, imagestream, and tag
	// TODO: in the future support showing the user lists to choose from.
	ist := getImageStream(reader, out)

	// Create a BuildConfig:
	objRef := kapi.ObjectReference{
		Kind:      "ImageStreamTag",
		Namespace: ist.namespace,
		Name:      fmt.Sprintf("%s:%s", ist.name, ist.tag),
	}

	bc := &api.BuildConfig{
		ObjectMeta: kapi.ObjectMeta{
			Name: appName,
		},
		Spec: api.BuildConfigSpec{
			Triggers: []api.BuildTriggerPolicy{
				{
					Type: api.ConfigChangeBuildTriggerType,
				},
			},
			BuildSpec: api.BuildSpec{
				Source: api.BuildSource{
					Type: api.BuildSourceGit,
					Git: &api.GitBuildSource{
						URI: gitRepoUrl.String(),
					},
				},
				Strategy: api.BuildStrategy{
					Type: api.SourceBuildStrategyType,
					SourceStrategy: &api.SourceBuildStrategy{
						From: objRef,
						Env:  envs,
					},
				},
				Output: api.BuildOutput{
					To: &kapi.ObjectReference{
						Kind: "ImageStreamTag",

						Name: appName + ":latest",
					},
				},
			},
		},
	}
	return bc, ist

	//	data, err := latest.Codec.Encode(bc)
	//	fmt.Fprint(out, string(data))

	//return nil
}

type ConstructedResourceEnvs struct {
	BuildConfigEnvs      []kapi.EnvVar
	DeploymentConfigEnvs []kapi.EnvVar
	SetOnAll             bool
}

// addEnvVars will prompt user for adding environment variables to desired resource, either bc, dc,
// both or on each bc and dc that the app construction creates.
func addEnvVars(reader io.Reader, out io.Writer) (ConstructedResourceEnvs, error) {
	addEnv := true
	resourceEnvVars := ConstructedResourceEnvs{}

	for addEnv {
		userChoice := util.PromptForString(reader, out, setEnvForResourceMsg)
		if userChoice == "4" {
			break
		}
		envString := util.PromptForString(reader, out, "Write down environment variables you would like to add in following form: 'KEY_1=VAL_1 ... KEY_N=VAL_N'")
		envVars, _, err := ParseEnv(strings.Split(envString, " "), reader)
		if err != nil {
			fmt.Printf("Error: %v", err)
		}

		switch userChoice {
		case "3":
			resourceEnvVars.BuildConfigEnvs = append(resourceEnvVars.BuildConfigEnvs, envVars...)
			resourceEnvVars.DeploymentConfigEnvs = append(resourceEnvVars.DeploymentConfigEnvs, envVars...)
		case "2":
			resourceEnvVars.DeploymentConfigEnvs = append(resourceEnvVars.DeploymentConfigEnvs, envVars...)
		case "1":
			resourceEnvVars.BuildConfigEnvs = append(resourceEnvVars.BuildConfigEnvs, envVars...)
		}

	}
	fmt.Printf("Adding env vars %v", resourceEnvVars)
	return resourceEnvVars, nil
}

type imageStreamTag struct {
	namespace string
	name      string
	tag       string
}

func defineDeploymentConfig(name string, ist *imageStreamTag, envs []kapi.EnvVar, ports []kapi.ContainerPort) *deployapi.DeploymentConfig {
	dc := &deployapi.DeploymentConfig{
		ObjectMeta: kapi.ObjectMeta{
			Name: name,
		},
		Triggers: []deployapi.DeploymentTriggerPolicy{
			{
				Type: deployapi.DeploymentTriggerOnImageChange,
				ImageChangeParams: &deployapi.DeploymentTriggerImageChangeParams{
					Automatic: true,
					ContainerNames: []string{
						ist.name,
					},
					From: kapi.ObjectReference{
						Namespace: ist.namespace,
						Name:      ist.name + ":latest",
						Kind:      "ImageStreamTag",
					},
				},
			},
		},

		Template: deployapi.DeploymentTemplate{
			Strategy: deployapi.DeploymentStrategy{
				Type: deployapi.DeploymentStrategyTypeRecreate,
			},
			ControllerTemplate: kapi.ReplicationControllerSpec{
				Replicas: 1,
				Selector: map[string]string{
					"selector": name,
				},
				Template: &kapi.PodTemplateSpec{
					ObjectMeta: kapi.ObjectMeta{
						Labels: map[string]string{
							"selector": name,
						},
					},
					Spec: kapi.PodSpec{
						Containers: []kapi.Container{
							{
								Name:  ist.name,
								Image: ist.namespace + "/" + ist.name + ":" + ist.tag,
								Ports: ports,
								Env:   envs,
							},
						},
					},
				},
			},
		},
	}
	return dc
}

func defineService(name string, port *Port, selectorName string) *kapi.Service {
	s := &kapi.Service{
		ObjectMeta: kapi.ObjectMeta{
			Name: name,
		},
		Spec: kapi.ServiceSpec{
			Ports: []kapi.ServicePort{
				{
					Protocol:   port.protocol,
					Port:       port.port,
					TargetPort: kutil.NewIntOrStringFromInt(port.port),
				},
			},
			Selector: map[string]string{
				"selector": selectorName,
			},
		},
	}
	return s
}

type Port struct {
	port      int
	protocol  kapi.Protocol
	asService bool
	asRoute   bool
	routeHost string
}

func createContainerPorts(ports []*Port) []kapi.ContainerPort {
	cPorts := []kapi.ContainerPort{}
	for _, port := range ports {
		cPorts = append(cPorts, kapi.ContainerPort{
			ContainerPort: port.port,
			Protocol:      port.protocol,
		})
	}
	return cPorts
}

func collectPorts(f *clientcmd.Factory, reader io.Reader, out io.Writer, namespace, imagestream, tag string) []*Port {
	osclient, _, _ := f.Clients()
	istClient := osclient.ImageStreamTags(namespace)
	ist, _ := istClient.Get(imagestream, tag)

	ports := []string{}
	for exposed := range ist.Image.DockerImageMetadata.Config.ExposedPorts {
		ports = append(ports, strings.Split(exposed, " ")...)
	}

	appPorts := []*Port{}

	for _, sp := range ports {
		p := docker.Port(sp)
		port, err := strconv.Atoi(p.Port())
		if err != nil {
			fmt.Fprintf(out, "failed to parse port %q: %v", p.Port(), err)
			continue
		}

		appPorts = append(appPorts, &Port{
			port:     port,
			protocol: kapi.ProtocolTCP,
		})
	}

	for true {
		fmt.Fprintf(out, "The %s:%s imagestreamtag exposes the following ports:\n", imagestream, tag)
		for i, port := range appPorts {
			fmt.Fprintf(out, "%d) %d (service=%v,route=%v)\n", i, port.port, port.asService, port.asRoute)
		}
		fmt.Fprintf(out, "%v) Add a port\n", len(appPorts))
		fmt.Fprintf(out, "%v) Done\n", len(appPorts)+1)
		choice, _ := strconv.Atoi(util.PromptForString(reader, out, "Select an option: "))

		var updatedPort *Port
		switch choice {
		case len(appPorts):
			newPort, _ := strconv.Atoi(util.PromptForString(reader, out, "Port number: "))
			updatedPort = &Port{port: newPort, protocol: kapi.ProtocolTCP}
			appPorts = append(appPorts, updatedPort)
		case len(appPorts) + 1:
			return appPorts
		default:
			updatedPort = appPorts[choice]
		}

		updatedPort.asService = util.PromptForBool(reader, out, "Expose as a service(y/n): ")

		updatedPort.asRoute = util.PromptForBool(reader, out, "Expose a route(y/n):")
		if updatedPort.asRoute {
			updatedPort.routeHost = util.PromptForString(reader, out, "Route hostname: ")
		}
	}
	return appPorts
}

func createObjects2(f *clientcmd.Factory, items *kapi.List) {
	mapper, typer := f.Factory.Object()
	bulk := configcmd.Bulk{
		Mapper:            mapper,
		Typer:             typer,
		RESTClientFactory: f.Factory.RESTClient,
	}
	namespace, _, _ := f.DefaultNamespace()
	if errs := bulk.Create(items, namespace); len(errs) != 0 {
		fmt.Printf("Error creating objects: %v", errs)
	}
}
