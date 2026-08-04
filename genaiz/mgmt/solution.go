package mgmt

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserSolutionFacade Facade[[]UserSolution, broker.SolutionListParams]

type solutionListTaskFactory func() *task.Task[broker.SolutionListParams]
type solutionReduceTaskFactory func() *task.Task[broker.SolutionListParams]

type UserSolution struct {
	Id          int64 `cli:"Id"`
	Oem         string
	Handle      string
	Fqdn        string `cli:"Fqdn"`
	Version     string `cli:"Version"`
	Name        string `cli:"Name"`
	Description string
	Digest      string
	Created     int64 `cli:"Created"`
	Modified    int64
	Local       bool `cli:"Local?"`
	Released    bool `cli:"Published?"`
	Flags       *int

	matched string
}

func (us UserSolution) Match(filter string) *UserSolution {
	var lowFilter = strings.ToLower(filter)
	var fqdnVersion = fmt.Sprintf("%s:%s", us.Fqdn, us.Version)
	var matched string

	if _, err := strconv.Atoi(filter); err == nil {
		if strings.HasPrefix(cast.ToString(us.Id), filter) {
			matched = cast.ToString(us.Id)
		}
	} else if strings.EqualFold(us.Oem, filter) ||
		strings.HasPrefix(us.Oem, lowFilter) ||
		strings.HasPrefix(us.Fqdn, lowFilter) ||
		strings.HasPrefix(fqdnVersion, lowFilter) {
		// On autocomplete listings, we don't want to specify the sequence number in the CLI query
		matched = fmt.Sprintf("%s:%s", us.Fqdn, strings.Split(us.Version, "-")[0])
	}

	if matched == "" {
		return nil
	}

	return &UserSolution{
		Id:          us.Id,
		Oem:         us.Oem,
		Handle:      us.Handle,
		Fqdn:        us.Fqdn,
		Version:     us.Version,
		Name:        us.Name,
		Description: us.Description,
		Digest:      us.Digest,
		Created:     us.Created,
		Modified:    us.Modified,
		Local:       us.Local,
		Released:    us.Released,
		Flags:       us.Flags,
		matched:     matched,
	}
}

func (us UserSolution) Matched() string {
	if us.matched == "" {
		return fmt.Sprintf("%s:%s", us.Fqdn, us.Version)
	}

	return us.matched
}

func (us UserSolution) MarshalJSON() ([]byte, error) {
	var created, modified, readableDigest string

	if strings.Contains(us.Digest, "sha256") &&
		len(us.Digest) >= 8 {
		readableDigest = us.Digest[7:]
	}

	if us.Created > 0 {
		created = createdFormatter.FormatMillis(us.Created)
	}

	if us.Modified > 0 {
		modified = createdFormatter.FormatMillis(us.Modified)
	}

	return json.Marshal(&struct {
		Id          int64  `json:"id,omitempty"`
		Oem         string `json:"oem"`
		Handle      string `json:"handle"`
		Version     string `json:"version"`
		Fqdn        string `json:"fqdn"`
		Digest      string `json:"digest,omitempty"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Created     string `json:"created,omitempty"`
		Modified    string `json:"modified,omitempty"`
		Flags       *int   `json:"flags,omitempty"`
	}{
		Id:          us.Id,
		Oem:         us.Oem,
		Handle:      us.Handle,
		Version:     us.Version,
		Fqdn:        us.Fqdn,
		Digest:      readableDigest,
		Name:        us.Name,
		Description: us.Description,
		Created:     created,
		Modified:    modified,
		Flags:       us.Flags,
	})
}

func (us UserSolution) MarshalSlice() ([]string, error) {
	var created string

	if us.Created > 0 {
		created = createdFormatter.FormatMillis(us.Created)
	} else {
		created = "-"
	}

	return []string{
		cast.ToString(us.Id),
		us.Fqdn,
		us.Version,
		us.Name,
		created,
		stringz.YesOrNo(us.Local),
		stringz.YesOrNo(us.Released),
	}, nil
}

func ToUserSolution(sol *broker.Solution) *UserSolution {
	var result = &UserSolution{
		Oem:         sol.Oem,
		Handle:      sol.Handle,
		Version:     sol.GetVersion(),
		Fqdn:        sol.GetFqdn(),
		Digest:      stringz.NilToEmpty(sol.Digest),
		Name:        sol.Name,
		Description: sol.Description,
		Local:       sol.Digest == nil,
		Released:    sol.IsReleased(),
		Flags:       sol.Flags,
	}

	if sol.Id != nil {
		result.Id = *sol.Id
	}

	if sol.Created != nil {
		result.Created = *sol.Created
	}

	if sol.Modified != nil {
		result.Modified = *sol.Modified
	}

	return result
}

type userSolutionsFacade struct {
	baseLoggingFacade
	params *broker.SolutionListParams
}

func (usf *userSolutionsFacade) Filtering(filter string) Provider[[]UserSolution] {
	return &userSolutionsProvider{
		Plan: task.Plan{
			Logger: usf.logger,
		},
		filter:                    filter,
		params:                    usf.params,
		solutionListTaskFactory:   broker.NewSolutionListTask,
		solutionReduceTaskFactory: broker.NewSolutionReduceTask,
	}
}

func (usf *userSolutionsFacade) Provider() Provider[[]UserSolution] {
	return usf.Filtering("")
}

func (usf *userSolutionsFacade) WithLogger(logger *logrus.Logger) Facade[[]UserSolution, broker.SolutionListParams] {
	usf.logger = logger
	return usf
}

func (usf *userSolutionsFacade) WithParams(params *broker.SolutionListParams) Facade[[]UserSolution, broker.SolutionListParams] {
	usf.params = params
	return usf
}

type userSolutionsProvider struct {
	task.Plan
	filter                    string
	params                    *broker.SolutionListParams
	solutionListTaskFactory   solutionListTaskFactory
	solutionReduceTaskFactory solutionReduceTaskFactory
}

func (usp userSolutionsProvider) Get() ([]UserSolution, task.Error) {
	var solutions []broker.Solution
	var workers []task.Worker
	var failure interface{}

	usp.OnReturn = func(i interface{}) { solutions = i.([]broker.Solution) }
	usp.OnFailure = func(i interface{}) { failure = i }

	if usp.params.Oem != "" {
		workers = append(workers, task.NewWorker(usp.params, usp.solutionListTaskFactory()))
	}

	// Always reduce, no matter what, local solutions should just be flagged
	workers = append(workers, task.NewWorker(usp.params, usp.solutionReduceTaskFactory()))
	usp.Sequence(workers...)

	if failure == nil {
		var result = make([]UserSolution, 0)

		for _, sol := range solutions {
			var solution = ToUserSolution(&sol)

			if usp.filter == "" {
				result = append(result, *solution)
			} else if matched := solution.Match(usp.filter); matched != nil {
				result = append(result, *matched)
			}
		}

		if len(result) > 1 {
			slices.SortFunc(result, func(a, b UserSolution) int {
				// Reverse ordering
				if a.Created > b.Created {
					return -1
				} else if a.Created < b.Created {
					return 1
				}

				return 0
			})
		}

		return result, nil
	}

	return nil, task.NewFailure(failure)
}

func NewUserSolutionFacade() UserSolutionFacade {
	return &userSolutionsFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}
