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
			Legacy  Keys
			NoCache Keys
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
		Env struct {
			Context Keys
			File    Keys
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

			Arches      Keys
			Description Keys
			Extras      Keys
			Handle      Keys
			InputPorts  Keys
			Name        Keys
			NoUpdate    Keys
			Oem         Keys
			OutputPorts Keys
			PropSpecs   Keys
			Rebuild     Keys
			Type        Keys
			Version     Keys
		}
		Run struct {
			EnvFile     Keys
			EnvVars     Keys
			Image       Keys
			MountInput  Keys
			MountLog    Keys
			MountOutput Keys
			MountVar    Keys
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
		Log struct {
			Format Keys
			Level  Keys
		}
		Publish struct {
			Broker      Keys
			ConfigType  Keys
			Description Keys
			Handle      Keys
			Name        Keys
			Oem         Keys
			Version     Keys
		}
	}
	Workflow struct {
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
		Links struct {
			Add struct {
				ConfigType Keys
			}
			Remove struct {
				ConfigType Keys
			}
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
	Genaiz.Function.Build.Legacy = newKeys("Sf.Build.Legacy", "SF_BUILD_LEGACY")
	Genaiz.Function.Build.NoCache = newKeys("Sf.Build.NoCache", "SF_BUILD_NOCACHE")
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
	Genaiz.Function.Create.Recipe = newKeys("Sf.Create.Recipe", "SF_CREATE_RECIPE")
	Genaiz.Function.Create.SolutionPath = newKeys("Sf.Create.SolutionPath", "SF_CREATE_SN_PATH")
	Genaiz.Function.Create.Type = newKeys("Sf.Create.Type", "SF_CREATE_TYPE")
	Genaiz.Function.Create.Version = newKeys("Sf.Create.Version", "SF_CREATE_VERSION")
	Genaiz.Function.Env.Context = newKeys("Sf.Env.Context", "SF_ENV_CONTEXT")
	Genaiz.Function.Env.File = newKeys("Sf.Env.File", "SF_ENV_FILE")
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
	Genaiz.Function.Publish.Extras = newKeys("Sf.Publish.Extras", "SF_PUBLISH_EXTRAS")
	Genaiz.Function.Publish.Description = newKeys("Sf.Publish.Description", "SF_PUBLISH_DESCRIPTION")
	Genaiz.Function.Publish.Handle = newKeys("Sf.Publish.Handle", "SF_PUBLISH_HANDLE")
	Genaiz.Function.Publish.InputPorts = newKeys("Sf.Publish.InputPorts", "")
	Genaiz.Function.Publish.Name = newKeys("Sf.Publish.Name", "SF_PUBLISH_NAME")
	Genaiz.Function.Publish.NoUpdate = newKeys("Sf.Publish.NoUpdate", "SF_PUBLISH_NO_UPDATE")
	Genaiz.Function.Publish.Oem = newKeys("Sf.Publish.Oem", "SF_PUBLISH_OEM")
	Genaiz.Function.Publish.OutputPorts = newKeys("Sf.Publish.OutputPorts", "")
	Genaiz.Function.Publish.PropSpecs = newKeys("Sf.Publish.PropSpecs", "")
	Genaiz.Function.Publish.Rebuild = newKeys("Sf.Publish.Rebuild", "SF_PUBLISH_REBUILD")
	Genaiz.Function.Publish.Type = newKeys("Sf.Publish.Type", "SF_PUBLISH_TYPE")
	Genaiz.Function.Publish.Version = newKeys("Sf.Publish.Version", "SF_PUBLISH_VERSION")
	Genaiz.Function.Publish.DataPortAdd.Input.Desc = newKeys("Sf.Publish.DataPortAdd.Input.Desc", "SF_PUBLISH_DATA_PORT_ADD_DESC")
	Genaiz.Function.Publish.DataPortAdd.Input.Name = newKeys("Sf.Publish.DataPortAdd.Input.Name", "SF_PUBLISH_DATA_PORT_ADD_NAME")
	Genaiz.Function.Publish.DataPortAdd.Output.Desc = newKeys("Sf.Publish.DataPortAdd.Output.Desc", "SF_PUBLISH_DATA_PORT_ADD_DESC")
	Genaiz.Function.Publish.DataPortAdd.Output.Name = newKeys("Sf.Publish.DataPortAdd.Output.Name", "SF_PUBLISH_DATA_PORT_ADD_NAME")
	Genaiz.Function.Publish.PropSpecAdd.DefaultValue = newKeys("Sf.Publish.PropSpecAdd.DefaultValue", "SF_PUBLISH_PROP_SPEC_ADD_DEFAULT_VALUE")
	Genaiz.Function.Publish.PropSpecAdd.Description = newKeys("Sf.Publish.PropSpecAdd.Description", "SF_PUBLISH_PROP_SPEC_ADD_DESCRIPTION")
	Genaiz.Function.Publish.PropSpecAdd.EnumValue = newKeys("Sf.Publish.PropSpecAdd.EnumValue", "SF_PUBLISH_PROP_SPEC_ADD_ENUM_VALUE")
	Genaiz.Function.Publish.PropSpecAdd.Name = newKeys("Sf.Publish.PropSpecAdd.Name", "SF_PUBLISH_PROP_SPEC_ADD_NAME")
	Genaiz.Function.Publish.PropSpecAdd.Type = newKeys("Sf.Publish.PropSpecAdd.Type", "SF_PUBLISH_PROP_SPEC_ADD_TYPE")
	Genaiz.Function.Publish.PropSpecEdit.DefaultValue = newKeys("Sf.Publish.PropSpecEdit.DefaultValue", "SF_PUBLISH_PROP_SPEC_EDIT_DEFAULT_VALUE")
	Genaiz.Function.Publish.PropSpecEdit.Description = newKeys("Sf.Publish.PropSpecEdit.Description", "SF_PUBLISH_PROP_SPEC_EDIT_DESCRIPTION")
	Genaiz.Function.Publish.PropSpecEdit.EnumAddValue = newKeys("Sf.Publish.PropSpecEdit.EnumAddValue", "SF_PUBLISH_PROP_SPEC_EDIT_ENUM_ADD_VALUE")
	Genaiz.Function.Publish.PropSpecEdit.EnumRemoveValue = newKeys("Sf.Publish.PropSpecEdit.EnumRemoveValue", "SF_PUBLISH_PROP_SPEC_EDIT_ENUM_RM_VALUE")
	Genaiz.Function.Publish.PropSpecEdit.EnumValue = newKeys("Sf.Publish.PropSpecEdit.EnumValue", "SF_PUBLISH_PROP_SPEC_EDIT_ENUM_VALUE")
	Genaiz.Function.Publish.PropSpecEdit.Name = newKeys("Sf.Publish.PropSpecEdit.Name", "SF_PUBLISH_PROP_SPEC_EDIT_NAME")
	Genaiz.Function.Run.EnvFile = newKeys("Sf.Run.EnvFile", "SF_RUN_ENV_FILE")
	Genaiz.Function.Run.EnvVars = newKeys("Sf.Run.EnvVar", "SF_RUN_ENV_VAR")
	Genaiz.Function.Run.Image = newKeys("Sf.Run.Image", "SF_RUN_IMAGE")
	Genaiz.Function.Run.MountInput = newKeys("Sf.Run.Input", "SF_RUN_MOUNT_INPUT")
	Genaiz.Function.Run.MountOutput = newKeys("Sf.Run.Output", "SF_RUN_MOUNT_OUTPUT")
	Genaiz.Function.Run.MountLog = newKeys("Sf.Run.Log", "SF_RUN_MOUNT_LOG")
	Genaiz.Function.Run.MountVar = newKeys("Sf.Run.Var", "SF_RUN_MOUNT_VAR")
	Genaiz.Function.Run.Prefix = newKeys("Sf.Run.Prefix", "SF_RUN_CONTAINER_PREFIX")
	Genaiz.Function.Start.EnvFile = newKeys("Sf.Start.EnvFile", "SF_RUN_ENV_FILE")
	Genaiz.Function.Start.EnvVars = newKeys("Sf.Start.EnvVar", "SF_RUN_ENV_VAR")
	Genaiz.Function.Start.Image = newKeys("Sf.Start.Image", "SF_RUN_IMAGE")
	Genaiz.Function.Start.MountInput = newKeys("Sf.Start.Input", "SF_RUN_MOUNT_INPUT")
	Genaiz.Function.Start.MountOutput = newKeys("Sf.Start.Output", "SF_RUN_MOUNT_OUTPUT")
	Genaiz.Function.Start.MountLog = newKeys("Sf.Start.Log", "SF_RUN_MOUNT_LOG")
	Genaiz.Function.Start.MountVar = newKeys("Sf.Start.Var", "SF_RUN_MOUNT_VAR")
	Genaiz.Function.Start.Name = newKeys("Sf.Start.Name", "SF_RUN_CONTAINER_NAME")
	Genaiz.Function.Start.Prefix = newKeys("Sf.Start.Prefix", "SF_RUN_CONTAINER_PREFIX")
	Genaiz.Function.Start.Preserve = newKeys("Sf.Start.Preserve", "Sf_RUN_CONTAINER_PRESERVE")
	Genaiz.Function.Start.Replace = newKeys("Sf.Start.Replace", "SF_RUN_REPLACE")
	Genaiz.Function.Stop.Image = newKeys("Sf.Stop.Image", "SF_RUN_IMAGE")
	Genaiz.Function.Stop.Name = newKeys("Sf.Stop.Name", "SF_RUN_CONTAINER_NAME")
	Genaiz.Function.Stop.Prefix = newKeys("Sf.Stop.Prefix", "Sf_RUN_CONTAINER_PREFIX")
	Genaiz.Function.Stop.Preserve = newKeys("Sf.Stop.Preserve", "Sf_RUN_CONTAINER_PRESERVE")
	Genaiz.Function.Test.EnvFile = newKeys("Sf.Test.EnvFile", "SF_RUN_ENV_FILE")
	Genaiz.Function.Test.EnvVars = newKeys("Sf.Test.EnvVar", "SF_RUN_ENV_VAR")
	Genaiz.Function.Test.Image = newKeys("Sf.Test.Image", "SF_RUN_IMAGE")
	Genaiz.Function.Test.MountInput = newKeys("Sf.Test.Input", "SF_RUN_MOUNT_INPUT")
	Genaiz.Function.Test.MountOutput = newKeys("Sf.Test.Output", "SF_RUN_MOUNT_OUTPUT")
	Genaiz.Function.Test.MountLog = newKeys("Sf.Test.Log", "SF_RUN_MOUNT_LOG")
	Genaiz.Function.Test.MountVar = newKeys("Sf.Test.Var", "SF_RUN_MOUNT_VAR")
	Genaiz.Function.Test.Prefix = newKeys("Sf.Test.Prefix", "SF_RUN_CONTAINER_PREFIX")
	Genaiz.Solution.Create.ConfigType = newKeys("Solution.Create.ConfigType", "SN_CREATE_CONFIG_TYPE")
	Genaiz.Solution.Create.Description = newKeys("Solution.Create.Description", "SN_CREATE_DESCRIPTION")
	Genaiz.Solution.Create.Handle = newKeys("Solution.Create.Handle", "SN_CREATE_HANDLE")
	Genaiz.Solution.Create.Name = newKeys("Solution.Create.Name", "SN_CREATE_NAME")
	Genaiz.Solution.Create.Oem = newKeys("Solution.Create.Oem", "SN_CREATE_OEM")
	Genaiz.Solution.Create.Version = newKeys("Solution.Create.Version", "SN_CREATE_VERSION")
	Genaiz.Solution.Create.Workflow.Description = newKeys("Solution.Creation.Workflow.Description", "SN_CREATE_WORKFLOW_DESCRIPTION")
	Genaiz.Solution.Create.Workflow.Handle = newKeys("Solution.Create.Workflow.Handle", "SN_CREATE_WORKFLOW_HANDLE")
	Genaiz.Solution.Create.Workflow.Name = newKeys("Solution.Create.Workflow.Name", "SN_CREATE_WORKFLOW_NAME")
	Genaiz.Solution.Log.Format = newKeys("Solution.Log.Format", "SN_LOG_FORMAT")
	Genaiz.Solution.Log.Level = newKeys("Solution.Log.Level", "SN_LOG_LEVEL")
	Genaiz.Solution.Publish.Broker = newKeys("Solution.Publish.Broker", "SN_PUBLISH_BROKER")
	Genaiz.Solution.Publish.ConfigType = newKeys("Solution.Publish.ConfigType", "SN_PUBLISH_CONFIG_TYPE")
	Genaiz.Solution.Publish.Description = newKeys("Solution.Publish.Description", "SN_PUBLISH_DESCRIPTION")
	Genaiz.Solution.Publish.Handle = newKeys("Solution.Publish.Handle", "SN_PUBLISH_HANDLE")
	Genaiz.Solution.Publish.Name = newKeys("Solution.Publish.Name", "SN_PUBLISH_NAME")
	Genaiz.Solution.Publish.Oem = newKeys("Solution.Publish.Oem", "SN_PUBLISH_OEM")
	Genaiz.Solution.Publish.Version = newKeys("Solution.Publish.Version", "SN_PUBLISH_VERSION")
	Genaiz.Workflow.Links.Add.ConfigType = newKeys("Workflow.Links.Add.ConfigType", "WF_LINKS_ADD_CONFIG_TYPE")
	Genaiz.Workflow.Links.Remove.ConfigType = newKeys("Workflow.Links.Remove.ConfigType", "WF_LINKS_RM_CONFIG_TYPE")
	Genaiz.Workflow.Nodes.Add.ConfigType = newKeys("Workflow.Nodes.Add.ConfigType", "WF_NODES_ADD_CONFIG_TYPE")
	Genaiz.Workflow.Nodes.Add.Description = newKeys("Workflow.Nodes.Add.Description", "WF_NODES_ADD_DESCRIPTION")
	Genaiz.Workflow.Nodes.Add.Deserialized = newKeys("Workflow.Nodes.Add.Deserialized", "WF_NODES_ADD_DESERIALIZED")
	Genaiz.Workflow.Nodes.Add.Handle = newKeys("Workflow.Nodes.Add.Handle", "WF_NODES_ADD_HANDLE")
	Genaiz.Workflow.Nodes.Add.Name = newKeys("Workflow.Nodes.Add.Name", "WF_NODES_ADD_NAME")
	Genaiz.Workflow.Nodes.Add.Oem = newKeys("Workflow.Nodes.Add.Oem", "WF_NODES_ADD_OEM")
	Genaiz.Workflow.Nodes.Add.Sequence = newKeys("Workflow.Nodes.Add.Seq", "WF_NODES_ADD_SEQ")
	Genaiz.Workflow.Nodes.Add.Serialized = newKeys("Workflow.Nodes.Add.Serialized", "WF_NODES_ADD_SERIALIZED")
	Genaiz.Workflow.Nodes.Add.Version = newKeys("Workflow.Nodes.Add.Version", "WF_NODES_ADD_VERSION")
	Genaiz.Workflow.Nodes.Remove.ConfigType = newKeys("Workflow.Nodes.Remove.ConfigType", "WF_NODES_RM_CONFIG_TYPE")
}

func newKeys(docKey, envKey string) Keys {
	return Keys{
		Doc: docKey,
		Env: envKey,
	}
}
