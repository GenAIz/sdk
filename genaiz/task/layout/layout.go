package layout

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/lang/enumz"
)

type ArchType = string
type ConfigType = string
type FunctionType = string

const (
	ArchTypeX86    ArchType = "x86"
	ArchTypeX86_64 ArchType = "x86_64"
	ArchTypeArm    ArchType = "arm"
	ArchTypeArm64  ArchType = "arm64"

	ConfigTypeJson ConfigType = "json"
	ConfigTypeNone ConfigType = ""
	ConfigTypeToml ConfigType = "toml"
	ConfigTypeYaml ConfigType = "yaml"

	FunctionTypeConnector FunctionType = "connector"
	FunctionTypeFunction  FunctionType = "function"
	FunctionTypeTrigger   FunctionType = "trigger"
)

var (
	ArchTypes     = enumz.NewEnumType(ArchTypeX86, ArchTypeX86_64, ArchTypeArm, ArchTypeArm64)
	ConfigTypes   = enumz.NewEnumType(ConfigTypeJson, ConfigTypeNone, ConfigTypeToml, ConfigTypeYaml)
	FunctionTypes = enumz.NewEnumType(FunctionTypeConnector, FunctionTypeFunction, FunctionTypeTrigger)
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
	var fileType = filez.GetFileType(path)
	var ct, err = ConfigTypes.FromString(fileType)

	panicz.PanicIfError(err)
	switch *ct {
	case ConfigTypeJson:
		break
	case ConfigTypeToml:
		break
	case ConfigTypeYaml:
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
