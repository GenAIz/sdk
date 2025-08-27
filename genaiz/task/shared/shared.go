package shared

import (
	"errors"
	"os"
	"path/filepath"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/filez"
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

	ErrorConfigFileExists  = errors.New("config file exists")
	ErrorConfigFileInvalid = errors.New("config file is invalid")
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
	ConfigName   string
	ConfigType   *ConfigType
	ConfigFolder string
}

// GetConfigFile returns the file path of the config file described by the params
func (cp ConfigParams) GetConfigFile(paths ...string) string {
	var filePaths []string
	var typeString string

	if cp.ConfigType != nil {
		typeString = "." + *cp.ConfigType
	}

	filePaths = append(filePaths, paths...)
	filePaths = append(filePaths, cp.ConfigName+typeString)
	return filepath.Join(filePaths...)
}

// GetConfigPath returns the file path of the config file described by the params.
func (cp ConfigParams) GetConfigPath() string {
	return cp.GetConfigFile(cp.ConfigFolder)
}

// IsConfigTypeNone indicates whether a value is equivalent to no config
func (cp ConfigParams) IsConfigTypeNone() bool {
	return cp.ConfigType == nil || *cp.ConfigType == ConfigTypeNone
}

// ResolveConfigPath will return the path of an existing config file with a corresponding found error it the file already exists, otherwise it'll return the path without any errors. The method can return errors of invalid paths.
func (cp ConfigParams) ResolveConfigPath() (string, error) {
	if cp.IsConfigTypeNone() {
		var reset func()
		var err error

		if reset, err = dirz.ChangeWorkingDir(cp.ConfigFolder); err == nil {
			defer reset()
			var file string

			if file, err = filez.FirstNamedFile(cp.GetConfigFile()); err == nil {
				return filepath.Join(cp.ConfigFolder, file), ErrorConfigFileExists
			}
		}

		return "", err
	} else {
		var result = cp.GetConfigPath()

		if info, err := os.Stat(result); err == nil {
			if info.IsDir() {
				return "", ErrorConfigFileInvalid
			}

			return result, ErrorConfigFileExists
		}

		return result, nil
	}
}
