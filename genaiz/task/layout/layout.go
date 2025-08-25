package layout

import (
	"genaiz.com/genaiz/lang/enumz"
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
