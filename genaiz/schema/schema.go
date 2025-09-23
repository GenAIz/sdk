// Package schema provides schematic elements for validating Genaiz.yaml files. It also provides a central structure for all configuration key constants used by the genaiz toolkits.
package schema

var (
	Genaiz = &Document{}
)

type Document struct {
	Function struct {
		Build struct {
			Tag     string
			Version string
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
	Genaiz.Function.Build.Tag = "Sf.Build.Tag"
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
