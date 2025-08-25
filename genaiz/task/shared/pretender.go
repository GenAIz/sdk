package shared

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
)

type ConfigPretender interface {
	PretendDeleteByField(string, string, string)

	PretendDelete(string)

	PretendSlice(string, []string)

	PretendValue(string, string)
}

type keyValueFilePretender struct {
	filename string

	deleteByFieldFmt string
	deleteFmt        string
	pretendFmt       string
}

func (kp keyValueFilePretender) PretendDeleteByField(key, field, value string) {
	fmt.Printf(kp.deleteByFieldFmt+"\n", key, field, value, kp.filename)
}

func (kp keyValueFilePretender) PretendDelete(key string) {
	fmt.Printf(kp.deleteFmt+"\n", key, kp.filename)
}

func (kp keyValueFilePretender) PretendSlice(key string, slice []string) {
	var value = "[\"" + strings.Join(slice, "\",\"") + "\"]"

	kp.PretendValue(key, value)
}

func (kp keyValueFilePretender) PretendValue(key string, value string) {
	fmt.Printf(kp.pretendFmt+"\n", key, value, kp.filename)
}

func NewConfigPretender(path string) ConfigPretender {
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
			filename: path,

			deleteByFieldFmt: "yq -i 'del(.%s | select(.%s == \"%s\"))' %s",
			deleteFmt:        "yq -i 'del(.%s)' %s",
			pretendFmt:       "yq -i e '.%s=\"%s\"' %s",
		}
	}

	panic("config type is not supported by pretenders")
}

func PretendDeleteByField(pretender ConfigPretender, provider func() (string, string, string)) {
	var key, field, value = provider()

	if key != "" && field != "" && value != "" {
		pretender.PretendDeleteByField(key, field, value)
	}
}

func PretendDelete(pretender ConfigPretender, provider func() string) {
	var key = provider()

	if key != "" {
		pretender.PretendDelete(key)
	}
}

func PretendMap(pretender ConfigPretender, provider func() map[string]string) {
	for k, v := range provider() {
		if v != "" {
			pretender.PretendValue(k, v)
		}
	}
}

func PretendSlice(pretender ConfigPretender, provider func() (string, []string)) {
	var key, slice = provider()

	if slice != nil {
		pretender.PretendSlice(key, cast.ToStringSlice(slice))
	}
}

func PretendValue(pretender ConfigPretender, provider func() (string, string)) {
	var key, value = provider()

	if value != "" {
		pretender.PretendValue(key, value)
	}
}
