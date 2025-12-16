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

	PretendDeleteBySelect(string, []ConfigSelect)

	PretendDelete(string)

	PretendSlice(string, []string)

	PretendValue(string, string)
}

type ConfigSelect struct {
	Field string
	Value string
}

type keyValueFilePretender struct {
	filename   string
	deleteFmt  string
	pretendFmt string
	selectFmt  string
}

func (kp keyValueFilePretender) PretendDeleteByField(key, field, value string) {
	var selection = fmt.Sprintf(" | "+kp.selectFmt, field, value)

	fmt.Printf(kp.deleteFmt+"\n", key, selection, kp.filename)
}

func (kp keyValueFilePretender) PretendDeleteBySelect(key string, selects []ConfigSelect) {
	var selection []string

	for _, s := range selects {
		if s.Field != "" && s.Value != "" {
			selection = append(selection, fmt.Sprintf(kp.selectFmt, s.Field, s.Value))
		}
	}

	if len(selection) > 0 {
		fmt.Printf(kp.deleteFmt+"\n", key, " | "+strings.Join(selection, " | "))
	}
}

func (kp keyValueFilePretender) PretendDelete(key string) {
	fmt.Printf(kp.deleteFmt+"\n", key, "", kp.filename)
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
			filename:   path,
			deleteFmt:  "yq -i 'del(.%s%s)' %s",
			pretendFmt: "yq -i e '.%s=\"%s\"' %s",
			selectFmt:  "select(.%s == \"%s\")",
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

func PretendDeleteBySelect(pretender ConfigPretender, provider func() (string, []ConfigSelect)) {
	var key, selects = provider()

	if key != "" && len(selects) > 0 {
		pretender.PretendDeleteBySelect(key, selects)
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
