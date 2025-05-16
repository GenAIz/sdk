package docker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/moby/go-archive"

	"genaiz.com/genaiz/lang/stringz"
	"genaiz.com/genaiz/task"
)

type HasReference interface {
	GetNamePrefix() string

	GetReference() string

	GetVersion() string
}

type BuildParams struct {
	task.Env
	Dockerfile    string
	DockerContext string
	DockerTag     string
	DockerVersion string
	Prune         bool
}

func (p *BuildParams) GetNamePrefix() string {
	var tag string

	if p.DockerVersion != "" {
		tag = stringz.MultiTagLabel(p.DockerTag, "-", p.DockerVersion)
	}

	// We need to keep any valid prefix in the tag
	return strings.ReplaceAll(tag, "/", "-")
}

func (p *BuildParams) GetReference() string {
	return stringz.SingleTagLabel(p.DockerTag, ":", p.GetVersion())
}

func (p *BuildParams) GetVersion() string {
	return stringz.FirstNonEmpty(p.DockerVersion, "latest")
}

func NewBuildTask() *task.Task[BuildParams] {
	return &task.Task[BuildParams]{
		Name:         "docker-build",
		OnPrepare:    handleBuildContext,
		OnIncomplete: handleBuildCreate,
		OnComplete:   handleBuildPrune,
		OnPretend:    handleBuildPretend,
	}
}

func handleBuildContext(params *BuildParams, state *task.State) error {
	var version = params.GetVersion()
	var reference = params.GetReference()
	var listFilters = filters.NewArgs(
		filters.Arg("reference", reference),
	)

	state.Logger.Debugf("Finding a docker image for reference [%s]", reference)

	if summaries, err := dockerClient.ImageList(params.Context, image.ListOptions{Filters: listFilters}); err == nil {
		if len(summaries) == 0 {
			state.Logger.Debugf("Could not find an image for reference [%s]", reference)
			return errors.New("not found")
		} else if version == "latest" {
			state.Logger.Debugf("Version latest requires a fresh build")
			return errors.New("latest needs to be rebuilt")
		} else {
			state.Logger.Warningf("Image with reference [%s] exists, skipping", reference)
			state.Output = "Found image " + reference
			return nil
		}
	} else {
		return err
	}
}

func handleBuildCreate(params *BuildParams, state *task.State) error {
	var reference = params.GetReference()
	var buildCtx, _ = archive.TarWithOptions(params.DockerContext, &archive.TarOptions{})
	var options = types.ImageBuildOptions{
		Dockerfile: params.Dockerfile,
		Tags:       []string{reference},
		Labels:     map[string]string{"sf": params.DockerTag},
	}

	state.Logger.Debugf("Building a docker image tagged [%s]", reference)

	if resp, err := dockerClient.ImageBuild(params.Context, buildCtx, options); err == nil {
		var scanner = bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			var output Output

			if err = json.Unmarshal(scanner.Bytes(), &output); err == nil {
				if output.Stream != "" {
					state.Logger.Debugf("%s", strings.TrimSuffix(output.Stream, "\n"))
				}
			} else {
				state.Logger.Warningf("Could not parse json with error: %s", err)
				state.Logger.Debugf("String: %s", scanner.Text())
			}
		}

		return nil
	} else {
		return err
	}
}

func handleBuildPretend(params *BuildParams, state *task.State) error {
	var version = stringz.FirstNonEmpty(params.DockerVersion, "latest")
	var reference = stringz.SingleTagLabel(params.DockerTag, ":", version)

	state.Logger.Debugf("Pretending to build a docker image tagged [%s]", reference)

	if params.Dockerfile == "" {
		fmt.Printf("docker build -t %s %s\n", reference, params.DockerContext)
	} else {
		fmt.Printf("docker build -f %s -t %s %s\n", params.Dockerfile, reference, params.DockerContext)
	}

	if params.Prune {
		fmt.Printf("docker image prune --filter label=%s=%s\n", "sf", params.DockerTag)
	}

	state.Completed = true
	return nil
}

func handleBuildPrune(params *BuildParams, state *task.State) error {
	if params.Prune {
		var pruneFilters = filters.NewArgs(
			filters.Arg("label", "sf="+params.DockerTag),
		)

		if report, err := dockerClient.ImagesPrune(params.Context, pruneFilters); err == nil {
			for _, deleted := range report.ImagesDeleted {
				state.Logger.Debugf("Removed dangling image id [%s], no longer in use", deleted.Deleted)
			}
		} else {
			state.Logger.Warningf("Could not prune on [%s] with error: %s", params.DockerTag, err)
		}
	} else {
		state.Logger.Debugf("Pruning disabled, skipping")
	}

	return nil
}
