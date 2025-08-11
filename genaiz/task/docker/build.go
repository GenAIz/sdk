package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/moby/go-archive"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/mapz"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

const (
	hashPrefix = "sha256:"
)

var (
	ErrorLatestBuild = errors.New("latest needs to be rebuilt")
	ErrorNoBuild     = errors.New("could not find a build image")
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

func (p *BuildParams) GetFilters() filters.Args {
	return filters.NewArgs(
		filters.Arg("reference", p.GetReference()),
	)
}

func (p *BuildParams) GetFiltersByVersion() filters.Args {
	if p.DockerVersion != "latest" {
		return p.GetFilters()
	}

	return filters.NewArgs(
		filters.Arg("reference", p.DockerTag+":*"),
	)
}

func (p *BuildParams) GetFiltersByTag(field string) filters.Args {
	var value string

	if i := strings.Index(p.DockerTag, "/"); i >= 0 {
		value = p.DockerTag[i+1:]
	} else {
		value = p.DockerTag
	}

	return filters.NewArgs(filters.Arg(field, value+"*"))
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

func NewInspectTask() *task.Task[BuildParams] {
	return &task.Task[BuildParams]{
		Name:       "docker-inspect",
		OnPrepare:  handleInspectContext,
		OnComplete: handleInspectComplete,
		OnPretend:  handleInspectPretend,
	}
}

func NewListTask() *task.Task[BuildParams] {
	return &task.Task[BuildParams]{
		Name:       "docker-list",
		OnPrepare:  handleListContext,
		OnComplete: handleListComplete,
	}
}

func formatCreated(created int64, threshold time.Time) string {
	var createdTime = time.UnixMilli(created * 1000)

	if createdTime.After(threshold) {
		return createdTime.Format(time.TimeOnly)
	} else {
		return createdTime.Format(time.DateOnly)
	}
}

func handleBuildContext(params *BuildParams, state *task.State) error {
	var version = params.GetVersion()
	var reference = params.GetReference()
	var listFilters = params.GetFilters()
	var summaries []image.Summary
	var err error

	state.Logger.Debugf("Finding a docker image for reference [%s]", reference)

	if summaries, err = dockerClient.ImageList(params.Context, image.ListOptions{Filters: listFilters}); err == nil {
		if len(summaries) == 0 {
			state.Logger.Debugf("Could not find an image for reference [%s]", reference)
			err = ErrorNoBuild
		} else if version == "latest" {
			state.Logger.Debugf("Version latest requires a fresh build")
			state.Output = summaries[0].ID
			err = ErrorLatestBuild
		} else {
			state.Logger.Warningf("Found image with reference [%s]", reference)
			state.Output = summaries[0].ID
			return nil
		}
	}

	return err
}

func handleBuildCreate(params *BuildParams, state *task.State) error {
	var reference = params.GetReference()
	var buildCtx, _ = archive.TarWithOptions(params.DockerContext, &archive.TarOptions{})
	var options = build.ImageBuildOptions{
		Dockerfile: params.Dockerfile,
		Tags:       []string{reference},
		Labels:     map[string]string{"sf": params.DockerTag},
	}

	state.Logger.Debugf("Building a docker image tagged [%s]", reference)

	if resp, err := dockerClient.ImageBuild(params.Context, buildCtx, options); err == nil {
		var scanner = bufio.NewScanner(resp.Body)
		defer filez.CloseSilently(resp.Body)

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
	fmt.Printf("docker image list --filter reference=%s\n", reference)

	if state.Error != nil {
		state.Logger.Debugf("Incomplete build context needs to be built [%s]", state.Error)

		if params.Dockerfile == "" {
			fmt.Printf("docker build -t %s %s\n", reference, params.DockerContext)
		} else {
			fmt.Printf("docker build -f %s -t %s %s\n", params.Dockerfile, reference, params.DockerContext)
		}

		if params.Prune {
			fmt.Printf("docker image prune --filter label=%s=%s\n", "sf", params.DockerTag)
		}
	}

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

func handleListContext(params *BuildParams, state *task.State) error {
	var imageFilters = params.GetFiltersByVersion()
	var summaries []image.Summary
	var err error

	if summaries, err = dockerClient.ImageList(params.Context, image.ListOptions{Filters: imageFilters}); err == nil {
		if len(summaries) == 0 {
			state.Logger.Debugf("Could not list images for the provided filter")
			return ErrorNoBuild
		} else {
			state.Internal = summaries
		}

		if params.DockerTag != "" {
			var containers []container.Summary
			var containerOptions = container.ListOptions{
				All:     true,
				Filters: params.GetFiltersByTag("name"),
			}

			if containers, err = dockerClient.ContainerList(params.Context, containerOptions); err == nil {
				state.Containers = &containers
				return nil
			}
		}
	}

	return err
}

func handleListComplete(params *BuildParams, state *task.State) error {
	var year, month, day = time.Now().In(time.Local).Date()
	var today = time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	var output bytes.Buffer

	if state.Internal != nil {
		var imgList = state.Internal.([]image.Summary)

		state.Logger.Debugf("Listing Docker Images for Smart Function [%s], Version [%s]", params.DockerTag, params.DockerVersion)
		listImages(imgList, &output, today)
		state.Output = output.String()
		output.Reset()

		if state.HasContainers() {
			var imagesById = mapz.Mapped(imgList, func(summary image.Summary) string {
				return summary.ID
			})

			state.Logger.Debugf("Listing Docker Containers for Smart Function [%s], Version [%s]", params.DockerTag, params.DockerVersion)
			listContainers(*state.Containers, imagesById, &output, today)
			state.Output += "\n" + output.String()
		} else {
			state.Output += "\nNo containers associated with the listed images\n"
		}
	} else {
		state.Output = fmt.Sprintf("No images associated with Smart Function [%s]\n", params.DockerTag)
	}

	state.Internal = nil
	return nil
}

func handleInspectComplete(params *BuildParams, state *task.State) error {
	if state.Output != "" {
		var err error
		var resp image.InspectResponse

		state.Logger.Debugf("Inspecting docker image [%s]", state.Output)

		if resp, err = dockerClient.ImageInspect(params.Context, state.Output); err == nil {
			var digest string

			if len(resp.RepoDigests) > 0 {
				if i := strings.LastIndex(resp.RepoDigests[0], hashPrefix); i >= 0 {
					digest = resp.RepoDigests[0][i+len(hashPrefix):]
				}
			}

			state.Internal = &shared.Identity{
				Id:      resp.ID,
				Hash:    digest,
				Version: params.GetVersion(),
			}
		}

		return err
	}

	return ErrorNoBuild
}

func handleInspectContext(params *BuildParams, state *task.State) error {
	var err error

	if state.Output == "" {
		if err = handleBuildContext(params, state); err != nil {
			if errors.Is(err, ErrorLatestBuild) {
				return nil
			}
		}

		if err == nil && state.Output == "" {
			return ErrorNoBuild
		}
	}

	return err
}

func handleInspectPretend(params *BuildParams, state *task.State) error {
	var reference = params.GetReference()

	if state.Error == nil || errors.Is(state.Error, ErrorLatestBuild) {
		state.Logger.Debugf("Pretending to inspect docker image [%s]", reference)
		state.Internal = &shared.Identity{
			Hash:    state.Output,
			Version: params.GetVersion(),
		}
		fmt.Printf("docker image inspect %s\n", state.Output)
		return nil
	}

	state.Logger.Errorf("Could not find an image to inspect for reference [%s]", reference)
	return ErrorNoBuild
}

func listContainers(containers []container.Summary, imageMap map[string]image.Summary, output *bytes.Buffer, threshold time.Time) {
	var writer = tabwriter.NewWriter(output, 1, 1, 3, ' ', 0)

	_, _ = fmt.Fprintf(writer, "CONTAINER\tID\tIMAGE\tSTATUS\tCREATED\n")

	for _, ct := range containers {
		if img, ok := imageMap[ct.ImageID]; ok {
			var created = formatCreated(ct.Created, threshold)
			var imageRepoTag, name string

			for _, tag := range img.RepoTags {
				if !strings.HasSuffix(tag, "latest") {
					imageRepoTag = tag
					break
				}
			}

			if len(ct.Names) > 0 {
				name = ct.Names[0][1:]
			}

			_, _ = fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", name, ct.ID[0:11], imageRepoTag, ct.Status, created)
		}
	}

	panicz.PanicIfError(writer.Flush())
}

func listImages(images []image.Summary, output *bytes.Buffer, threshold time.Time) {
	var writer = tabwriter.NewWriter(output, 1, 1, 3, ' ', 0)

	_, _ = fmt.Fprintf(writer, "REPOSITORY\tVERSION\tIMAGE ID\tCREATED\tSIZE\n")

	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			var created = formatCreated(img.Created, threshold)
			var parts = strings.Split(repoTag, ":")
			var hash = img.ID[7:19]
			var size = fmt.Sprintf("%3.1fMB", float64(img.Size/1024/1024))
			var version string

			if len(parts) == 2 {
				version = parts[1]
			}

			_, _ = fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", parts[0], version, hash, created, size)
		}
	}

	panicz.PanicIfError(writer.Flush())
}
