package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/spf13/cobra"
)

const (
	CreateSourceSecretRecommendedCommandName = "new-gitconfig"

	createSourceSecretLong = `
Create a new source secret

Source secrets are used to authenticate against Git server.


`
)

type CreateSourceSecretOptions struct {
	SecretNamespace string
	SecretName      string
	Username        string
	Password        string
	Token           string

	GitConfigPath   string
	CertificatePath string

	SecretsInterface client.SecretsInterface

	Out io.Writer
}

func NewCmdCreateSourceSecret(name, fullName string, f *cmdutil.Factory, out io.Writer, newSecretFullName, ocEditFullName string) *cobra.Command {
	o := &CreateSourceSecretOptions{Out: out}

	cmd := &cobra.Command{
		Use: fmt.Sprintf("%s SECRET_NAME --username=USERNAME --password=PASSWORD --token=TOKEN --cert-path=CERTIFICATE", name),
		Short: "Create a new source secret",
		Long: fmt.Sprintf(createSourceSecretLong, fullName, newSecretFullName, ocEditFullName),
		Run: func(c *cobra.Command, args []string) {
			if err := o.Complete(f, args); err != nil {
				cmdutil.CheckErr(cmdutil.UsageError(c, err.Error()))
			}

			if err := o.Validate(); err != nil {
				cmdutil.CheckErr(cmdutil.UsageError(c, err.Error()))
			}

			if err := o.CreateSourceSecret(); err != nil {
				cmdutil.CheckErr(err)
			}
		},
	}

	cmd.Flags().StringVar(&o.Username, "username", "", "username for Git authentication")
	cmd.Flags().StringVar(&o.Password, "password", "", "password for Git authentication")
	cmd.Flags().StringVar(&o.Token, "token", "", "token for Git authentication")
	cmd.Flags().StringVar(&o.CertificatePath, "cert-path", "", "path to the certificate file")
	// cmd.Flags().StringVar(&o.GitConfigPath, "gitconfig-path", "", "path to the gitconfig file")
	cmdutil.AddPrinterFlags(cmd)

	return cmd
}

func (o *CreateSourceSecretOptions) CreateSourceSecret() error {
	secret, err := o.MakeSourceSecret()
	if err != nil {
		return err
	}

	if _, err := o.SecretsInterface.Create(secret); err != nil {
		return err
	}

	// fmt.Fprintf(o.GetOut(), "secret/%s\n", secret.Name)

	return nil
}

type SourceConfigEntry struct {
	// Username    string
	Password    string
	Token       string
	Certificate string
}

const (
	GitConfigKey = ".gitconfig"
)

func (o *CreateSourceSecretOptions) MakeSourceSecret() (*api.Secret, error) {

	ca_content, err := ioutil.ReadFile(o.CertificatePath)
	if err != nil {
		return nil, err
	}

	gitAuth := SourceConfigEntry{
		// Username:    o.Username,
		Password:    o.Password,
		Token:       o.Token,
		Certificate: string(ca_content),
	}

	gitConfig := map[string]SourceConfigEntry{o.Username: gitAuth}

	gitConfigContent, err := json.Marshal(gitConfig)
	if err != nil {
		return nil, err
	}

	secret := &api.Secret{}
	secret.Namespace = o.SecretNamespace
	secret.Name = o.SecretName
	secret.Type = api.SecretTypeOpaque
	secret.Data = map[string][]byte{}
	secret.Data[GitConfigKey] = gitConfigContent

	return secret, nil
}



func (o *CreateSourceSecretOptions) Complete(f *cmdutil.Factory, args []string) error {
	for i, _ := range args {
		fmt.Printf(args[i])
	}
	
	if len(args) != 1 {
		return errors.New("must have exactly one argument: secret name")
	}
	o.SecretName = args[0]

	client, err := f.Client()
	if err != nil {
		return err
	}
	o.SecretNamespace, _, err = f.DefaultNamespace()
	if err != nil {
		return err
	}

	o.SecretsInterface = client.Secrets(o.SecretNamespace)

	return nil
}

func (o CreateSourceSecretOptions) Validate() error {
	if len(o.GitConfigPath) != 0 {
		err := o.CheckGitConfig()
		if err != nil {
			return err
		}
	}

	switch {
	case len(o.SecretNamespace) == 0 :
		return errors.New("source secret namespace must be present")
	case len(o.SecretName) == 0 :
		return errors.New("source secret name must be present")
	case len(o.Username) == 0 :
		return errors.New("source secret username must be present")
	case len(o.Password) == 0 && len(o.Token) == 0 :
		return errors.New("source secret password or token must be present")
	}

	if strings.Contains(o.Username, ":") {
		return fmt.Errorf("username '%v' is illegal because it contains a ':'", o.Username)
	}

	return nil
}

func (o CreateSourceSecretOptions) CheckGitConfig() error {
	params := map[string]string{
		"user": "",
		"password": "",
		"token": "",
	}

	git, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("exec unavailable - unable to locate git")
	}

	for i, _ := range params {
		out, err := exec.Command(git, "config", "--file", o.GitConfigPath, fmt.Sprintf("github.%s", i)).Output()
		if err != nil {
			return fmt.Errorf("unable to parse git config, %v", err)
		}
		params[i] = string(out)
	}

	if len(params["user"]) != 0 {
		o.Username = params["user"]
	}
	if len(params["password"]) != 0 {
		o.Password = params["password"]
	}
	if len(params["token"]) != 0 {
		o.Token = params["token"]
	}
	return nil
}
