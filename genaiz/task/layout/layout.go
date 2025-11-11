package layout

import (
	"maps"
	"path/filepath"
	"regexp"
	"strings"

	"genaiz.com/genaiz/lang/enumz"
	"genaiz.com/genaiz/task"
)

type ArchType = string
type FunctionType = string

// x86_64 is not getting renamed for golang no
//
//goland:noinspection GoSnakeCaseUsage
const (
	ArchTypeX86    ArchType = "x86"
	ArchTypeX86_64 ArchType = "x86_64"
	ArchTypeArm    ArchType = "arm"
	ArchTypeArm64  ArchType = "arm64"

	FunctionTypeConnector FunctionType = "connector"
	FunctionTypeFunction  FunctionType = "function"
	FunctionTypeTrigger   FunctionType = "trigger"
)

var (
	ArchTypes     = enumz.NewEnumType(ArchTypeX86, ArchTypeX86_64, ArchTypeArm, ArchTypeArm64)
	FunctionTypes = enumz.NewEnumType(FunctionTypeConnector, FunctionTypeFunction, FunctionTypeTrigger)

	outputRegex    = regexp.MustCompile(`out(?:put)?/?$`)
	timestampRegex = regexp.MustCompile(`\{timestamp(?::.+?)?}/?$`)
)

type initTracking struct {
	configFile string
	params     map[string]string
}

type InitLayout struct {
	DirInput  string
	DirLog    string
	DirOutput string
	DirVar    string
}

func NewInitLayout(in, out string) *InitLayout {
	var outTrimmed = strings.TrimSpace(out)
	var outBytes = []byte(outTrimmed)
	var result = &InitLayout{
		DirInput: strings.TrimSpace(in),
	}

	if timestampRegex.Match(outBytes) {
		result.DirLog = filepath.Join(outTrimmed, "log")
		result.DirOutput = filepath.Join(outTrimmed, "out")
		result.DirVar = filepath.Join(outTrimmed, "var")
	} else if outputRegex.Match(outBytes) {
		var dir = filepath.Dir(outTrimmed)

		result.DirLog = filepath.Join(dir, "log")
		result.DirOutput = outTrimmed
		result.DirVar = filepath.Join(dir, "var")
	} else {
		result.DirLog = filepath.Join(outTrimmed, "log")
		result.DirOutput = outTrimmed
		result.DirVar = filepath.Join(outTrimmed, "var")
	}

	return result
}

func NewRunLayout() *InitLayout {
	return NewInitLayout("run/in", "run/{timestamp}/out")
}

type InitState struct {
	initTracking
	state *task.State
}

func (is *InitState) AddParams(params map[string]string) {
	maps.Copy(is.params, params)
	is.state.Internal = is.initTracking
}

func (is *InitState) DefaultInput(value string) string {
	if value == "" {
		return is.params["input"]
	}

	return value
}

func (is *InitState) DefaultLog(value string) string {
	if value == "" {
		return is.params["log"]
	} else {
		return NewInitLayout("", value).DirLog
	}
}

func (is *InitState) DefaultOutput(value string) string {
	if value == "" {
		return is.params["output"]
	} else {
		return NewInitLayout("", value).DirOutput
	}
}

func (is *InitState) DefaultVar(value string) string {
	if value == "" {
		return is.params["var"]
	} else {
		return NewInitLayout("", value).DirVar
	}
}

func (is *InitState) GetConfigFile() string {
	return is.configFile
}

func (is *InitState) SetConfigFile(configFile string) {
	is.configFile = configFile
	is.state.Internal = is.initTracking
}

func NewInitState(state *task.State) *InitState {
	var current, ok = state.Internal.(initTracking)
	var initParams = map[string]string{}
	var configFile string

	if ok {
		maps.Copy(initParams, current.params)
		configFile = current.configFile
	}

	return &InitState{
		initTracking: initTracking{
			params:     initParams,
			configFile: configFile,
		},
		state: state,
	}
}
