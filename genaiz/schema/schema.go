// Package schema provides schematic elements for validating Genaiz.yaml files. It also provides a central structure for all configuration key constants used by the genaiz toolkits.
package schema

var (
	Genaiz = &Document{}
)

type Document struct {
	Account struct {
		Login struct {
			Password Keys
			Refresh  Keys
			Username Keys
		}
		Logout struct {
			Host     Keys
			Username Keys
		}
	}
	Function struct {
		Build struct {
			Context Keys
			File    Keys
			Label   Keys
			Prune   Keys
			Tag     Keys
			Version Keys
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
			SolutionPath Keys
			Type         Keys
			Version      Keys
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
			Arches      Keys
			Description Keys
			Handle      Keys
			Name        Keys
			NoUpdate    Keys
			Oem         Keys
			Rebuild     Keys
			Type        Keys
			Version     Keys
		}
	}
}

type Keys struct {
	Doc string
	Env string
}

func init() {
	Genaiz.Account.Login.Password = newKeys("p", "GENAIZ_PASSWORD")
	Genaiz.Account.Login.Refresh = newKeys("Account.Login.Refresh", "AC_LOGIN_REFRESH")
	Genaiz.Account.Login.Username = newKeys("Account.Login.Username", "GENAIZ_USERNAME")
	Genaiz.Account.Logout.Host = newKeys("Account.Logout.Host", "AC_LOGOUT_HOST")
	Genaiz.Account.Logout.Username = newKeys("Account.Logout.Username", "GENAIZ_USERNAME")
	Genaiz.Function.Build.Context = newKeys("Sf.Build.Context", "SF_BUILD_CONTEXT")
	Genaiz.Function.Build.File = newKeys("Sf.Build.File", "SF_BUILD_FILE")
	Genaiz.Function.Build.Label = newKeys("Sf.Build.Label", "SF_BUILD_LABEL")
	Genaiz.Function.Build.Prune = newKeys("Sf.Build.Prune", "SF_BUILD_PRUNE")
	Genaiz.Function.Build.Tag = newKeys("Sf.Build.Tag", "SF_BUILD_TAG")
	Genaiz.Function.Build.Version = newKeys("Sf.Build.Version", "SF_BUILD_VERSION")
	Genaiz.Function.Create.Arches = newKeys("Sf.Create.Arches", "SF_CREATE_ARCHES")
	Genaiz.Function.Create.ConfigType = newKeys("Sf.Create.ConfigType", "SF_CREATE_CONFIG_TYPE")
	Genaiz.Function.Create.Handle = newKeys("Sf.Create.Handle", "SF_CREATE_HANDLE")
	Genaiz.Function.Create.MountInput = newKeys("Sf.Create.Input", "SF_CREATE_MOUNT_INPUT")
	Genaiz.Function.Create.MountOutput = newKeys("Sf.Create.Output", "SF_CREATE_MOUNT_OUTPUT")
	Genaiz.Function.Create.Name = newKeys("Sf.Create.Name", "SF_CREATE_NAME")
	Genaiz.Function.Create.Oem = newKeys("Sf.Create.Oem", "SF_CREATE_OEM")
	Genaiz.Function.Create.Recipe = newKeys("SF.Create.Recipe", "SF_CREATE_RECIPE")
	Genaiz.Function.Create.SolutionPath = newKeys("SF.Create.SolutionPath", "SF_CREATE_SN_PATH")
	Genaiz.Function.Create.Type = newKeys("Sf.Create.Type", "SF_CREATE_TYPE")
	Genaiz.Function.Create.Version = newKeys("SF.Create.Version", "SF_CREATE_VERSION")
	Genaiz.Function.Init.Arches = newKeys("Sf.Init.Arches", "SF_INIT_ARCHES")
	Genaiz.Function.Init.ConfigType = newKeys("Sf.Init.ConfigType", "SF_INIT_CONFIG_TYPE")
	Genaiz.Function.Init.Handle = newKeys("Sf.Init.Handle", "SF_INIT_HANDLE")
	Genaiz.Function.Init.MountInput = newKeys("Sf.Init.Input", "SF_INIT_MOUNT_INPUT")
	Genaiz.Function.Init.MountOutput = newKeys("Sf.Init.Output", "SF_INIT_MOUNT_OUTPUT")
	Genaiz.Function.Init.Name = newKeys("Sf.Init.Name", "SF_INIT_NAME")
	Genaiz.Function.Init.Oem = newKeys("Sf.Init.Oem", "SF_INIT_OEM")
	Genaiz.Function.Init.SolutionPath = newKeys("SF.Init.SolutionPath", "SF_INIT_SN_PATH")
	Genaiz.Function.Init.Type = newKeys("Sf.Init.Type", "SF_INIT_TYPE")
	Genaiz.Function.Init.Version = newKeys("Sf.Init.Version", "SF_INIT_VERSION")
	Genaiz.Function.Publish.Arches = newKeys("Sf.Publish.Arches", "SF_PUBLISH_ARCHES")
	Genaiz.Function.Publish.Handle = newKeys("Sf.Publish.Handle", "SF_PUBLISH_HANDLE")
	Genaiz.Function.Publish.Name = newKeys("Sf.Publish.Name", "SF_PUBLISH_NAME")
	Genaiz.Function.Publish.NoUpdate = newKeys("Sf.Publish.NoUpdate", "SF_PUBLISH_NO_UPDATE")
	Genaiz.Function.Publish.Oem = newKeys("Sf.Publish.Oem", "SF_PUBLISH_OEM")
	Genaiz.Function.Publish.Rebuild = newKeys("Sf.Publish.Rebuild", "SF_PUBLISH_REBUILD")
	Genaiz.Function.Publish.Type = newKeys("Sf.Publish.Type", "SF_PUBLISH_TYPE")
	Genaiz.Function.Publish.Version = newKeys("Sf.Publish.Version", "SF_PUBLISH_VERSION")
}

func newKeys(docKey, envKey string) Keys {
	return Keys{
		Doc: docKey,
		Env: envKey,
	}
}
