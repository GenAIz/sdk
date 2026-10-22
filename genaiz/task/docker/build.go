package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/moby/go-archive"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/mapz"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker/adapt"
	"genaiz.com/genaiz/task/shared"
)

const (
	hashPrefix = "sha256:"
)

var (
	BuildTasks struct {
		LegacyTaskFactory   BuildTaskFactory // LegacyTaskFactory builds a task which can be used with an instance of BuildParams to launch a Docker build with the legacy deprecated builder of dockerd
		StandardTaskFactory BuildTaskFactory // StandardTaskFactory builds a task which can be used with an instance of BuildParams to launch a Docker build with the buildx plugin

		InspectTaskFactory InspectTaskFactory // InspectTaskFactory builds a task which can be used with an instance of BuildParams to launch a Docker image inspect command.
		ListTaskFactory    ListTaskFactory    // ListTaskFactory builds a task which can be used with an instance of BuildParams to query images from dockerd
	}
)

type BuildTask = *task.Task[BuildParams]
type BuildTaskFactory = func() BuildTask
type InspectTask = *task.Task[BuildParams]
type InspectTaskFactory = func() InspectTask
type ListTask = *task.Task[BuildParams]
type ListTaskFactory = func() ListTask

type BuildOutput struct {
	Stream string
}

type BuildParams struct {
	task.Env
	adapt.BuildArguments
	Prune bool
}

func handleBuildContext(params *BuildParams, state *task.State) error {
	var dockerClient = dockerFactory.Get()
	var version = params.GetVersion()
	var reference = params.GetReference()
	var listOptions = client.ImageListOptions{
		Filters: params.GetFilters(),
	}
	var result client.ImageListResult
	var err error

	state.Logger.Debugf("Finding a docker image for reference [%s]", reference)

	if result, err = dockerClient.ImageList(params.Context, listOptions); err == nil {
		if len(result.Items) == 0 {
			state.Logger.Debugf("Could not find an image for reference [%s]", reference)
			err = Errors.ListImageNotFound(reference)
		} else if version == "latest" {
			state.Logger.Debugf("Version latest requires a fresh build")
			state.Output = result.Items[0].ID
			err = Errors.BuildImageRebuildRequired()
		} else {
			state.Logger.Warningf("Found image with reference [%s]", reference)
			state.Output = result.Items[0].ID
			return nil
		}
	}

	return err
}

func handleBuildForkCreate(dockerPath string, params *BuildParams, state *task.State) error {
	var args = params.CliArguments()
	var cmd = exec.CommandContext(params.Context, dockerPath, args...)
	var fork = forkFactory.Get(cmd)
	var err error

	if state.Logger.IsLevelEnabled(logrus.DebugLevel) {
		fork.WithPipeErr(stateError(state))
	} else {
		fork.WithStdErr(params.ErrorStream()).
			WithStdOut(params.OutputStream())
	}

	state.Logger.Debugf("Building a docker image with cmd [%s]", cmd.String())

	if err = fork.Run(params.Context); err == nil {
		if err = fork.GetWaitError(); err == nil {
			state.Output = ""
			return nil
		}
	}

	return err
}

func handleBuildForkPrune(dockerPath string, params *BuildParams, state *task.State) error {
	if state.Error == nil {
		if params.Prune {
			var args = params.PruneArguments()
			var cmd = exec.CommandContext(params.Context, dockerPath, args...)
			var fork = forkFactory.Get(cmd)
			var err error

			if state.Logger.IsLevelEnabled(logrus.DebugLevel) {
				fork.WithPipeOut(stateDebug(state))
			}

			state.Logger.Debugf("Pruning build cache with command [%s]", cmd.String())

			if err = fork.Run(params.Context); err != nil {
				state.Logger.Errorf("Prune failed with error %s", err)
			}
		} else {
			state.Logger.Debugf("Pruning dangling images disabled, skipping")
		}

		return handleBuildReport(params, state)
	}

	return state.Error
}

func handleBuildLegacyCreate(params *BuildParams, state *task.State) error {
	var dockerClient = dockerFactory.Get()
	var reference = params.GetReference()
	var buildReader io.ReadCloser
	var err error

	state.Logger.Debugf("Building a docker image tagged [%s] with the legacy builder", reference)

	if buildReader, err = archive.TarWithOptions(params.DockerContext, &archive.TarOptions{}); err == nil {
		var options = client.ImageBuildOptions{
			Dockerfile: params.Dockerfile,
			Tags:       []string{reference},
			Remove:     true,
			Platforms:  params.GetPlatforms(),
			PullParent: true,
			NoCache:    params.NoCache,
		}
		var result client.ImageBuildResult

		defer filez.CloseSilently(buildReader)

		if params.Label {
			options.Labels = map[string]string{"sf": params.DockerRepository}
		}

		if result, err = dockerClient.ImageBuild(params.Context, buildReader, options); err == nil {
			var scanner = bufio.NewScanner(result.Body)
			defer filez.CloseSilently(result.Body)

			for scanner.Scan() {
				var output BuildOutput

				if err = json.Unmarshal(scanner.Bytes(), &output); err == nil {
					if output.Stream != "" {
						state.Logger.Debugf("%s", strings.TrimSuffix(output.Stream, "\n"))
					}
				} else {
					state.Logger.Warningf("Could not parse json with error: %s", err)
					state.Logger.Debugf("String: %s", scanner.Text())
				}
			}

			state.Output = ""
			return nil
		}
	}

	return err
}

func handleBuildPretend(params *BuildParams, state *task.State) error {
	var reference = params.GetReference()

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
			fmt.Printf("docker image prune --filter label=%s=%s\n", "sf", params.DockerRepository)
		}
	}

	return nil
}

func handleBuildPrune(params *BuildParams, state *task.State) error {
	if state.Error == nil {
		if params.Prune {
			var dockerClient = dockerFactory.Get()
			var pruneOptions = client.ImagePruneOptions{
				Filters: make(client.Filters).
					Add("label", "sf="+params.DockerRepository),
			}

			if result, err := dockerClient.ImagePrune(params.Context, pruneOptions); err == nil {
				for _, deleted := range result.Report.ImagesDeleted {
					state.Logger.Debugf("Removed dangling image id [%s], no longer in use", deleted.Deleted)
				}
			} else {
				state.Logger.Errorf("Prune failed with error %s", err)
			}
		} else {
			state.Logger.Debugf("Pruning dangling images disabled, skipping")
		}

		return handleBuildReport(params, state)
	}

	return state.Error
}

func handleBuildReport(params *BuildParams, state *task.State) error {
	if state.Error == nil {
		var dockerClient = dockerFactory.Get()
		var listOptions = client.ImageListOptions{
			Filters: params.GetFilters(),
		}
		var result client.ImageListResult
		var err error

		if result, err = dockerClient.ImageList(params.Context, listOptions); err == nil {
			if len(result.Items) >= 1 {
				var img = result.Items[0]
				var prefixLength = len(hashPrefix)
				var shortId = img.ID[prefixLength : prefixLength+12]
				var size = fmt.Sprintf("%3.1fMB", float64(img.Size/1024/1024))

				state.Reportf("Built image %s - %s size %s", params.GetReference(), shortId, size)
				return nil
			}
		}

		return err
	}

	return state.Error
}

func handleListContext(params *BuildParams, state *task.State) error {
	var dockerClient = dockerFactory.Get()
	var listOptions = client.ImageListOptions{
		Filters: params.GetFiltersByVersion(),
	}
	var imageResult client.ImageListResult
	var err error

	if imageResult, err = dockerClient.ImageList(params.Context, listOptions); err == nil {
		var dockerState = NewClientState(state)

		if len(imageResult.Items) > 0 {
			var containerOptions = client.ContainerListOptions{
				All:     true,
				Filters: params.GetFiltersByRepo("name"),
			}
			var containerResult client.ContainerListResult

			dockerState.AddImages(imageResult.Items...)

			if containerResult, err = dockerClient.ContainerList(params.Context, containerOptions); err == nil {
				dockerState.AddContainers(containerResult.Items...)
				return nil
			}
		} else {
			state.Logger.Debugf("Could not list images for the provided filter")
			return Errors.BuildImageNotFound()
		}
	}

	return err
}

func handleListComplete(params *BuildParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if dockerState.HasImages() {
		var formatter = timez.NewTodayFormatter()
		var imgList = dockerState.GetImages()
		var output bytes.Buffer

		state.Logger.Debugf("Listing Docker Images for Smart Function [%s], Version [%s]", params.DockerRepository, params.DockerVersion)
		listImages(imgList, &output, formatter)
		state.Output = output.String()
		output.Reset()

		if dockerState.HasContainers() {
			var imagesById = mapz.Mapped(imgList, func(summary image.Summary) string {
				return summary.ID
			})

			state.Logger.Debugf("Listing Docker Containers for Smart Function [%s], Version [%s]", params.DockerRepository, params.DockerVersion)
			listContainers(dockerState.GetContainers(), imagesById, &output, formatter)
			state.Output += "\n" + output.String()
		} else {
			state.Output += "\nNo containers associated with the listed images\n"
		}
	} else {
		state.Output = fmt.Sprintf("No images associated with Smart Function [%s]\n", params.DockerRepository)
	}

	dockerState.Reset()
	return nil
}

func handleInspectComplete(params *BuildParams, state *task.State) error {
	if state.Output != "" {
		var dockerClient = dockerFactory.Get()
		var result client.ImageInspectResult
		var err error

		state.Logger.Debugf("Inspecting docker image [%s]", state.Output)

		if result, err = dockerClient.ImageInspect(params.Context, state.Output); err == nil {
			var digest string

			if len(result.RepoDigests) > 0 {
				if i := strings.LastIndex(result.RepoDigests[0], hashPrefix); i >= 0 {
					digest = result.RepoDigests[0][i:]
				} else {
					state.Logger.Errorf("Found invalid repo digest [%s] for image [%s]", result.RepoDigests[0], state.Output)
				}
			}

			state.Internal = &shared.Identity{
				Id:      result.ID,
				Hash:    digest,
				Version: params.GetVersion(),
			}
			return nil
		}

		return err
	}

	return Errors.BuildImageNotFound()
}

func handleInspectContext(params *BuildParams, state *task.State) error {
	if state.Output == "" {
		var err error

		if err = handleBuildContext(params, state); err != nil {
			if errors.Is(err, Errors.BuildImageRebuildRequired()) {
				state.Internal = nil
				return nil
			}
		}

		if err == nil && state.Output == "" {
			return Errors.BuildImageNotFound()
		}

		return err
	}

	return nil
}

func handleInspectPretend(params *BuildParams, state *task.State) error {
	var reference = params.GetReference()

	if state.Error == nil || errors.Is(state.Error, Errors.BuildImageRebuildRequired()) {
		state.Logger.Debugf("Pretending to inspect docker image [%s]", reference)
		state.Internal = &shared.Identity{
			Hash:    state.Output,
			Version: params.GetVersion(),
		}
		fmt.Printf("docker image inspect %s\n", state.Output)
		return nil
	}

	state.Logger.Errorf("Could not find an image to inspect for reference [%s]", reference)
	return state.Error
}

func listContainers(containers []container.Summary, imageMap map[string]image.Summary, output *bytes.Buffer, formatter timez.Formatter) {
	var writer = tabwriter.NewWriter(output, 1, 1, 3, ' ', 0)

	_, _ = fmt.Fprintf(writer, "CONTAINER\tID\tIMAGE\tSTATUS\tCREATED\n")

	for _, ct := range containers {
		if img, ok := imageMap[ct.ImageID]; ok {
			var created = formatter.FormatSeconds(ct.Created)
			var imageRepoTag, name string

			if len(img.RepoTags) >= 1 {
				imageRepoTag = img.RepoTags[0]
			} else {
				imageRepoTag = img.ID[7:19]
			}

			if len(ct.Names) > 0 {
				name = ct.Names[0][1:]
			}

			_, _ = fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", name, ct.ID[7:19], imageRepoTag, ct.Status, created)
		}
	}

	panicz.PanicIfError(writer.Flush())
}

func listImages(images []image.Summary, output *bytes.Buffer, formatter timez.Formatter) {
	var writer = tabwriter.NewWriter(output, 1, 1, 3, ' ', 0)

	_, _ = fmt.Fprintf(writer, "REPOSITORY\tVERSION\tIMAGE ID\tCREATED\tSIZE\n")

	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			var created = formatter.FormatSeconds(img.Created)
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

func init() {
	BuildTasks.LegacyTaskFactory = newBuildLegacyTask
	BuildTasks.StandardTaskFactory = newBuildTask
	BuildTasks.InspectTaskFactory = newInspectTask
	BuildTasks.ListTaskFactory = newListTask
}

func newBuildForkTask(dockerPath string) *task.Task[BuildParams] {
	return &task.Task[BuildParams]{
		Name:         "docker-build-fork",
		OnPrepare:    handleBuildContext,
		OnIncomplete: lang.Assists(dockerPath, handleBuildForkCreate),
		OnComplete:   lang.Assists(dockerPath, handleBuildForkPrune),
		OnPretend:    handleBuildPretend,
	}
}

func newBuildLegacyTask() *task.Task[BuildParams] {
	return &task.Task[BuildParams]{
		Name:         "docker-build-legacy",
		OnPrepare:    handleBuildContext,
		OnIncomplete: handleBuildLegacyCreate,
		OnComplete:   handleBuildPrune,
		OnPretend:    handleBuildPretend,
	}
}

func newBuildTask() *task.Task[BuildParams] {
	if dockerPath, err := exec.LookPath("docker"); err == nil {
		if os.Geteuid() == 0 {
			// Since we are forking without the ability to validate the identity of the binary, prevent root from using build
			panic("can not invoke the docker binary as root when building")
		}

		return newBuildForkTask(dockerPath)
	}

	fmt.Println("DEPRECATED: The legacy Moby builder is deprecated and will be removed in a future release.\n" +
		"You should install the docker-cli with docker-buildx to build images: https://docs.docker.com/go/buildx/")
	return newBuildLegacyTask()
}

func newInspectTask() *task.Task[BuildParams] {
	return &task.Task[BuildParams]{
		Name:       "docker-inspect",
		OnPrepare:  handleInspectContext,
		OnComplete: handleInspectComplete,
		OnPretend:  handleInspectPretend,
	}
}

func newListTask() *task.Task[BuildParams] {
	return &task.Task[BuildParams]{
		Name:       "docker-list",
		OnPrepare:  handleListContext,
		OnComplete: handleListComplete,
	}
}

func stateDebug(state *task.State) func(string) {
	return func(s string) {
		state.Logger.Debugf("%s", s)
	}
}

func stateError(state *task.State) func(string) {
	var indexRegex = regexp.MustCompile(`^#\d+\s+`)

	return func(s string) {
		if i := strings.Index(strings.ToLower(s), "error: "); i == 0 {
			state.Logger.Errorf("%s", s[7:])
		} else if s != "" {
			// docker build logs everything to STDERR, including debug messages
			if !strings.HasPrefix(s, "#") {
				state.Logger.Debugf("%s", s)
			} else if i = strings.Index(s, " "); i > 0 && indexRegex.Match([]byte(s)) {
				state.Logger.Debugf("%s", s[i+1:])
			}
		}
	}
}
