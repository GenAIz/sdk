package adapt

import (
	"strings"

	"github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"genaiz.com/genaiz-lib/lang/stringz"
)

// BuildArguments represents the set of arguments the build docker task support when using the Deprecated legacy builder on dockerd or an available `buildx` command from the docker CLI.
type BuildArguments struct {
	Dockerfile       string
	DockerContext    string
	DockerRepository string
	DockerVersion    string
	Label            bool
	NoCache          bool
	Platform         string
}

func (a BuildArguments) CliArguments() []string {
	var args = []string{"build", "--pull", "-t", a.GetReference()}

	if a.Dockerfile != "" {
		args = append(args, "-f", a.Dockerfile)
	}

	if a.Platform != "" {
		args = append(args, "--platform", a.Platform)
	}

	if a.Label {
		args = append(args, "--label")
	}

	if a.NoCache {
		args = append(args, "--no-cache")
	}

	// Needs to be last
	if a.DockerContext == "" {
		args = append(args, ".")
	} else {
		args = append(args, a.DockerContext)
	}

	return args
}

// GetFilters adapts DockerRepository and the optional DockerVersion to a reference string wrapping it as a client.Filters.
func (a BuildArguments) GetFilters() client.Filters {
	return make(client.Filters).
		Add("reference", a.GetReference())
}

// GetFiltersByRepo adapts DockerRepository to the specified field, excluding the first component specified in the repository string if there's more than one component. It returns a client.Filters.
//
// Example: "repository/namespace1/namespace2" would return "namespace1/namespace2"
func (a BuildArguments) GetFiltersByRepo(field string) client.Filters {
	var value string

	if i := strings.Index(a.DockerRepository, "/"); i >= 0 {
		value = a.DockerRepository[i+1:]
	} else {
		value = a.DockerRepository
	}

	return make(client.Filters).Add(field, value+"*")
}

// GetFiltersByVersion adapts DockerRepository and the optional DockerVersion to a glob string if the DockerVersion is explicitly set to "latest", otherwise the version string will be part of the filter value, returned as a client.Filters
func (a BuildArguments) GetFiltersByVersion() client.Filters {
	if a.DockerVersion == "latest" {
		return make(client.Filters).
			Add("reference", a.DockerRepository+":*")
	}

	return a.GetFilters()
}

// GetPlatforms adapts Platform to a slice of ocispec.Platform objects. It only supports OS and Architecture for now
func (a BuildArguments) GetPlatforms() []ocispec.Platform {
	if a.Platform != "" {
		var ociPlatform = &ocispec.Platform{}

		if parts := strings.Split(a.Platform, "/"); len(parts) == 2 {
			ociPlatform.OS = parts[0]
			ociPlatform.Architecture = parts[1]
		} else {
			ociPlatform.OS = a.Platform
		}

		return []ocispec.Platform{*ociPlatform}
	}

	return []ocispec.Platform{}
}

// GetReference construct a reference Docker image string of the form "repository:version"
func (a BuildArguments) GetReference() string {
	return stringz.SingleTagLabel(a.DockerRepository, ":", a.GetVersion())
}

// GetVersion will return the DockerVersion is it is set, otherwise it returns "latest"
func (a BuildArguments) GetVersion() string {
	return stringz.FirstNonEmpty(a.DockerVersion, "latest")
}

// PruneArguments is used to generate corresponding buildx prune CLI arguments to a build command. Because we use a BuildX forked process to launch builds, avoiding the deprecated build on dockerd, pruning images and layers implies a call to `buildx prune`.
func (a BuildArguments) PruneArguments() []string {
	var result = []string{"buildx", "prune", "--force"}

	if !a.NoCache {
		// We'll keep half a day's worth of cached build artifacts to help the build process
		result = append(result, "--filter", "until=12h")
	}

	return result
}
