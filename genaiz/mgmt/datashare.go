package mgmt

import (
	"cmp"
	"encoding/json"
	"slices"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type DataInstanceListTaskFactory func() *task.Task[broker.DataInstanceListParams]

// UserLinkInstance is an adapter for broker.DataLinkInstance used to display both DataStores and DataSources, which share the same definition
type UserLinkInstance struct {
	Id          *int64 `cli:"Id"`
	Name        string `cli:"Name"`
	Description string
	Oem         string
	Handle      string
	Fqdn        string `cli:"Fqdn"`
	Version     string `cli:"Version"`
	Sequence    *int
	Flags       *int
	Visibility  string
	Properties  map[string]string
	Created     int64 `cli:"Created"`
	Modified    int64
	Active      bool `cli:"Active?"`
}

func (uli UserLinkInstance) MarshalJSON() ([]byte, error) {
	var created, modified string

	if uli.Created > 0 {
		created = createdFormatter.FormatMillis(uli.Created)
	}

	if uli.Modified > 0 {
		modified = createdFormatter.FormatMillis(uli.Modified)
	}

	return json.Marshal(&struct {
		Id          *int64            `json:"id,omitempty"`
		Oem         string            `json:"oem"`
		Handle      string            `json:"handle"`
		Version     string            `json:"version"`
		Name        string            `json:"name,omitempty"`
		Description string            `json:"description,omitempty"`
		Visibility  string            `json:"visibility"`
		Properties  map[string]string `json:"props,omitempty"`
		Created     string            `json:"created,omitempty"`
		Modified    string            `json:"modified,omitempty"`
		Flags       *int              `json:"flags,omitempty"`
	}{
		Id:          uli.Id,
		Oem:         uli.Oem,
		Handle:      uli.Handle,
		Version:     uli.Version,
		Name:        uli.Name,
		Description: uli.Description,
		Visibility:  uli.Visibility,
		Properties:  uli.Properties,
		Created:     created,
		Modified:    modified,
		Flags:       uli.Flags,
	})
}

func (uli UserLinkInstance) MarshalSlice() ([]string, error) {
	var created string

	if uli.Created > 0 {
		created = createdFormatter.FormatMillis(uli.Created)
	} else {
		created = "-"
	}

	return []string{
		cast.ToString(uli.Id),
		uli.Name,
		uli.Fqdn,
		uli.Version,
		created,
		stringz.YesOrNo(uli.Active),
	}, nil
}

func ToUserLinkInstance(dli *broker.DataLinkInstance) *UserLinkInstance {
	var result = &UserLinkInstance{
		Id:          dli.Id,
		Name:        dli.Name,
		Description: dli.Description,
		Oem:         dli.DataLink.Oem,
		Handle:      dli.DataLink.Handle,
		Sequence:    dli.DataLink.Seq,
		Fqdn:        dli.DataLink.GetFqdn(),
		Visibility:  dli.Visibility,
		Properties:  dli.Properties,
		Created:     dli.Created,
		Modified:    dli.Modified,
		Flags:       dli.Flags,
		Active:      dli.IsActive(),
	}

	if dli.DataLink.IsReleased() {
		// omit -rc- on released data links
		result.Version = dli.DataLink.Version
	} else {
		result.Version = dli.DataLink.GetVersion()
	}

	return result
}

type userLinkInstancesProvider struct {
	task.Plan
	filter                      string
	params                      *broker.DataInstanceListParams
	dataInstanceListTaskFactory DataInstanceListTaskFactory
}

func (udp *userLinkInstancesProvider) Get() ([]UserLinkInstance, task.Error) {
	var instances []broker.DataLinkInstance
	var workers []task.Worker
	var failure interface{}

	udp.OnReturn = func(i interface{}) { instances = i.([]broker.DataLinkInstance) }
	udp.OnFailure = func(i interface{}) { failure = i }
	workers = append(workers, task.NewWorker(udp.params, udp.dataInstanceListTaskFactory()))
	udp.Sequence(workers...)

	if failure == nil {
		var result = make([]UserLinkInstance, 0)

		for _, ds := range instances {
			var instance = ToUserLinkInstance(&ds)

			result = append(result, *instance)
		}

		if len(result) > 1 {
			slices.SortFunc(result, func(a, b UserLinkInstance) int {
				return cmp.Compare(b.Created, a.Created)
			})
		}

		return result, nil
	}

	return nil, task.NewFailure(failure)
}
