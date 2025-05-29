package sf

import (
	"os"
	"path/filepath"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/layout"
)

type PublishOptions struct {
	optionArchTypes    *config.ListOption
	optionFqdn         *config.StringOption
	optionFunctionName *config.StringOption
	optionFunctionType *config.StringOption
	optionOem          *config.StringOption
	optionVersion      *config.StringOption
}

func NewPublishOptions() *PublishOptions {
	return newPublishOptions("Publish")
}

func newOptionArchTypes(cmd string) *config.ListOption {
	return &config.ListOption{
		Option: config.Option{
			Key:       "SF." + cmd + ".Arches",
			Param:     "arch",
			Usage:     "a list of architecture types assigned to the function scope of execution. Supported: x86, x86_64, arm and arm64",
			Validator: config.AllFromEnumerated(layout.ArchTypes),
		},
	}
}

func newOptionFqdn(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".FQDN",
			Param: "fqdn",
			Usage: "fully qualified domain name to which the function belongs",
		},
	}
}

func newOptionFunctionName(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".FunctionName",
			Param: "functionName",
			Short: "n",
			Usage: "name of the function to create, one character or more. Only alphanumeric, dots, dashes and underscores are allowed",
			DefaultGetter: func(repo *config.Repo) any {
				var wd, _ = os.Getwd()

				return filepath.Base(wd)
			},
			Validator: config.Validation.FolderName,
		},
	}
}

func newOptionFunctionType(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF." + cmd + ".FunctionType",
			Param:        "functionType",
			Short:        "t",
			DefaultValue: "function",
			Usage:        "type of the function to create, only \"connector\", \"function\" or \"trigger\" are supported",
			Validator:    config.AnyOfEnumerated(layout.FunctionTypes),
		},
	}
}

func newOptionOem(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".OEM",
			Param: "oem",
			Usage: "unique id belonging to the publisher of the function",
		},
	}
}

func newOptionVersion(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Version",
			Param: "version",
			Short: "v",
			Usage: "initial version to use for the smart function",
		},
	}
}

func newPublishOptions(cmd string) *PublishOptions {
	return &PublishOptions{
		optionArchTypes:    newOptionArchTypes(cmd),
		optionFqdn:         newOptionFqdn(cmd),
		optionFunctionName: newOptionFunctionName(cmd),
		optionFunctionType: newOptionFunctionType(cmd),
		optionOem:          newOptionOem(cmd),
		optionVersion:      newOptionVersion(cmd),
	}
}
