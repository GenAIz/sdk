// Package schema provides schematic elements for validating Genaiz.yaml files. It also provides a central structure for all configuration key constants used by the genaiz toolkits.
package schema

import (
	_ "embed"
	"errors"
	"strings"

	"github.com/spf13/viper"
)

var (
	Genaiz = &Document{}
	//go:embed genaiz.json
	GenaizSchema []byte
)

// Document is the registry containing all Keys used by the genaiz commands
type Document struct {
	Account struct {
		Activate struct {
			Username Keys
		}
		List struct {
			Printer Keys
		}
		Login struct {
			Password Keys
			Refresh  Keys
			Username Keys
		}
		Logout struct {
			Username Keys
		}
	}
	DataLink struct {
		Create struct {
			ConfigType  Keys
			Description Keys
			Handle      Keys
			Name        Keys
			Oem         Keys
			UserDefined Keys
			Version     Keys
		}
		List struct {
			Account     Keys
			AccountOnly Keys
			Printer     Keys
		}
		PropSpecAdd struct {
			ConfigType   Keys
			DefaultValue Keys
			Description  Keys
			EnumValue    Keys
			Handle       Keys
			Name         Keys
			Oem          Keys
			Secret       Keys
			Type         Keys
			UserDefined  Keys
			Version      Keys
		}
		PropSpecEdit struct {
			ConfigType      Keys
			DefaultValue    Keys
			Description     Keys
			EnumAddValue    Keys
			EnumRemoveValue Keys
			EnumValue       Keys
			Handle          Keys
			Name            Keys
			Oem             Keys
			UserDefined     Keys
			Version         Keys
		}
		PropSpecRemove struct {
			ConfigType  Keys
			Handle      Keys
			Name        Keys
			Oem         Keys
			UserDefined Keys
			Version     Keys
		}
		ProxyAdd struct {
			ConfigType  Keys
			Handle      Keys
			Oem         Keys
			Tcp         Keys
			Udp         Keys
			UserDefined Keys
			Version     Keys
		}
		ProxyRm struct {
			ConfigType  Keys
			Handle      Keys
			Oem         Keys
			UserDefined Keys
			Version     Keys
		}
		Publish struct {
			ConfigType       Keys
			Handle           Keys
			Oem              Keys
			PublishedVersion Keys
			UserDefined      Keys
			Version          Keys
		}
		Sync struct {
			ConfigType  Keys
			Handle      Keys
			Oem         Keys
			Sequence    Keys
			UserDefined Keys
			Version     Keys
		}
	}
	Function struct {
		Build struct {
			Context       Keys
			File          Keys
			Label         Keys
			LegacyBuilder Keys
			NoCache       Keys
			Platform      Keys
			Prune         Keys
			Repository    Keys
			Version       Keys
		}
		Create struct {
			Arches       Keys
			ConfigType   Keys
			Description  Keys
			Handle       Keys
			MountInput   Keys
			MountOutput  Keys
			Name         Keys
			Oem          Keys
			Recipe       Keys
			Repository   Keys
			SolutionPath Keys
			Type         Keys
			Version      Keys
		}
		Env struct {
			Context    Keys
			File       Keys
			NoPropSync Keys
		}
		Init struct {
			Arches       Keys
			ConfigType   Keys
			Description  Keys
			Handle       Keys
			MountInput   Keys
			MountOutput  Keys
			Name         Keys
			Oem          Keys
			SolutionPath Keys
			Type         Keys
			Version      Keys
		}
		Publish struct {
			DataPortAdd struct {
				Input struct {
					Desc Keys
					Name Keys
				}
				Output struct {
					Desc Keys
					Name Keys
				}
			}
			DataSourceAdd struct {
				Handle       Keys
				NoValidation Keys
				Oem          Keys
				Version      Keys
			}
			DataSourceRemove struct {
				Handle  Keys
				Oem     Keys
				Version Keys
			}
			DataStoreAdd struct {
				Handle       Keys
				NoValidation Keys
				Oem          Keys
				Version      Keys
			}
			DataStoreRemove struct {
				Handle  Keys
				Oem     Keys
				Version Keys
			}
			OutboundProxyAdd struct {
				Inactive Keys
				Tcp      Keys
				Udp      Keys
			}
			PropSpecAdd struct {
				DefaultValue Keys
				Description  Keys
				EnumValue    Keys
				Name         Keys
				Type         Keys
			}
			PropSpecEdit struct {
				DefaultValue    Keys
				Description     Keys
				EnumAddValue    Keys
				EnumRemoveValue Keys
				EnumValue       Keys
				Name            Keys
			}
			Arches          Keys
			Account         Keys
			DataSources     Keys
			DataStores      Keys
			Description     Keys
			Extras          Keys
			Handle          Keys
			InputPorts      Keys
			Internal        Keys
			Name            Keys
			NoUpdate        Keys
			Oem             Keys
			OutboundProxies Keys
			OutputPorts     Keys
			Printer         Keys
			PropSpecs       Keys
			Rebuild         Keys
			ResultValues    Keys
			Type            Keys
			Version         Keys
		}
		Run struct {
			EnvFile     Keys
			EnvVars     Keys
			Image       Keys
			MountInput  Keys
			MountLog    Keys
			MountOutput Keys
			MountVar    Keys
			NoPropSync  Keys
			Prefix      Keys
		}
		Start struct {
			EnvFile     Keys
			EnvVars     Keys
			Image       Keys
			MountInput  Keys
			MountLog    Keys
			MountOutput Keys
			MountVar    Keys
			Name        Keys
			NoPropSync  Keys
			Prefix      Keys
			Preserve    Keys
			Replace     Keys
		}
		Stop struct {
			Image    Keys
			Name     Keys
			Prefix   Keys
			Preserve Keys
		}
		Test struct {
			EnvFile     Keys
			EnvVars     Keys
			Image       Keys
			MountInput  Keys
			MountLog    Keys
			MountOutput Keys
			MountVar    Keys
			NoPropSync  Keys
			Prefix      Keys
		}
	}
	Solution struct {
		Create struct {
			ConfigType  Keys
			Description Keys
			Handle      Keys
			Name        Keys
			Oem         Keys
			Version     Keys

			Workflow struct {
				Handle      Keys
				Name        Keys
				Description Keys
			}
		}
		List struct {
			Account     Keys
			AccountOnly Keys
			Printer     Keys
		}
		Log struct {
			Format Keys
			Level  Keys
		}
		Publish struct {
			Account     Keys
			ConfigType  Keys
			Description Keys
			Handle      Keys
			Name        Keys
			Oem         Keys
			Printer     Keys
			Version     Keys
		}
	}
	Workflow struct {
		Create struct {
			ConfigType  Keys
			Description Keys
			Name        Keys
		}
		Delete struct {
			ConfigType Keys
		}
		Links struct {
			Add struct {
				ConfigType   Keys
				NoValidation Keys
			}
			Remove struct {
				ConfigType   Keys
				NoValidation Keys
			}
		}
		List struct {
			Account Keys
			Printer Keys
		}
		Nodes struct {
			Add struct {
				ConfigType   Keys
				Description  Keys
				Deserialized Keys
				Handle       Keys
				Name         Keys
				Oem          Keys
				Sequence     Keys
				Serialized   Keys
				Version      Keys
			}
			Remove struct {
				ConfigType Keys
			}
		}
		Props struct {
			Add struct {
				NoSync       Keys
				NoValidation Keys
			}
			Edit struct {
				NoSync       Keys
				NoValidation Keys
			}
			List struct {
				NoSync Keys
			}
		}
	}
	Workspace struct {
		Create struct {
			Account     Keys
			Description Keys
			Printer     Keys
			RcEnabled   Keys
			Visibility  Keys
		}
		Flow struct {
			Create struct {
				Account     Keys
				Description Keys
				Name        Keys
				Printer     Keys
			}
		}
		List struct {
			Account     Keys
			DateMonthly Keys
			DateToday   Keys
			DateWeekly  Keys
			OwnerOnly   Keys
			Printer     Keys
			RcEnabled   Keys
		}
	}
}

// Keys describe a structure of string used to refer to a specific option or value read by the genaiz commands
type Keys struct {
	Doc        string   // Doc is the key a user should expect to see if the value is specified under a structured document
	Env        string   // Env is the key a user could use to specify the value under the environment of execution
	Pseudonyms []string // Pseudonyms is a list of alternate keys, which may refer to Doc in a structured document. Pseudonyms can be used to support key migrations between versions
}

// GetString will look for a string value in the provided viper.Viper registry using the Keys' Doc and Pseudonyms values. If no value is found it returns the provided defaultValue list, merged.
func (k Keys) GetString(viper *viper.Viper, defaultValue ...string) string {
	var result string

	if result = viper.GetString(k.Doc); result == "" {
		for _, pseudo := range k.Pseudonyms {
			if result = viper.GetString(pseudo); result != "" {
				return result
			}
		}
	} else {
		return result
	}

	return strings.Join(defaultValue, "")
}

// Unmarshall will look for a value in the provided viper.Viper registry using the Keys' Doc and Pseudonyms values.
func (k Keys) Unmarshall(viper *viper.Viper, ref any) error {
	var value any

	if value = viper.Get(k.Doc); value == nil {
		for _, pseudo := range k.Pseudonyms {
			if value = viper.Get(pseudo); value != nil {
				return viper.UnmarshalKey(pseudo, ref)
			}
		}
	} else {
		return viper.UnmarshalKey(k.Doc, &ref)
	}

	return errors.New("not found")
}

// Normalize is a utility method used to migrate configurations using pseudonym prefixes to long key formats
func Normalize(vp *viper.Viper) *viper.Viper {
	var result = viper.New()
	var merging = make(map[string]any)

	for _, key := range vp.AllKeys() {
		switch strings.ToLower(key)[0:3] {
		case "ac.":
			merging["account"+key[2:]] = vp.Get(key)
			break
		case "sf.":
			merging["function"+key[2:]] = vp.Get(key)
			break
		case "sn.":
			merging["solution"+key[2:]] = vp.Get(key)
			break
		case "wf.":
			merging["workflow"+key[2:]] = vp.Get(key)
			break
		default:
			result.Set(key, vp.Get(key))
		}
	}

	for k, v := range merging {
		if result.Get(k) == nil {
			result.Set(k, v)
		}
	}

	return result
}

func init() {
	Genaiz.Account.Activate.Username = newKeys("Account.Activate.Username", "AC_ACTIVATE_USERNAME", "Ac.Activate.Username")

	Genaiz.Account.List.Printer = newKeys("Account.List.Printer", "AC_LIST_PRINTER", "Ac.List.Printer")

	Genaiz.Account.Login.Password = newKeys("p", "GENAIZ_PASSWORD")
	Genaiz.Account.Login.Refresh = newKeys("Account.Login.Refresh", "AC_LOGIN_REFRESH", "Ac.Login.Refresh")
	Genaiz.Account.Login.Username = newKeys("Account.Login.Username", "GENAIZ_USERNAME", "Ac.Login.Username")

	Genaiz.Account.Logout.Username = newKeys("Account.Logout.Username", "GENAIZ_USERNAME", "Ac.Logout.Username")

	Genaiz.DataLink.Create.ConfigType = newKeys("DataLink.Create.ConfigType", "DK_CREATE_CONFIG_TYPE", "Dk.Create.ConfigType")
	Genaiz.DataLink.Create.Description = newKeys("DataLink.Create.Description", "DK_CREATE_DESC", "Dk.Create.Description")
	Genaiz.DataLink.Create.Handle = newKeys("DataLink.Create.Handle", "")
	Genaiz.DataLink.Create.Name = newKeys("DataLink.Create.Name", "DK_CREATE_NAME", "Dk.Create.Name")
	Genaiz.DataLink.Create.Oem = newKeys("DataLink.Create.Oem", "DK_CREATE_OEM", "Dk.Create.Oem")
	Genaiz.DataLink.Create.UserDefined = newKeys("DataLink.Create.UserDefined", "DK_CREATE_USER_DEFINED", "Dk.Create.UserDefined")
	Genaiz.DataLink.Create.Version = newKeys("DataLink.Create.Version", "DK_CREATE_VERSION", "Dk.Create.Version")

	Genaiz.DataLink.List.Account = newKeys("DataLink.List.Account", "DK_LIST_ACCOUNT", "Dk.List.Account")
	Genaiz.DataLink.List.AccountOnly = newKeys("DataLink.List.AccountOnly", "DK_LIST_ACCOUNT_ONLY", "Dk.List.AccountOnly")
	Genaiz.DataLink.List.Printer = newKeys("DataLink.List.Printer", "DK_LIST_PRINTER", "Dk.List.Printer")

	Genaiz.DataLink.PropSpecAdd.ConfigType = newKeys("DataLink.PropSpecAdd.ConfigType", "DK_CREATE_PROP_SPEC_ADD_CONFIG_TYPE", "Dk.PropSpecAdd.ConfigType")
	Genaiz.DataLink.PropSpecAdd.DefaultValue = newKeys("DataLink.PropSpecAdd.DefaultValue", "DK_PROP_SPEC_ADD_DEFAULT_VALUE", "Dk.PropSpecAdd.DefaultValue")
	Genaiz.DataLink.PropSpecAdd.Description = newKeys("DataLink.PropSpecAdd.Description", "DK_PROP_SPEC_ADD_DESCRIPTION", "Dk.PropSpecAdd.Description")
	Genaiz.DataLink.PropSpecAdd.EnumValue = newKeys("DataLink.PropSpecAdd.EnumValue", "DK_PROP_SPEC_ADD_ENUM_VALUE", "Dk.PropSpecAdd.EnumValue")
	Genaiz.DataLink.PropSpecAdd.Handle = newKeys("DataLink.PropSpecAdd.Handle", "DK_PROP_SPEC_ADD_HANDLE", "Dk.PropSpecAdd.Handle")
	Genaiz.DataLink.PropSpecAdd.Name = newKeys("DataLink.PropSpecAdd.Name", "DK_PROP_SPEC_ADD_NAME", "Dk.PropSpecAdd.Name")
	Genaiz.DataLink.PropSpecAdd.Oem = newKeys("DataLink.PropSpecAdd.Oem", "DK_PROP_SPEC_ADD_OEM", "Dk.PropSpecAdd.Oem")
	Genaiz.DataLink.PropSpecAdd.Secret = newKeys("DataLink.PropSpecAdd.Secret", "DK_PROP_SPEC_ADD_SECRET", "Dk.PropSpecAdd.Secret")
	Genaiz.DataLink.PropSpecAdd.Type = newKeys("DataLink.PropSpecAdd.Type", "DK_PROP_SPEC_ADD_TYPE", "Dk.PropSpecAdd.Type")
	Genaiz.DataLink.PropSpecAdd.UserDefined = newKeys("DataLink.PropSpecAdd.UserDefined", "DK_PROP_SPEC_ADD_USER_DEFINED", "Dk.PropSpecAdd.UserDefined")
	Genaiz.DataLink.PropSpecAdd.Version = newKeys("DataLink.PropSpecAdd.Version", "DK_PROP_SPEC_ADD_VERSION", "Dk.PropSpecAdd.Version")

	Genaiz.DataLink.PropSpecEdit.ConfigType = newKeys("DataLink.PropSpecEdit.ConfigType", "DK_CREATE_PROP_SPEC_EDIT_CONFIG_TYPE", "Dk.PropSpecEdit.ConfigType")
	Genaiz.DataLink.PropSpecEdit.DefaultValue = newKeys("DataLink.PropSpecEdit.DefaultValue", "DK_PROP_SPEC_EDIT_DEFAULT_VALUE", "Dk.PropSpecEdit.DefaultValue")
	Genaiz.DataLink.PropSpecEdit.Description = newKeys("DataLink.PropSpecEdit.Description", "DK_PROP_SPEC_EDIT_DESCRIPTION", "Dk.PropSpecEdit.Description")
	Genaiz.DataLink.PropSpecEdit.EnumValue = newKeys("DataLink.PropSpecEdit.EnumValue", "DK_PROP_SPEC_EDIT_ENUM_VALUE", "Dk.PropSpecEdit.EnumValue")
	Genaiz.DataLink.PropSpecEdit.EnumAddValue = newKeys("DataLink.PropSpecEdit.EnumAddValue", "SF_PUBLISH_PROP_SPEC_EDIT_ENUM_ADD_VALUE", "Dk.PropSpecEdit.EnumAddValue")
	Genaiz.DataLink.PropSpecEdit.EnumRemoveValue = newKeys("DataLink.PropSpecEdit.EnumRemoveValue", "SF_PUBLISH_PROP_SPEC_EDIT_ENUM_RM_VALUE", "Dk.PropSpecEdit.EnumRemoveValue")
	Genaiz.DataLink.PropSpecEdit.Handle = newKeys("DataLink.PropSpecEdit.Handle", "DK_PROP_SPEC_EDIT_HANDLE", "Dk.PropSpecEdit.Handle")
	Genaiz.DataLink.PropSpecEdit.Name = newKeys("DataLink.PropSpecEdit.Name", "DK_PROP_SPEC_EDIT_NAME", "Dk.PropSpecEdit.Name")
	Genaiz.DataLink.PropSpecEdit.Oem = newKeys("DataLink.PropSpecEdit.Oem", "DK_PROP_SPEC_EDIT_OEM", "Dk.PropSpecEdit.Oem")
	Genaiz.DataLink.PropSpecEdit.UserDefined = newKeys("DataLink.PropSpecEdit.UserDefined", "DK_PROP_SPEC_EDIT_USER_DEFINED", "Dk.PropSpecEdit.UserDefined")
	Genaiz.DataLink.PropSpecEdit.Version = newKeys("DataLink.PropSpecEdit.Version", "DK_PROP_SPEC_EDIT_VERSION", "Dk.PropSpecEdit.Version")

	Genaiz.DataLink.PropSpecRemove.ConfigType = newKeys("DataLink.PropSpecRemove.ConfigType", "DK_CREATE_PROP_SPEC_RM_CONFIG_TYPE", "Dk.PropSpecRemove.ConfigType")
	Genaiz.DataLink.PropSpecRemove.Handle = newKeys("DataLink.PropSpecRemove.Handle", "DK_PROP_SPEC_RM_HANDLE", "Dk.PropSpecRemove.Handle")
	Genaiz.DataLink.PropSpecRemove.Oem = newKeys("DataLink.PropSpecRemove.Oem", "DK_PROP_SPEC_RM_OEM", "Dk.PropSpecRemove.Oem")
	Genaiz.DataLink.PropSpecRemove.UserDefined = newKeys("DataLink.PropSpecRemove.UserDefined", "DK_PROP_SPEC_RM_USER_DEFINED", "Dk.PropSpecRemove.UserDefined")
	Genaiz.DataLink.PropSpecRemove.Version = newKeys("DataLink.PropSpecRemove.Version", "DK_PROP_SPEC_RM_VERSION", "Dk.PropSpecRemove.Version")

	Genaiz.DataLink.ProxyAdd.ConfigType = newKeys("DataLink.ProxyAdd.ConfigType", "DK_PROXY_ADD_CONFIG_TYPE", "Dk.ProxyAdd.ConfigType")
	Genaiz.DataLink.ProxyAdd.Handle = newKeys("DataLink.ProxyAdd.Handle", "DK_PROXY_ADD_HANDLE", "Dk.ProxyAdd.Handle")
	Genaiz.DataLink.ProxyAdd.Oem = newKeys("DataLink.ProxyAdd.Oem", "DK_PROXY_ADD_OEM", "Dk.ProxyAdd.Oem")
	Genaiz.DataLink.ProxyAdd.Tcp = newKeys("DataLink.ProxyAdd.Tcp", "DK_PROXY_ADD_TCP", "Dk.ProxyAdd.Tcp")
	Genaiz.DataLink.ProxyAdd.Udp = newKeys("DataLink.ProxyAdd.Udp", "DK_PROXY_ADD_UDP", "Dk.ProxyAdd.Udp")
	Genaiz.DataLink.ProxyAdd.UserDefined = newKeys("DataLink.ProxyAdd.UserDefined", "DK_PROXY_ADD_USER_DEFINED", "Dk.ProxyAdd.UserDefined")
	Genaiz.DataLink.ProxyAdd.Version = newKeys("DataLink.ProxyAdd.Version", "DK_PROXY_ADD_VERSION", "Dk.ProxyAdd.Version")

	Genaiz.DataLink.ProxyRm.ConfigType = newKeys("DataLink.ProxyRm.ConfigType", "DK_PROXY_RM_CONFIG_TYPE", "Dk.ProxyRm.ConfigType")
	Genaiz.DataLink.ProxyRm.Handle = newKeys("DataLink.ProxyRm.Handle", "DK_PROXY_RM_HANDLE", "Dk.ProxyRm.Handle")
	Genaiz.DataLink.ProxyRm.Oem = newKeys("DataLink.ProxyRm.Oem", "DK_PROXY_RM_OEM", "Dk.ProxyRm.Oem")
	Genaiz.DataLink.ProxyRm.UserDefined = newKeys("DataLink.ProxyRm.UserDefined", "DK_PROXY_RM_USER_DEFINED", "Dk.ProxyRm.UserDefined")
	Genaiz.DataLink.ProxyRm.Version = newKeys("DataLink.ProxyRm.Version", "DK_PROXY_RM_VERSION", "Dk.ProxyRm.Version")

	Genaiz.DataLink.Publish.ConfigType = newKeys("DataLink.Publish.ConfigType", "DK_PUBLISH_CONFIG_TYPE", "Dk.Publish.ConfigType")
	Genaiz.DataLink.Publish.Handle = newKeys("DataLink.Publish.Handle", "DK_PUBLISH_HANDLE", "Dk.Publish.Handle")
	Genaiz.DataLink.Publish.Oem = newKeys("DataLink.Publish.Oem", "DK_PUBLISH_OEM", "Dk.Publish.Oem")
	Genaiz.DataLink.Publish.UserDefined = newKeys("DataLink.Publish.UserDefined", "DK_PUBLISH_USER_DEFINED", "Dk.Publish.UserDefined")
	Genaiz.DataLink.Publish.PublishedVersion = newKeys("DataLink.Publish.PublishedVersion", "DK_PUBLISH_PUBLISHED_VERSION", "Dk.Publish.PublishedVersion")
	Genaiz.DataLink.Publish.Version = newKeys("DataLink.Publish.Version", "DK_PUBLISH_VERSION", "Dk.Publish.Version")

	Genaiz.DataLink.Sync.ConfigType = newKeys("DataLink.Sync.ConfigType", "DK_SYNC_CONFIG_TYPE", "Dk.Sync.ConfigType")
	Genaiz.DataLink.Sync.Handle = newKeys("DataLink.Sync.Handle", "DK_SYNC_HANDLE", "Dk.Sync.Handle")
	Genaiz.DataLink.Sync.Oem = newKeys("DataLink.Sync.Oem", "DK_SYNC_OEM", "Dk.Sync.Oem")
	Genaiz.DataLink.Sync.Sequence = newKeys("DataLink.Sync.Sequence", "DK_SYNC_SEQUENCE", "Dk.Sync.Sequence")
	Genaiz.DataLink.Sync.UserDefined = newKeys("DataLink.Sync.UserDefined", "DK_SYNC_USER_DEFINED", "Dk.Sync.UserDefined")
	Genaiz.DataLink.Sync.Version = newKeys("DataLink.Sync.Version", "DK_SYNC_VERSION", "Dk.Sync.Version")

	Genaiz.Function.Build.Context = newKeys("Function.Build.Context", "SF_BUILD_CONTEXT", "Sf.Build.Context")
	Genaiz.Function.Build.File = newKeys("Function.Build.File", "SF_BUILD_FILE", "Sf.Build.File")
	Genaiz.Function.Build.Label = newKeys("Function.Build.Label", "SF_BUILD_LABEL", "Sf.Build.Label")
	Genaiz.Function.Build.LegacyBuilder = newKeys("Function.Build.LegacyBuilder", "SF_BUILD_LEGACY", "Sf.Build.LegacyBuilder")
	Genaiz.Function.Build.NoCache = newKeys("Function.Build.NoCache", "SF_BUILD_NOCACHE", "Sf.Build.NoCache")
	Genaiz.Function.Build.Platform = newKeys("Function.Build.Platform", "SF_BUILD_PLATFORM", "Sf.Build.Platform")
	Genaiz.Function.Build.Prune = newKeys("Function.Build.Prune", "SF_BUILD_PRUNE", "Sf.Build.Prune")
	Genaiz.Function.Build.Repository = newKeys("Function.Build.Repository", "SF_BUILD_REPOSITORY", "Sf.Build.Tag", "Function.Build.Tag", "Sf.Build.Repository")
	Genaiz.Function.Build.Version = newKeys("Function.Build.Version", "SF_BUILD_VERSION", "Sf.Build.Version")

	Genaiz.Function.Create.Arches = newKeys("Function.Create.Arches", "SF_CREATE_ARCHES", "Sf.Create.Arches")
	Genaiz.Function.Create.ConfigType = newKeys("Function.Create.ConfigType", "SF_CREATE_CONFIG_TYPE", "Sf.Create.ConfigType")
	Genaiz.Function.Create.Handle = newKeys("Function.Create.Handle", "SF_CREATE_HANDLE", "Sf.Create.Handle")
	Genaiz.Function.Create.MountInput = newKeys("Function.Create.Input", "SF_CREATE_MOUNT_INPUT", "Sf.Create.Input")
	Genaiz.Function.Create.MountOutput = newKeys("Function.Create.Output", "SF_CREATE_MOUNT_OUTPUT", "Sf.Create.Output")
	Genaiz.Function.Create.Name = newKeys("Function.Create.Name", "SF_CREATE_NAME", "Sf.Create.Name")
	Genaiz.Function.Create.Oem = newKeys("Function.Create.Oem", "SF_CREATE_OEM", "Sf.Create.Oem")
	Genaiz.Function.Create.Recipe = newKeys("Function.Create.Recipe", "SF_CREATE_RECIPE", "Sf.Create.Recipe")
	Genaiz.Function.Create.Repository = newKeys("Function.Create.Repository", "SF_CREATE_REPOSITORY", "Sf.Create.Repository", "Function.Create.Tag", "Sf.Create.Repository")
	Genaiz.Function.Create.SolutionPath = newKeys("Function.Create.SolutionPath", "SF_CREATE_SN_PATH", "Sf.Create.SolutionPath")
	Genaiz.Function.Create.Type = newKeys("Function.Create.Type", "SF_CREATE_TYPE", "Sf.Create.Type")
	Genaiz.Function.Create.Version = newKeys("Function.Create.Version", "SF_CREATE_VERSION", "Sf.Create.Version")

	Genaiz.Function.Env.Context = newKeys("Function.Env.Context", "SF_ENV_CONTEXT", "Sf.Env.Context")
	Genaiz.Function.Env.File = newKeys("Function.Env.File", "SF_ENV_FILE", "Sf.Env.File")
	Genaiz.Function.Env.NoPropSync = newKeys("Function.Env.NoPropSync", "SF_ENV_NO_PROP_SYNC", "Sf.Env.NoPropSync")

	Genaiz.Function.Init.Arches = newKeys("Function.Init.Arches", "SF_INIT_ARCHES", "Sf.Init.Arches")
	Genaiz.Function.Init.ConfigType = newKeys("Function.Init.ConfigType", "SF_INIT_CONFIG_TYPE", "Sf.Init.ConfigType")
	Genaiz.Function.Init.Handle = newKeys("Function.Init.Handle", "SF_INIT_HANDLE", "Sf.Init.Handle")
	Genaiz.Function.Init.MountInput = newKeys("Function.Init.Input", "SF_INIT_MOUNT_INPUT", "Sf.Init.Input")
	Genaiz.Function.Init.MountOutput = newKeys("Function.Init.Output", "SF_INIT_MOUNT_OUTPUT", "Sf.Init.Output")
	Genaiz.Function.Init.Name = newKeys("Function.Init.Name", "SF_INIT_NAME", "Sf.Init.Name")
	Genaiz.Function.Init.Oem = newKeys("Function.Init.Oem", "SF_INIT_OEM", "Sf.Init.Oem")
	Genaiz.Function.Init.SolutionPath = newKeys("Function.Init.SolutionPath", "SF_INIT_SN_PATH", "Sf.Init.SolutionPath")
	Genaiz.Function.Init.Type = newKeys("Function.Init.Type", "SF_INIT_TYPE", "Sf.Init.Type")
	Genaiz.Function.Init.Version = newKeys("Function.Init.Version", "SF_INIT_VERSION", "Sf.Init.Version")

	Genaiz.Function.Publish.Account = newKeys("Function.Publish.Account", "SF_PUBLISH_ACCOUNT", "Sf.Publish.Account")
	Genaiz.Function.Publish.Arches = newKeys("Function.Publish.Arches", "SF_PUBLISH_ARCHES", "Sf.Publish.Arches")
	Genaiz.Function.Publish.DataSources = newKeys("Function.Publish.DataSources", "SF_PUBLISH_DATA_SOURCES", "Sf.Publish.DataSources")
	Genaiz.Function.Publish.DataStores = newKeys("Function.Publish.DataStores", "SF_PUBLISH_DATA_STORES", "Sf.Publish.DataStores")
	Genaiz.Function.Publish.Description = newKeys("Function.Publish.Description", "SF_PUBLISH_DESCRIPTION", "Sf.Publish.Description")
	Genaiz.Function.Publish.Extras = newKeys("Function.Publish.Extras", "SF_PUBLISH_EXTRAS", "Sf.Publish.Extras")
	Genaiz.Function.Publish.Handle = newKeys("Function.Publish.Handle", "SF_PUBLISH_HANDLE", "Sf.Publish.Handle")
	Genaiz.Function.Publish.Internal = newKeys("Function.Publish", "", "Sf.Publish")
	Genaiz.Function.Publish.InputPorts = newKeys("Function.Publish.InputPorts", "", "Sf.Publish.InputPorts")
	Genaiz.Function.Publish.Name = newKeys("Function.Publish.Name", "SF_PUBLISH_NAME", "Sf.Publish.Name")
	Genaiz.Function.Publish.NoUpdate = newKeys("Function.Publish.NoUpdate", "SF_PUBLISH_NO_UPDATE", "Sf.Publish.NoUpdate")
	Genaiz.Function.Publish.Oem = newKeys("Function.Publish.Oem", "SF_PUBLISH_OEM", "Sf.Publish.Oem")
	Genaiz.Function.Publish.OutputPorts = newKeys("Function.Publish.OutputPorts", "", "Sf.Publish.OutputPorts")
	Genaiz.Function.Publish.OutboundProxies = newKeys("Function.Publish.OutboundProxies", "", "Sf.Publish.OutboundProxies")
	Genaiz.Function.Publish.Printer = newKeys("Function.List.Printer", "SF_PUBLISH_PRINTER", "Sf.Publish.Printer")
	Genaiz.Function.Publish.PropSpecs = newKeys("Function.Publish.PropSpecs", "", "Sf.Publish.PropSpecs")
	Genaiz.Function.Publish.Rebuild = newKeys("Function.Publish.Rebuild", "SF_PUBLISH_REBUILD", "Sf.Publish.Rebuild")
	Genaiz.Function.Publish.ResultValues = newKeys("Function.Publish.ResultValues", "SF_PUBLISH_RESULT_VALUES", "Sf.Publish.ResultValues")
	Genaiz.Function.Publish.Type = newKeys("Function.Publish.Type", "SF_PUBLISH_TYPE", "Sf.Publish.Type")
	Genaiz.Function.Publish.Version = newKeys("Function.Publish.Version", "SF_PUBLISH_VERSION", "Sf.Publish.Version")

	Genaiz.Function.Publish.DataSourceAdd.Handle = newKeys("Function.Publish.DataSourceAdd.Handle", "SF_PUBLISH_DATA_SRC_ADD_HANDLE", "Sf.Publish.DataSourceAdd.Handle")
	Genaiz.Function.Publish.DataSourceAdd.NoValidation = newKeys("Function.Publish.DataSourceAdd.NoValidation", "SF_PUBLISH_DATA_SRC_ADD_NO_VALIDATION", "Sf.Publish.DataSourceAdd.NoValidation")
	Genaiz.Function.Publish.DataSourceAdd.Oem = newKeys("Function.Publish.DataSourceAdd.Oem", "SF_PUBLISH_DATA_SRC_ADD_OEM", "Sf.Publish.DataSourceAdd.Oem")
	Genaiz.Function.Publish.DataSourceAdd.Version = newKeys("Function.Publish.DataSourceAdd.Version", "SF_PUBLISH_DATA_SRC_ADD_VERSION", "Sf.Publish.DataSourceAdd.Version")

	Genaiz.Function.Publish.DataSourceRemove.Handle = newKeys("Function.Publish.DataSourceRemove.Handle", "SF_PUBLISH_DATA_RM_ADD_HANDLE", "Sf.Publish.DataSourceRemove.Handle")
	Genaiz.Function.Publish.DataSourceRemove.Oem = newKeys("Function.Publish.DataSourceRemove.Oem", "SF_PUBLISH_DATA_SRC_RM_OEM", "Sf.Publish.DataSourceRemove.Oem")
	Genaiz.Function.Publish.DataSourceRemove.Version = newKeys("Function.Publish.DataSourceRemove.Version", "SF_PUBLISH_DATA_SRC_RM_VERSION", "Sf.Publish.DataSourceRemove.Version")

	Genaiz.Function.Publish.DataStoreAdd.Handle = newKeys("Function.Publish.DataStoreAdd.Handle", "SF_PUBLISH_DATA_STR_ADD_HANDLE", "Sf.Publish.DataStoreAdd.Handle")
	Genaiz.Function.Publish.DataStoreAdd.NoValidation = newKeys("Function.Publish.DataStoreAdd.NoValidation", "SF_PUBLISH_DATA_STR_ADD_NO_VALIDATION", "Sf.Publish.DataStoreAdd.NoValidation")
	Genaiz.Function.Publish.DataStoreAdd.Oem = newKeys("Function.Publish.DataStoreAdd.Oem", "SF_PUBLISH_DATA_STR_ADD_OEM", "Sf.Publish.DataStoreAdd.Oem")
	Genaiz.Function.Publish.DataStoreAdd.Version = newKeys("Function.Publish.DataStoreAdd.Version", "SF_PUBLISH_DATA_STR_ADD_VERSION", "Sf.Publish.DataStoreAdd.Version")

	Genaiz.Function.Publish.DataStoreRemove.Handle = newKeys("Function.Publish.DataStoreRemove.Handle", "SF_PUBLISH_DATA_STR_RM_HANDLE", "Sf.Publish.DataStoreRemove.Handle")
	Genaiz.Function.Publish.DataStoreRemove.Oem = newKeys("Function.Publish.DataStoreRemove.Oem", "SF_PUBLISH_DATA_STR_RM_OEM", "Sf.Publish.DataStoreRemove.Oem")
	Genaiz.Function.Publish.DataStoreRemove.Version = newKeys("Function.Publish.DataStoreRemove.Version", "SF_PUBLISH_DATA_STR_RM_VERSION", "Sf.Publish.DataStoreRemove.Version")

	Genaiz.Function.Publish.DataPortAdd.Input.Desc = newKeys("Function.Publish.DataPortAdd.Input.Desc", "SF_PUBLISH_DATA_PORT_ADD_DESC", "Sf.Publish.DataPortAdd.Input.Desc")
	Genaiz.Function.Publish.DataPortAdd.Input.Name = newKeys("Function.Publish.DataPortAdd.Input.Name", "SF_PUBLISH_DATA_PORT_ADD_NAME", "Sf.Publish.DataPortAdd.Input.Name")
	Genaiz.Function.Publish.DataPortAdd.Output.Desc = newKeys("Function.Publish.DataPortAdd.Output.Desc", "SF_PUBLISH_DATA_PORT_ADD_DESC", "Sf.Publish.DataPortAdd.Output.Desc")
	Genaiz.Function.Publish.DataPortAdd.Output.Name = newKeys("Function.Publish.DataPortAdd.Output.Name", "SF_PUBLISH_DATA_PORT_ADD_NAME", "Sf.Publish.DataPortAdd.Output.Name")

	Genaiz.Function.Publish.OutboundProxyAdd.Inactive = newKeys("Function.Publish.OutboundProxyAdd.Inactive", "SF_PUBLISH_OUTBOUND_PROXY_ADD_INACTIVE", "Sf.Publish.OutboundProxyAdd.Inactive")
	Genaiz.Function.Publish.OutboundProxyAdd.Tcp = newKeys("Function.Publish.OutboundProxyAdd.Tcp", "SF_PUBLISH_OUTBOUND_PROXY_ADD_TCP", "Sf.Publish.OutboundProxyAdd.Tcp")
	Genaiz.Function.Publish.OutboundProxyAdd.Udp = newKeys("Function.Publish.OutboundProxyAdd.Udp", "SF_PUBLISH_OUTBOUND_PROXY_ADD_UDP", "Sf.Publish.OutboundProxyAdd.Udp")

	Genaiz.Function.Publish.PropSpecAdd.DefaultValue = newKeys("Function.Publish.PropSpecAdd.DefaultValue", "SF_PUBLISH_PROP_SPEC_ADD_DEFAULT_VALUE", "Sf.Publish.PropSpecAdd.DefaultValue")
	Genaiz.Function.Publish.PropSpecAdd.Description = newKeys("Function.Publish.PropSpecAdd.Description", "SF_PUBLISH_PROP_SPEC_ADD_DESCRIPTION", "Sf.Publish.PropSpecAdd.Description")
	Genaiz.Function.Publish.PropSpecAdd.EnumValue = newKeys("Function.Publish.PropSpecAdd.EnumValue", "SF_PUBLISH_PROP_SPEC_ADD_ENUM_VALUE", "Sf.Publish.PropSpecAdd.EnumValue")
	Genaiz.Function.Publish.PropSpecAdd.Name = newKeys("Function.Publish.PropSpecAdd.Name", "SF_PUBLISH_PROP_SPEC_ADD_NAME", "Sf.Publish.PropSpecAdd.Name")
	Genaiz.Function.Publish.PropSpecAdd.Type = newKeys("Function.Publish.PropSpecAdd.Type", "SF_PUBLISH_PROP_SPEC_ADD_TYPE", "Sf.Publish.PropSpecAdd.Type")

	Genaiz.Function.Publish.PropSpecEdit.DefaultValue = newKeys("Function.Publish.PropSpecEdit.DefaultValue", "SF_PUBLISH_PROP_SPEC_EDIT_DEFAULT_VALUE", "Sf.Publish.PropSpecEdit.DefaultValue")
	Genaiz.Function.Publish.PropSpecEdit.Description = newKeys("Function.Publish.PropSpecEdit.Description", "SF_PUBLISH_PROP_SPEC_EDIT_DESCRIPTION", "Sf.Publish.PropSpecEdit.Description")
	Genaiz.Function.Publish.PropSpecEdit.EnumAddValue = newKeys("Function.Publish.PropSpecEdit.EnumAddValue", "SF_PUBLISH_PROP_SPEC_EDIT_ENUM_ADD_VALUE", "Sf.Publish.PropSpecEdit.EnumAddValue")
	Genaiz.Function.Publish.PropSpecEdit.EnumRemoveValue = newKeys("Function.Publish.PropSpecEdit.EnumRemoveValue", "SF_PUBLISH_PROP_SPEC_EDIT_ENUM_RM_VALUE", "Sf.Publish.PropSpecEdit.EnumRemoveValue")
	Genaiz.Function.Publish.PropSpecEdit.EnumValue = newKeys("Function.Publish.PropSpecEdit.EnumValue", "SF_PUBLISH_PROP_SPEC_EDIT_ENUM_VALUE", "Sf.Publish.PropSpecEdit.EnumValue")
	Genaiz.Function.Publish.PropSpecEdit.Name = newKeys("Function.Publish.PropSpecEdit.Name", "SF_PUBLISH_PROP_SPEC_EDIT_NAME", "Sf.Publish.PropSpecEdit.Name")

	Genaiz.Function.Run.EnvFile = newKeys("Function.Run.EnvFile", "SF_RUN_ENV_FILE", "Sf.Run.EnvFile")
	Genaiz.Function.Run.EnvVars = newKeys("Function.Run.EnvVar", "SF_RUN_ENV_VAR", "Sf.Run.EnvVar")
	Genaiz.Function.Run.Image = newKeys("Function.Run.Image", "SF_RUN_IMAGE", "Sf.Run.Image")
	Genaiz.Function.Run.MountInput = newKeys("Function.Run.Input", "SF_RUN_MOUNT_INPUT", "Sf.Run.Input")
	Genaiz.Function.Run.MountOutput = newKeys("Function.Run.Output", "SF_RUN_MOUNT_OUTPUT", "Sf.Run.Output")
	Genaiz.Function.Run.MountLog = newKeys("Function.Run.Log", "SF_RUN_MOUNT_LOG", "Sf.Run.Log")
	Genaiz.Function.Run.MountVar = newKeys("Function.Run.Var", "SF_RUN_MOUNT_VAR", "Sf.Run.Var")
	Genaiz.Function.Run.NoPropSync = newKeys("Function.Run.NoPropSync", "SF_RUN_NO_PROP_SYNC", "Sf.Run.NoPropSync")
	Genaiz.Function.Run.Prefix = newKeys("Function.Run.Prefix", "SF_RUN_CONTAINER_PREFIX", "Sf.Run.Prefix")

	Genaiz.Function.Start.EnvFile = newKeys("Function.Start.EnvFile", "SF_RUN_ENV_FILE", "Sf.Start.EnvFile")
	Genaiz.Function.Start.EnvVars = newKeys("Function.Start.EnvVar", "SF_RUN_ENV_VAR", "Sf.Start.EnvVar")
	Genaiz.Function.Start.Image = newKeys("Function.Start.Image", "SF_RUN_IMAGE", "Sf.Start.Image")
	Genaiz.Function.Start.MountInput = newKeys("Function.Start.Input", "SF_RUN_MOUNT_INPUT", "Sf.Start.Input")
	Genaiz.Function.Start.MountOutput = newKeys("Function.Start.Output", "SF_RUN_MOUNT_OUTPUT", "Sf.Start.Output")
	Genaiz.Function.Start.MountLog = newKeys("Function.Start.Log", "SF_RUN_MOUNT_LOG", "Sf.Start.Log")
	Genaiz.Function.Start.MountVar = newKeys("Function.Start.Var", "SF_RUN_MOUNT_VAR", "Sf.Start.Var")
	Genaiz.Function.Start.Name = newKeys("Function.Start.Name", "SF_RUN_CONTAINER_NAME", "Sf.Start.Name")
	Genaiz.Function.Start.NoPropSync = newKeys("Function.Start.NoPropSync", "SF_START_NO_PROP_SYNC", "Sf.Start.NoPropSync")
	Genaiz.Function.Start.Prefix = newKeys("Function.Start.Prefix", "SF_RUN_CONTAINER_PREFIX", "Sf.Start.Prefix")
	Genaiz.Function.Start.Preserve = newKeys("Function.Start.Preserve", "Sf_RUN_CONTAINER_PRESERVE", "Sf.Start.Preserve")
	Genaiz.Function.Start.Replace = newKeys("Function.Start.Replace", "SF_RUN_REPLACE", "Sf.Start.Replace")

	Genaiz.Function.Stop.Image = newKeys("Function.Stop.Image", "SF_RUN_IMAGE", "Sf.Stop.Image")
	Genaiz.Function.Stop.Name = newKeys("Function.Stop.Name", "SF_RUN_CONTAINER_NAME", "Sf.Stop.Name")
	Genaiz.Function.Stop.Prefix = newKeys("Function.Stop.Prefix", "Sf_RUN_CONTAINER_PREFIX", "Sf.Stop.Prefix")
	Genaiz.Function.Stop.Preserve = newKeys("Function.Stop.Preserve", "Sf_RUN_CONTAINER_PRESERVE", "Sf.Stop.Preserve")

	Genaiz.Function.Test.EnvFile = newKeys("Function.Test.EnvFile", "SF_RUN_ENV_FILE", "Sf.Test.EnvFile")
	Genaiz.Function.Test.EnvVars = newKeys("Function.Test.EnvVar", "SF_RUN_ENV_VAR", "Sf.Test.EnvVar")
	Genaiz.Function.Test.Image = newKeys("Function.Test.Image", "SF_RUN_IMAGE", "Sf.Test.Image")
	Genaiz.Function.Test.MountInput = newKeys("Function.Test.Input", "SF_RUN_MOUNT_INPUT", "Sf.Test.Input")
	Genaiz.Function.Test.MountOutput = newKeys("Function.Test.Output", "SF_RUN_MOUNT_OUTPUT", "Sf.Test.Output")
	Genaiz.Function.Test.MountLog = newKeys("Function.Test.Log", "SF_RUN_MOUNT_LOG", "Sf.Test.Log")
	Genaiz.Function.Test.MountVar = newKeys("Function.Test.Var", "SF_RUN_MOUNT_VAR", "Sf.Test.Var")
	Genaiz.Function.Test.NoPropSync = newKeys("Function.Test.NoPropSync", "SF_TEST_NO_PROP_SYNC", "Sf.Test.NoPropSync")
	Genaiz.Function.Test.Prefix = newKeys("Function.Test.Prefix", "SF_RUN_CONTAINER_PREFIX", "Sf.Test.Prefix")

	Genaiz.Solution.Create.ConfigType = newKeys("Solution.Create.ConfigType", "SN_CREATE_CONFIG_TYPE", "Sn.Create.ConfigType")
	Genaiz.Solution.Create.Description = newKeys("Solution.Create.Description", "SN_CREATE_DESCRIPTION", "Sn.Create.Description")
	Genaiz.Solution.Create.Handle = newKeys("Solution.Create.Handle", "SN_CREATE_HANDLE", "Sn.Create.Handle")
	Genaiz.Solution.Create.Name = newKeys("Solution.Create.Name", "SN_CREATE_NAME", "Sn.Create.Name")
	Genaiz.Solution.Create.Oem = newKeys("Solution.Create.Oem", "SN_CREATE_OEM", "Sn.Create.Oem")
	Genaiz.Solution.Create.Version = newKeys("Solution.Create.Version", "SN_CREATE_VERSION", "Sn.Create.Version")
	Genaiz.Solution.Create.Workflow.Description = newKeys("Solution.Create.Workflow.Description", "SN_CREATE_WORKFLOW_DESCRIPTION", "Sn.Create.Workflow.Description")
	Genaiz.Solution.Create.Workflow.Handle = newKeys("Solution.Create.Workflow.Handle", "SN_CREATE_WORKFLOW_HANDLE", "Sn.Create.Workflow.Handle")
	Genaiz.Solution.Create.Workflow.Name = newKeys("Solution.Create.Workflow.Name", "SN_CREATE_WORKFLOW_NAME", "Sn.Create.Workflow.Name")

	Genaiz.Solution.List.Account = newKeys("Solution.List.Account", "SN_LIST_ACCOUNT", "Sn.List.Account")
	Genaiz.Solution.List.AccountOnly = newKeys("Solution.List.AccountOnly", "SN_LIST_ACCOUNT_ONLY", "Sn.List.AccountOnly")
	Genaiz.Solution.List.Printer = newKeys("Solution.List.Printer", "SN_LIST_PRINTER", "Sn.List.Printer")

	Genaiz.Solution.Log.Format = newKeys("Solution.Log.Format", "SN_LOG_FORMAT", "Sn.Log.Format")
	Genaiz.Solution.Log.Level = newKeys("Solution.Log.Level", "SN_LOG_LEVEL", "Sn.Log.Level")

	Genaiz.Solution.Publish.Account = newKeys("Solution.Publish.Account", "SN_PUBLISH_ACCOUNT", "Sn.Publish.Account")
	Genaiz.Solution.Publish.ConfigType = newKeys("Solution.Publish.ConfigType", "SN_PUBLISH_CONFIG_TYPE", "Sn.Publish.ConfigType")
	Genaiz.Solution.Publish.Description = newKeys("Solution.Publish.Description", "SN_PUBLISH_DESCRIPTION", "Sn.Publish.Description")
	Genaiz.Solution.Publish.Handle = newKeys("Solution.Publish.Handle", "SN_PUBLISH_HANDLE", "Sn.Publish.Handle")
	Genaiz.Solution.Publish.Name = newKeys("Solution.Publish.Name", "SN_PUBLISH_NAME", "Sn.Publish.Name")
	Genaiz.Solution.Publish.Oem = newKeys("Solution.Publish.Oem", "SN_PUBLISH_OEM", "Sn.Publish.Oem")
	Genaiz.Solution.Publish.Printer = newKeys("Solution.Publish.Printer", "SN_PUBLISH_PRINTER", "Sn.Publish.Printer")
	Genaiz.Solution.Publish.Version = newKeys("Solution.Publish.Version", "SN_PUBLISH_VERSION", "Sn.Publish.Version")

	Genaiz.Workflow.Create.ConfigType = newKeys("Workflow.Create.ConfigType", "WF_CREATE_CONFIG_TYPE", "Wf.Create.ConfigType")
	Genaiz.Workflow.Create.Description = newKeys("Workflow.Create.Description", "WF_CREATE_DESCRIPTION", "Wf.Create.Description")
	Genaiz.Workflow.Create.Name = newKeys("Workflow.Create.Name", "WF_CREATE_NAME", "Wf.Create.Name")

	Genaiz.Workflow.Delete.ConfigType = newKeys("Workflow.Delete.ConfigType", "WF_DELETE_CONFIG_TYPE", "Wf.Delete.ConfigType")

	Genaiz.Workflow.Links.Add.ConfigType = newKeys("Workflow.Links.Add.ConfigType", "WF_LINKS_ADD_CONFIG_TYPE", "Wf.Links.Add.ConfigType")
	Genaiz.Workflow.Links.Add.NoValidation = newKeys("Workflow.Links.Add.NoValidation", "WF_LINKS_ADD_NO_VALIDATION", "Wf.Links.Add.NoValidation")
	Genaiz.Workflow.Links.Remove.ConfigType = newKeys("Workflow.Links.Remove.ConfigType", "WF_LINKS_RM_CONFIG_TYPE", "Wf.Links.Remove.ConfigType")
	Genaiz.Workflow.Links.Remove.NoValidation = newKeys("Workflow.Links.Remove.NoValidation", "WF_LINKS_RM_NO_VALIDATION", "Wf.Links.Remove.NoValidation")

	Genaiz.Workflow.List.Account = newKeys("Workflow.List.Account", "WF_LIST_ACCOUNT", "Wf.List.Account")
	Genaiz.Workflow.List.Printer = newKeys("Workflow.List.Printer", "WF_LIST_PRINTER", "Wf.List.Printer")

	Genaiz.Workflow.Nodes.Add.ConfigType = newKeys("Workflow.Nodes.Add.ConfigType", "WF_NODES_ADD_CONFIG_TYPE", "Wf.Nodes.Add.ConfigType")
	Genaiz.Workflow.Nodes.Add.Description = newKeys("Workflow.Nodes.Add.Description", "WF_NODES_ADD_DESCRIPTION", "Wf.Nodes.Add.Description")
	Genaiz.Workflow.Nodes.Add.Deserialized = newKeys("Workflow.Nodes.Add.Deserialized", "WF_NODES_ADD_DESERIALIZED", "Wf.Nodes.Add.Deserialized")
	Genaiz.Workflow.Nodes.Add.Handle = newKeys("Workflow.Nodes.Add.Handle", "WF_NODES_ADD_HANDLE", "Wf.Nodes.Add.Handle")
	Genaiz.Workflow.Nodes.Add.Name = newKeys("Workflow.Nodes.Add.Name", "WF_NODES_ADD_NAME", "Wf.Nodes.Add.Name")
	Genaiz.Workflow.Nodes.Add.Oem = newKeys("Workflow.Nodes.Add.Oem", "WF_NODES_ADD_OEM", "Wf.Nodes.Add.Oem")
	Genaiz.Workflow.Nodes.Add.Sequence = newKeys("Workflow.Nodes.Add.Seq", "WF_NODES_ADD_SEQ", "Wf.Nodes.Add.Seq")
	Genaiz.Workflow.Nodes.Add.Serialized = newKeys("Workflow.Nodes.Add.Serialized", "WF_NODES_ADD_SERIALIZED", "Wf.Nodes.Add.Serialized")
	Genaiz.Workflow.Nodes.Add.Version = newKeys("Workflow.Nodes.Add.Version", "WF_NODES_ADD_VERSION", "Wf.Nodes.Add.Version")
	Genaiz.Workflow.Nodes.Remove.ConfigType = newKeys("Workflow.Nodes.Remove.ConfigType", "WF_NODES_RM_CONFIG_TYPE", "Wf.Nodes.Remove.ConfigType")

	Genaiz.Workflow.Props.Add.NoSync = newKeys("Workflow.Props.Add.NoSync", "WF_PROPS_ADD_NO_SYNC", "Wf.Props.Add.NoSync")
	Genaiz.Workflow.Props.Add.NoValidation = newKeys("Workflow.Props.Add.NoValidation", "WF_PROPS_ADD_NO_VALIDATION", "Wf.Props.Add.NoValidation")
	Genaiz.Workflow.Props.Edit.NoSync = newKeys("Workflow.Props.Edit.NoSync", "WF_PROPS_EDIT_NO_SYNC", "Wf.Props.Edit.NoSync")
	Genaiz.Workflow.Props.Edit.NoValidation = newKeys("Workflow.Props.Edit.NoValidation", "WF_PROPS_EDIT_NO_VALIDATION", "Wf.Props.Edit.NoValidation")
	Genaiz.Workflow.Props.List.NoSync = newKeys("Workflow.Props.List.NoSync", "WF_PROPS_LIST_NO_SYNC", "Wf.Props.List.NoSync")

	Genaiz.Workspace.Create.Account = newKeys("Workspace.Create.Account", "WS_CREATE_ACCOUNT", "Ws.Create.Account")
	Genaiz.Workspace.Create.Description = newKeys("Workspace.Create.Description", "WS_CREATE_DESCRIPTION", "Ws.Create.Description")
	Genaiz.Workspace.Create.Printer = newKeys("Workspace.Create.Printer", "WS_CREATE_PRINTER", "Ws.Create.Printer")
	Genaiz.Workspace.Create.RcEnabled = newKeys("Workspace.Create.RcEnabled", "WS_CREATE_RC_ENABLED", "Ws.Create.RcEnabled")
	Genaiz.Workspace.Create.Visibility = newKeys("Workspace.Create.Visibility", "WS_CREATE_VISIBILITY", "Ws.Create.Visibility")

	Genaiz.Workspace.Flow.Create.Account = newKeys("Workspace.Flow.Create.Account", "WS_FLOW_CREATE_ACCOUNT", "Ws.Flow.Create.Account")
	Genaiz.Workspace.Flow.Create.Description = newKeys("Workspace.Flow.Create.Description", "WS_FLOW_CREATE_DESCRIPTION", "Ws.Flow.Create.Description")
	Genaiz.Workspace.Flow.Create.Name = newKeys("Workspace.Flow.Create.Name", "WS_FLOW_CREATE_NAME", "Ws.Flow.Create.Name")
	Genaiz.Workspace.Flow.Create.Printer = newKeys("Workspace.Flow.Create.Printer", "WS_FLOW_CREATE_PRINTER", "WS.Flow.Create.Printer")

	Genaiz.Workspace.List.Account = newKeys("Workspace.List.Account", "WS_LIST_ACCOUNT", "Ws.List.Account")
	Genaiz.Workspace.List.DateMonthly = newKeys("Workspace.List.DateMonthly", "WS_LIST_DATE_MONTHLY", "Ws.List.DateMonthly")
	Genaiz.Workspace.List.DateToday = newKeys("Workspace.List.DateToday", "WS_LIST_DATE_TODAY", "Ws.List.DateToday")
	Genaiz.Workspace.List.DateWeekly = newKeys("Workspace.List.DateWeekly", "WS_LIST_DATE_WEEKLY", "Ws.List.DateWeekly")
	Genaiz.Workspace.List.OwnerOnly = newKeys("Workspace.List.OwnerOnly", "WS_LIST_OWNER_ONLY", "Ws.List.OwnerOnly")
	Genaiz.Workspace.List.Printer = newKeys("Workspace.List.Printer", "WS_LIST_PRINTER", "Ws.List.Printer")
	Genaiz.Workspace.List.RcEnabled = newKeys("Workspace.List.RcEnabled", "WS_LIST_RC_ENABLED", "Ws.List.RcEnabled")
}

func newKeys(docKey, envKey string, pseudonyms ...string) Keys {
	return Keys{
		Doc:        docKey,
		Env:        envKey,
		Pseudonyms: pseudonyms,
	}
}
