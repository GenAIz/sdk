package layout

import (
	"maps"

	"genaiz.com/genaiz/lang/enumz"
	"genaiz.com/genaiz/task"
)

type ArchType = string
type FunctionType = string

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
)

type initTracking struct {
	configFile string
	params     map[string]string
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

func (is *InitState) DefaultOutput(value string) string {
	if value == "" {
		return is.params["output"]
	}

	return value
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
