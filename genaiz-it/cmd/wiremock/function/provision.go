package function

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/wiremock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/panicz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type ProvisionExecutor struct {
	Context context.Context
	Cli     *BaseCli
	Options *ProvisionOptions

	mappingNames []string
}

type ProvisionOptions struct {
	optionDockerImage *config.StringOption
}

func (pe *ProvisionExecutor) getAuthFile() string {
	var home, err = os.UserHomeDir()

	panicz.PanicIfError(err)
	return filepath.Join(home, ".docker", "config.json")
}

func (pe *ProvisionExecutor) getPassword() []byte {
	var password = os.Getenv("GENAIZ_PASS")

	if password == "" {
		password = "genaiz_pass"
	}

	return []byte(password)
}

func (pe *ProvisionExecutor) getUsername() string {
	var user = os.Getenv("GENAIZ_USER")

	if user == "" {
		return "genaiz_user"
	}

	return user
}

func (pe *ProvisionExecutor) getRepositoryAndVersion() (string, string) {
	var image = pe.Cli.ledger.GetString(pe.Options.optionDockerImage)

	if image != "" {
		var parts = strings.Split(image, ":")

		if len(parts) == 1 {
			return parts[0], "latest"
		}

		return parts[0], parts[1]
	}

	return "", ""
}

func (pe *ProvisionExecutor) digest() (string, error) {
	var repository, version = pe.getRepositoryAndVersion()

	if repository != "" {
		var inspectTask = docker.NewInspectTask()
		var buildParams = &docker.BuildParams{
			Env: task.Env{
				Context: pe.Context,
			},
			DockerTag:     repository,
			DockerVersion: version,
		}
		var state *task.State

		if state = inspectTask.Execute(buildParams, &logrus.Logger{}); state.Error == nil {
			return state.Output, nil
		}

		return state.Output, state.Error
	}

	return "", nil
}

func (pe *ProvisionExecutor) login() (string, error) {
	var loginTask = docker.NewLoginTask()
	var loginPassword = pe.getPassword()
	var loginParams = &docker.LoginParams{
		Env: task.Env{
			Context: pe.Context,
		},
		AuthFile: pe.getAuthFile(),
		Host:     pe.Cli.GetRegistryUrl(),
		Username: pe.getUsername(),
		Password: &loginPassword,
	}
	var state *task.State

	if state = loginTask.Execute(loginParams, &logrus.Logger{}); state.Error == nil {
		return state.Output, nil
	}

	return state.Output, state.Error
}

func (pe *ProvisionExecutor) update(digest string, token string) error {
	var wiremockClient = wiremock.NewWiremockClient(pe.Cli.GetWiremockUrl())
	var stubs []*wiremock.AdminMapping
	var err error

	if stubs, err = wiremockClient.GetStubsByPath("/v1/sf"); err == nil {
		for _, stub := range stubs {
			if slices.Contains(pe.mappingNames, stub.Name) {
				var body = *stub.Response.JsonBody
				var provision = body["data"].(map[string]interface{})
				var sf = provision["sf"].(map[string]interface{})

				sf["digest"] = digest
				provision["Auth"] = token

				if err = wiremockClient.UpdateStub(stub.Id.String(), stub); err != nil {
					break
				}
			}
		}
	}

	return err
}

func NewProvision(cli *BaseCli) *cobra.Command {
	var options = NewProvisionOptions()
	var provision = &cobra.Command{
		Use:     "provision MAPPINGS...",
		Short:   "Invokes wiremock function provision setup",
		Long:    "Invokes wiremock function provision setup on mock-broker/v1/sf/provision",
		Example: "genaiz-it wiremock function provision mock_publish_simple_function_provision_ok",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = NewProvisionExecutor(cmd.Context(), cli, options, args...)
			var digest, token string
			var err error

			if token, err = exec.login(); err == nil {
				if digest, err = exec.digest(); err == nil {
					err = exec.update(digest, token)
				}
			}

			cobra.CheckErr(err)
		},
	}

	cli.ledger.Register(provision, options.optionDockerImage)
	return provision
}

func NewProvisionExecutor(ctx context.Context, cli *BaseCli, options *ProvisionOptions, names ...string) *ProvisionExecutor {
	return &ProvisionExecutor{
		Context: ctx,
		Cli:     cli,
		Options: options,

		mappingNames: names,
	}
}

func NewProvisionOptions() *ProvisionOptions {
	return &ProvisionOptions{
		optionDockerImage: newOptionDockerImage(),
	}
}

func newOptionDockerImage() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Param: "image",
			Short: "i",
			Usage: "Docker image to provision",
		},
	}
}
