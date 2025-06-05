package sf

import (
	"os"
	"path/filepath"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/layout"
)

type PublishOptions struct {
	optionArches  *config.ListOption
	optionFqdn    *config.StringOption
	optionHandle  *config.StringOption
	optionName    *config.StringOption
	optionType    *config.StringOption
	optionOem     *config.StringOption
	optionVersion *config.StringOption
}

func NewPublishOptions() *PublishOptions {
	return newPublishOptions("Publish")
}

func newOptionArches(cmd string) *config.ListOption {
	return &config.ListOption{
		Option: config.Option{
			Key:       "SF." + cmd + ".Arches",
			Param:     "arch",
			Usage:     "a list of architectures assigned to the function scope of execution. Supported: x86, x86_64, arm and arm64",
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

func newOptionHandle(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Handle",
			Param: "handle",
			Usage: "a unique string identifying the function across its oem, valid characters can only be alphanumerics, the dot, dash and underscore",
			DefaultSetter: func(repo *config.Repo) any {
				var wd, _ = os.Getwd()

				return filepath.Base(wd)
			},
			Validator: config.Validation.FolderName,
		},
	}
}

func newOptionName(defaultOption *config.StringOption, cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Name",
			Param: "name",
			Short: "n",
			Usage: "display name of the function to create, defaults to the handle value if not provided",
			DefaultGetter: func(repo *config.Repo) any {
				return repo.GetString(defaultOption)
			},
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

func newOptionType(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF." + cmd + ".Type",
			Param:        "type",
			Short:        "t",
			DefaultValue: "function",
			Usage:        "type of the function to create, only \"connector\", \"function\" or \"trigger\" are supported",
			Validator:    config.AnyOfEnumerated(layout.FunctionTypes),
		},
	}
}

func newPublishOptions(cmd string) *PublishOptions {
	var optionHandle = newOptionHandle(cmd)

	return &PublishOptions{
		optionArches:  newOptionArches(cmd),
		optionFqdn:    newOptionFqdn(cmd),
		optionHandle:  optionHandle,
		optionName:    newOptionName(optionHandle, cmd),
		optionType:    newOptionType(cmd),
		optionOem:     newOptionOem(cmd),
		optionVersion: newOptionVersion(cmd),
	}
}
