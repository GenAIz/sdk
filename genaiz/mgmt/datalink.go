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

type UserDataLinkFacade Facade[[]UserDataLink, broker.DataLinkListParams]

type ListLinksTaskFactory func() *task.Task[broker.DataLinkListParams]
type ReduceLinksTaskFactory func() *task.Task[broker.DataLinkListParams]

type UserDataLink struct {
	Id          int64 `cli:"Id"`
	Oem         string
	Handle      string
	Fqdn        string `cli:"Fqdn"`
	Version     string `cli:"Version"`
	Name        string `cli:"Name"`
	Description string
	Created     int64 `cli:"Created"`
	Modified    int64
	Local       bool `cli:"Local?"`
	Released    bool `cli:"Released?"`
	Flags       *int
}

func (udl UserDataLink) MarshalJSON() ([]byte, error) {
	var created, modified string

	if udl.Created > 0 {
		created = createdFormatter.FormatMillis(udl.Created)
	}

	if udl.Modified > 0 {
		modified = createdFormatter.FormatMillis(udl.Modified)
	}

	return json.Marshal(&struct {
		Id          int64  `json:"id,omitempty"`
		Oem         string `json:"oem"`
		Handle      string `json:"handle"`
		Version     string `json:"version"`
		Fqdn        string `json:"fqdn"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Created     string `json:"created,omitempty"`
		Modified    string `json:"modified,omitempty"`
		Flags       *int   `json:"flags,omitempty"`
	}{
		Id:          udl.Id,
		Oem:         udl.Oem,
		Handle:      udl.Handle,
		Version:     udl.Version,
		Fqdn:        udl.Fqdn,
		Name:        udl.Name,
		Description: udl.Description,
		Created:     created,
		Modified:    modified,
		Flags:       udl.Flags,
	})
}

func (udl UserDataLink) MarshalSlice() ([]string, error) {
	var created string

	if udl.Created > 0 {
		created = createdFormatter.FormatMillis(udl.Created)
	} else {
		created = "-"
	}

	return []string{
		cast.ToString(udl.Id),
		udl.Fqdn,
		udl.Version,
		udl.Name,
		created,
		stringz.YesOrNo(udl.Local),
		stringz.YesOrNo(udl.Released),
	}, nil
}

func (udl UserDataLink) Match(filter string) bool {
	if filter != "" {
		var lowFilter = strings.ToLower(filter)

		if _, err := strconv.Atoi(filter); err == nil {
			return strings.HasPrefix(cast.ToString(udl.Id), filter)
		}

		return strings.EqualFold(udl.Oem, filter) ||
			strings.HasPrefix(udl.Oem, lowFilter) ||
			strings.HasPrefix(udl.Fqdn, lowFilter) ||
			strings.HasPrefix(fmt.Sprintf("%s:%s", udl.Fqdn, udl.Version), lowFilter)
	}

	return true
}

func ToUserDataLink(dataLink *broker.DataLink) *UserDataLink {
	var result = &UserDataLink{
		Oem:         dataLink.Oem,
		Handle:      dataLink.Handle,
		Fqdn:        dataLink.GetFqdn(),
		Version:     dataLink.GetVersion(),
		Name:        dataLink.Name,
		Description: dataLink.Description,
		Local:       dataLink.Created == nil,
		Released:    dataLink.IsReleased(),
		Flags:       dataLink.Flags,
	}

	if dataLink.Id != nil {
		result.Id = *dataLink.Id
	}

	if dataLink.Created != nil {
		result.Created = *dataLink.Created
	}

	if dataLink.Modified != nil {
		result.Modified = *dataLink.Modified
	}

	return result
}

type userDataLinksFacade struct {
	baseLoggingFacade
	params *broker.DataLinkListParams
}

func (udf userDataLinksFacade) Filtering(filter string) Provider[[]UserDataLink] {
	return &userDataLinksProvider{
		Plan: task.Plan{
			Logger: udf.logger,
		},
		filter:                 filter,
		params:                 udf.params,
		listLinksTaskFactory:   broker.NewDataLinkListTask,
		reduceLinksTaskFactory: broker.NewDataLinkReduceTask,
	}
}

func (udf userDataLinksFacade) Provider() Provider[[]UserDataLink] {
	return udf.Filtering("")
}

func (udf userDataLinksFacade) WithLogger(logger *logrus.Logger) Facade[[]UserDataLink, broker.DataLinkListParams] {
	udf.logger = logger
	return udf
}

func (udf userDataLinksFacade) WithParams(params *broker.DataLinkListParams) Facade[[]UserDataLink, broker.DataLinkListParams] {
	udf.params = params
	return udf
}

type userDataLinksProvider struct {
	task.Plan
	filter                 string
	params                 *broker.DataLinkListParams
	listLinksTaskFactory   ListLinksTaskFactory
	reduceLinksTaskFactory ReduceLinksTaskFactory
}

func (udp *userDataLinksProvider) Get() ([]UserDataLink, task.Error) {
	var dataLinks []broker.DataLink
	var workers []task.Worker
	var failure interface{}

	udp.OnReturn = func(i interface{}) { dataLinks = i.([]broker.DataLink) }
	udp.OnFailure = func(i interface{}) { failure = i }

	if udp.params.Oem != "" {
		workers = append(workers, task.NewWorker(udp.params, udp.listLinksTaskFactory()))
	}

	// Always reduce, no matter what, local solutions should just be flagged
	workers = append(workers, task.NewWorker(udp.params, udp.reduceLinksTaskFactory()))
	udp.Sequence(workers...)

	if failure == nil {
		var result = make([]UserDataLink, 0)

		for _, dl := range dataLinks {
			var dataLink = ToUserDataLink(&dl)

			if dataLink.Match(udp.filter) {
				result = append(result, *dataLink)
			}
		}

		if len(result) > 1 {
			slices.SortFunc(result, func(a, b UserDataLink) int {
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

func NewUserDataLinkFacade() UserDataLinkFacade {
	return &userDataLinksFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}
