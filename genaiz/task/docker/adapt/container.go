package adapt

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/stringz"
)

const (
	EnvFunctionType = "SF_TYPE"
	EnvInputPath    = "SF_INPUT_PATH"
	EnvLogPath      = "SF_LOG_PATH"
	EnvOutputPath   = "SF_OUTPUT_PATH"
	EnvProgressFile = "SF_PROGRESS_FILE"
	EnvResultFile   = "SF_RESULT_FILE"
	EnvStatusFile   = "SF_STATUS_FILE"
	EnvVarPath      = "SF_VAR_PATH"
)

var (
	EnvKeys = []string{EnvFunctionType, EnvInputPath, EnvLogPath, EnvOutputPath,
		EnvProgressFile, EnvResultFile, EnvStatusFile, EnvVarPath}
	EnvFileMap = map[string]string{
		EnvProgressFile: "$SF_VAR_PATH/progress",
		EnvResultFile:   "$SF_VAR_PATH/result",
		EnvStatusFile:   "$SF_VAR_PATH/status",
	}

	Mounts struct {
		Input  ContainerMountPoint
		Log    ContainerMountPoint
		Output ContainerMountPoint
		Var    ContainerMountPoint
	}
)

type ContainerArguments struct {
	DockerImage string
	EnvVars     map[string]string
	Force       bool
	MountInput  string
	MountOutput string
	MountLog    string
	MountVar    string
	Name        string
	Prefix      string
}

// GetEnvironment returns the environment specification to be copied to a container upon creation or start
func (c ContainerArguments) GetEnvironment() map[string]string {
	if c.EnvVars == nil {
		c.EnvVars = make(map[string]string)
	}

	return c.EnvVars
}

// GetMountBinds returns the mount points and their associated bindings from the local filesystem
func (c ContainerArguments) GetMountBinds() []ContainerMountBind {
	var result []ContainerMountBind

	if c.MountInput != "" {
		result = append(result, Mounts.Input.MakeBind(c.MountInput))
	}

	if c.MountLog != "" {
		result = append(result, Mounts.Log.MakeBind(c.MountLog))
	}

	if c.MountOutput != "" {
		result = append(result, Mounts.Output.MakeBind(c.MountOutput))
	}

	if c.MountVar != "" {
		result = append(result, Mounts.Var.MakeBind(c.MountVar))
	}

	return result
}

// GetName attempts returning a name for the container targeted or created, if no Name or Prefix were set in the arguments, the method returns an empty string
func (c ContainerArguments) GetName(summaries []container.Summary) string {
	if c.Name != "" {
		return c.Name
	} else if c.Prefix != "" {
		return fmt.Sprintf("%s-%d", c.Prefix, len(summaries))
	}

	return ""
}

// MakeDisposableName generates a name from Name or Prefix, or from a default `container` with a milliseconds timestamp
func (c ContainerArguments) MakeDisposableName() string {
	var result = ""

	if c.Name != "" {
		result = c.Name
	} else if c.Prefix != "" {
		if matched, _ := regexp.MatchString("-\\d+$", c.Prefix); matched {
			var i = strings.LastIndex(c.Prefix, "-")

			result = c.Prefix[:i]
		} else {
			result = c.Prefix
		}
	} else {
		result = "container"
	}

	return result + fmt.Sprintf("-d%d", time.Now().UnixMilli())
}

// MountArguments returns a string of formatted environment and mount arguments corresponding to MountInput, MountOutput, MountLog and MountVar. This is useful when using a fork process or for explaining the equivalent command executed by a task
func (c ContainerArguments) MountArguments() string {
	var stringParams = []string{
		c.fmtMountBindArg(&Mounts.Input, c.MountInput),
		c.fmtMountBindArg(&Mounts.Output, c.MountOutput),
		c.fmtMountBindArg(&Mounts.Log, c.MountLog),
		c.fmtMountBindArg(&Mounts.Var, c.MountVar),
		c.fmtMountEnvArg(&Mounts.Input, c.MountInput),
		c.fmtMountEnvArg(&Mounts.Output, c.MountOutput),
		c.fmtMountEnvArg(&Mounts.Log, c.MountLog),
		c.fmtMountEnvArg(&Mounts.Var, c.MountVar),
	}

	return strings.Join(stringz.AllNonEmpty(stringParams...), "   ")
}

func (c ContainerArguments) fmtMountBindArg(mount *ContainerMountPoint, path string) string {
	if path != "" {
		var local = filez.FromWorkDir(path)
		var readOnly = ""

		if mount.ReadOnly {
			readOnly = ",ro"
		}

		return fmt.Sprintf("--mount type=bind,src=%s,dst=%s%s \\\n", local, mount.MountPath, readOnly)
	}

	return ""
}

func (c ContainerArguments) fmtMountEnvArg(mount *ContainerMountPoint, path string) string {
	if path != "" {
		return fmt.Sprintf("--env %s=%s \\\n", mount.EnvVar, mount.MountPath)
	}

	return ""
}

type ContainerMountBind struct {
	ContainerMountPoint
	HostPath string
}

type ContainerMountPoint struct {
	EnvVar    string
	MountPath string
	ReadOnly  bool
}

// GetEnvKeyValue returns a pair of environment variable KEY and its corresponding default value
func (p ContainerMountPoint) GetEnvKeyValue() (string, string) {
	return p.EnvVar, p.MountPath
}

// MakeBind wraps a ContainerMountPoint into a ContainerMountBind using the specified hostPath as bind path
func (p ContainerMountPoint) MakeBind(hostPath string) ContainerMountBind {
	return ContainerMountBind{
		ContainerMountPoint: p,
		HostPath:            hostPath,
	}
}

// MakeMount creates a moby mount.Mount object with the specified hostDir as the binding source
func (p ContainerMountPoint) MakeMount(hostDir string) mount.Mount {
	return mount.Mount{
		Type:     mount.TypeBind,
		Source:   hostDir,
		Target:   p.MountPath,
		ReadOnly: p.ReadOnly,
		BindOptions: &mount.BindOptions{
			CreateMountpoint: true,
		},
	}
}

func init() {
	Mounts.Input = ContainerMountPoint{
		EnvVar:    EnvInputPath,
		MountPath: "/mnt/in",
		ReadOnly:  true,
	}

	Mounts.Log = ContainerMountPoint{
		EnvVar:    EnvLogPath,
		MountPath: "/mnt/log",
		ReadOnly:  false,
	}

	Mounts.Output = ContainerMountPoint{
		EnvVar:    EnvOutputPath,
		MountPath: "/mnt/out",
		ReadOnly:  false,
	}

	Mounts.Var = ContainerMountPoint{
		EnvVar:    EnvVarPath,
		MountPath: "/mnt/var",
		ReadOnly:  false,
	}
}
