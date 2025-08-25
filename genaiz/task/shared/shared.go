package shared

import (
	"errors"
	"path/filepath"

	"genaiz.com/genaiz/lang/enumz"
)

const (
	ConfigTypeJson ConfigType = "json"
	ConfigTypeNone ConfigType = ""
	ConfigTypeToml ConfigType = "toml"
	ConfigTypeYaml ConfigType = "yaml"
)

var (
	ConfigTypes = enumz.NewEnumType(ConfigTypeJson, ConfigTypeNone, ConfigTypeToml, ConfigTypeYaml)

	ErrorConflict  = errors.New("conflicting version signatures")
	ErrorDuplicate = errors.New("duplicate version signatures")
	ErrorNoChanges = errors.New("no changes needed")
)

type ConfigType = string

// Identity is a shared data type which applies to an entity in a remote or local system. It's made to compare signatures between different sources. A source will typically need an Auth string to give access to the entity located on the provided Path and then return its Hash.
type Identity struct {
	Id      string // Id if there is a singular key integer to access the resource, it should be accessible for further requests
	Hash    string // Hash is a sha256 string which represents the signature of the Identity
	Auth    string // Auth is the base64 encoded Private Authentication Token to access the Identity
	Path    string // Path is a URL string representing the location of the Identity
	Version string // Version is a revision string for comparing the same entities at different revisions
}

func (i Identity) HasIdentifier() bool {
	return i.Id != ""
}

func (i Identity) HasRepoIdentifier() bool {
	return i.HasIdentifier() && i.Hash != ""
}

type ConfigParams struct {
	ConfigName string
	ConfigType *ConfigType
}

// GetConfigFile returns the file path of the config file described by the params
func (cp ConfigParams) GetConfigFile(paths ...string) string {
	var filePaths []string

	filePaths = append(filePaths, paths...)
	filePaths = append(filePaths, cp.ConfigName+"."+*cp.ConfigType)
	return filepath.Join(filePaths...)
}

// IsConfigTypeNone indicates whether a value is equivalent to no config
func (cp ConfigParams) IsConfigTypeNone() bool {
	return cp.ConfigType == nil || *cp.ConfigType == ConfigTypeNone
}
