package layout

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz/lang/enumz"
	"genaiz.com/genaiz/lang/panicz"
)

type ArchType = string
type ConfigType = string
type FunctionType = string

const (
	archTypeX86    ArchType = "x86"
	archTypeX86_64 ArchType = "x86_64"
	archTypeArm    ArchType = "arm"
	archTypeArm64  ArchType = "arm64"

	configTypeJson ConfigType = "json"
	configTypeNone ConfigType = ""
	configTypeToml ConfigType = "toml"
	configTypeYaml ConfigType = "yaml"

	functionTypeConnector FunctionType = "connector"
	functionTypeFunction  FunctionType = "function"
	functionTypeTrigger   FunctionType = "trigger"
)

var (
	ArchTypes     = enumz.NewEnumType(archTypeX86, archTypeX86_64, archTypeArm, archTypeArm64)
	ConfigTypes   = enumz.NewEnumType(configTypeJson, configTypeNone, configTypeToml, configTypeYaml)
	FunctionTypes = enumz.NewEnumType(functionTypeConnector, functionTypeFunction, functionTypeTrigger)
)

type ConfigPretender interface {
	PretendSlice(string, []string)

	PretendValue(string, string)
}

type keyValueFilePretender struct {
	filename   string
	pretendFmt string
}

func (kp keyValueFilePretender) PretendSlice(key string, slice []string) {
	var value = "[\"" + strings.Join(slice, "\",\"") + "\"]"

	kp.PretendValue(key, value)
}

func (kp keyValueFilePretender) PretendValue(key string, value string) {
	fmt.Printf(kp.pretendFmt+"\n", key, value, kp.filename)
}

func newConfigPretender(path string) ConfigPretender {
	var ct, err = ConfigTypes.FromString(filepath.Ext(path)[1:])

	panicz.PanicIfError(err)
	switch *ct {
	case configTypeJson:
		break
	case configTypeToml:
		break
	case configTypeYaml:
		return &keyValueFilePretender{
			filename:   path,
			pretendFmt: "yq -i e '.%s=%s' %s",
		}
	}

	panic("config type is not supported by pretenders")
}

func pretendMap(pretender ConfigPretender, provider func() map[string]string) {
	for k, v := range provider() {
		if v != "" {
			pretender.PretendValue(k, v)
		}
	}
}

func pretendSlice(pretender ConfigPretender, provider func() (string, []string)) {
	var key, slice = provider()

	if slice != nil {
		pretender.PretendSlice(key, cast.ToStringSlice(slice))
	}
}

func pretendValue(pretender ConfigPretender, provider func() (string, string)) {
	var key, value = provider()

	if value != "" {
		pretender.PretendValue(key, value)
	}
}
