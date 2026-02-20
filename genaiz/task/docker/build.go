package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/moby/go-archive"
	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/mapz"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/lang"
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
	Dockerfile       string
	DockerContext    string
	DockerRepository string
	DockerVersion    string
	Label            bool
	NoCache          bool
	Prune            bool
	Streams          *BuildStreams
}

type BuildOutput struct {
	Stream string
}

type BuildStreams struct {
	Err *os.File
	In  *os.File
	Out *os.File
}

func (p *BuildParams) GetFilters() filters.Args {
	return filters.NewArgs(
		filters.Arg("reference", p.GetReference()),
	)
}

func (p *BuildParams) GetFiltersByRepo(field string) filters.Args {
	var value string

	if i := strings.Index(p.DockerRepository, "/"); i >= 0 {
		value = p.DockerRepository[i+1:]
	} else {
		value = p.DockerRepository
	}

	return filters.NewArgs(filters.Arg(field, value+"*"))
}

func (p *BuildParams) GetFiltersByVersion() filters.Args {
	if p.DockerVersion != "latest" {
		return p.GetFilters()
	}

	return filters.NewArgs(
		filters.Arg("reference", p.DockerRepository+":*"),
	)
}

func (p *BuildParams) GetReference() string {
	return stringz.SingleTagLabel(p.DockerRepository, ":", p.GetVersion())
}

func (p *BuildParams) GetVersion() string {
	return stringz.FirstNonEmpty(p.DockerVersion, "latest")
}

func (p *BuildParams) getErrorStream() *os.File {
	if p.Streams != nil && p.Streams.Err != nil {
		return p.Streams.Err
	}

	return os.Stderr
}

func (p *BuildParams) getInputStream() *os.File {
	if p.Streams != nil && p.Streams.In != nil {
		return p.Streams.In
	}

	return os.Stdin
}

func (p *BuildParams) getOutputStream() *os.File {
	if p.Streams != nil && p.Streams.Out != nil {
		return p.Streams.Out
	}

	return os.Stdout
}

func (p *BuildParams) toBuildArgs() []string {
	var args = []string{"build", "--pull", "-t", p.GetReference()}

	if p.Dockerfile != "" {
		args = append(args, "-f", p.Dockerfile)
	}

	if p.Label {
		args = append(args, "--label")
	}

	if p.NoCache {
		args = append(args, "--no-cache")
	}

	// Needs to be last
	if p.DockerContext == "" {
		args = append(args, ".")
	} else {
		args = append(args, p.DockerContext)
	}

	return args
}

func (p *BuildParams) toPruneArgs() []string {
	var result = []string{"buildx", "prune", "--force"}

	if !p.NoCache {
		// We'll keep half a day's worth of cached build artifacts to help the build process
		result = append(result, "--filter", "until=12h")
	}

	return result
}

func NewBuildTask() *task.Task[BuildParams] {
	if dockerPath, err := exec.LookPath("docker"); err == nil {
		if os.Geteuid() == 0 {
			// Since we are forking without the ability to validate the identity of the binary, prevent root from using build
			panic("can not invoke the docker binary as root when building")
		}

		return NewBuildForkTask(dockerPath)
	} else {
		fmt.Println("DEPRECATED: The legacy Moby builder is deprecated and will be removed in a future release.\n" +
			"You should install the docker-cli with docker-buildx to build images: https://docs.docker.com/go/buildx/")
		return NewBuildLegacyTask()
	}
}

func NewBuildLegacyTask() *task.Task[BuildParams] {
	return &task.Task[BuildParams]{
		Name:         "docker-build-legacy",
		OnPrepare:    handleBuildContext,
		OnIncomplete: handleBuildLegacyCreate,
		OnComplete:   handleBuildPrune,
		OnPretend:    handleBuildPretend,
	}
}

func NewBuildForkTask(dockerPath string) *task.Task[BuildParams] {
	return &task.Task[BuildParams]{
		Name:         "docker-build-fork",
		OnPrepare:    handleBuildContext,
		OnIncomplete: lang.Assists(dockerPath, handleBuildForkCreate),
		OnComplete:   lang.Assists(dockerPath, handleBuildForkPrune),
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

func formatCreated(createdSeconds int64, threshold time.Time) string {
	var createdTime = time.UnixMilli(createdSeconds * 1000)

	if createdTime.After(threshold) {
		return createdTime.Format(time.TimeOnly)
	} else {
		return createdTime.Format(time.DateOnly)
	}
}

func handleBuildContext(params *BuildParams, state *task.State) error {
	var dockerClient = dockerFactory.Get()
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

func handleBuildForkCreate(dockerPath string, params *BuildParams, state *task.State) error {
	var args = params.toBuildArgs()
	var cmd = exec.CommandContext(params.Context, dockerPath, args...)
	var fork = forkFactory.Get(cmd)
	var err error

	if state.Logger.IsLevelEnabled(logrus.DebugLevel) {
		fork.WithPipeErr(stateError(state))
	} else {
		fork.WithStdErr(params.getErrorStream()).
			WithStdOut(params.getOutputStream())
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
			var args = params.toPruneArgs()
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
	var buildCtx, _ = archive.TarWithOptions(params.DockerContext, &archive.TarOptions{})
	var options = build.ImageBuildOptions{
		Dockerfile: params.Dockerfile,
		Tags:       []string{reference},
		Remove:     true,
		PullParent: true,
		NoCache:    params.NoCache,
	}

	if params.Label {
		options.Labels = map[string]string{"sf": params.DockerRepository}
	}

	state.Logger.Debugf("Building a docker image tagged [%s] with the legacy builder", reference)

	if resp, err := dockerClient.ImageBuild(params.Context, buildCtx, options); err == nil {
		var scanner = bufio.NewScanner(resp.Body)
		defer filez.CloseSilently(resp.Body)

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
	} else {
		return err
	}
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
			var pruneFilters = filters.NewArgs(
				filters.Arg("label", "sf="+params.DockerRepository),
			)

			if report, err := dockerClient.ImagesPrune(params.Context, pruneFilters); err == nil {
				for _, deleted := range report.ImagesDeleted {
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
		var listFilters = params.GetFilters()
		var summaries []image.Summary
		var err error

		if summaries, err = dockerClient.ImageList(params.Context, image.ListOptions{Filters: listFilters}); err == nil {
			if len(summaries) >= 1 {
				var img = summaries[0]
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
	var imageFilters = params.GetFiltersByVersion()
	var images []image.Summary
	var err error

	if images, err = dockerClient.ImageList(params.Context, image.ListOptions{Filters: imageFilters}); err == nil {
		var dockerState = NewClientState(state)

		if len(images) > 0 {
			var containers []container.Summary
			var containerOptions = container.ListOptions{
				All:     true,
				Filters: params.GetFiltersByRepo("name"),
			}

			dockerState.AddImages(images...)

			if containers, err = dockerClient.ContainerList(params.Context, containerOptions); err == nil {
				dockerState.AddContainers(containers...)
				return nil
			}
		} else {
			state.Logger.Debugf("Could not list images for the provided filter")
			return ErrorNoBuild
		}
	}

	return err
}

func handleListComplete(params *BuildParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if dockerState.HasImages() {
		var year, month, day = time.Now().In(time.Local).Date()
		var today = time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		var imgList = dockerState.GetImages()
		var output bytes.Buffer

		state.Logger.Debugf("Listing Docker Images for Smart Function [%s], Version [%s]", params.DockerRepository, params.DockerVersion)
		listImages(imgList, &output, today)
		state.Output = output.String()
		output.Reset()

		if dockerState.HasContainers() {
			var imagesById = mapz.Mapped(imgList, func(summary image.Summary) string {
				return summary.ID
			})

			state.Logger.Debugf("Listing Docker Containers for Smart Function [%s], Version [%s]", params.DockerRepository, params.DockerVersion)
			listContainers(dockerState.GetContainers(), imagesById, &output, today)
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
		var resp image.InspectResponse
		var err error

		state.Logger.Debugf("Inspecting docker image [%s]", state.Output)

		if resp, err = dockerClient.ImageInspect(params.Context, state.Output); err == nil {
			var digest string

			if len(resp.RepoDigests) > 0 {
				if i := strings.LastIndex(resp.RepoDigests[0], hashPrefix); i >= 0 {
					digest = resp.RepoDigests[0][i:]
				} else {
					state.Logger.Errorf("Found invalid repo digest [%s] for image [%s]", resp.RepoDigests[0], state.Output)
				}
			}

			state.Internal = &shared.Identity{
				Id:      resp.ID,
				Hash:    digest,
				Version: params.GetVersion(),
			}
			return nil
		}

		return err
	}

	return ErrorNoBuild
}

func handleInspectContext(params *BuildParams, state *task.State) error {
	if state.Output == "" {
		var err error

		if err = handleBuildContext(params, state); err != nil {
			if errors.Is(err, ErrorLatestBuild) {
				return nil
			}
		}

		if err == nil && state.Output == "" {
			return ErrorNoBuild
		}

		return err
	}

	return nil
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
	return state.Error
}

func listContainers(containers []container.Summary, imageMap map[string]image.Summary, output *bytes.Buffer, threshold time.Time) {
	var writer = tabwriter.NewWriter(output, 1, 1, 3, ' ', 0)

	_, _ = fmt.Fprintf(writer, "CONTAINER\tID\tIMAGE\tSTATUS\tCREATED\n")

	for _, ct := range containers {
		if img, ok := imageMap[ct.ImageID]; ok {
			var created = formatCreated(ct.Created, threshold)
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
