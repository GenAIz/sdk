package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

type DataLinkParams struct {
	Broker
	*shared.ConfigParams

	Description  string
	Handle       string
	Name         string
	NoValidation bool
	Oem          string
	Version      string
}

func (dlp DataLinkParams) ToDataLink() *DataLink {
	return &DataLink{
		Description: dlp.Description,
		Handle:      dlp.Handle,
		Name:        dlp.Name,
		Oem:         dlp.Oem,
		Version:     dlp.Version,
	}
}

func (dlp DataLinkParams) ToString() string {
	return fmt.Sprintf("%s/%s:%s", dlp.Oem, dlp.Handle, dlp.Version)
}

func (dlp DataLinkParams) isEqual(link *DataLink) bool {
	return strings.EqualFold(dlp.Oem, link.Oem) &&
		strings.EqualFold(dlp.Handle, link.Handle) &&
		strings.EqualFold(dlp.Version, link.Version)
}

func (dlp DataLinkParams) isValid() bool {
	return dlp.Oem != "" && dlp.Handle != "" && dlp.Version != ""
}

func NewDataLinkFindTask() *task.Task[DataLinkParams] {
	return &task.Task[DataLinkParams]{
		Name:       "data-link-find",
		OnPrepare:  handleDataLinkFindContext,
		OnComplete: handleDataLinkFindComplete,
		OnPretend:  handleDataLinkFindPretend,
	}
}

func NewDataLinkPublishTask() *task.Task[DataLinkParams] {
	return &task.Task[DataLinkParams]{
		Name:       "data-link-publish",
		OnPrepare:  handleDataLinkPublishContext,
		OnComplete: handleDataLinkPublishComplete,
		OnPretend:  handleDataLinkPublishPretend,
	}
}

func handleDataLinkAvailableError(params *DataLinkParams, state *task.State) error {
	if state.Internal != nil {
		var errorText = fmt.Sprintf("Data Link for [%s/%s] is unavailable", params.Oem, params.Handle)

		if links, ok := state.Internal.([]DataLink); ok && len(links) > 0 {
			errorText += ", the following are available:"

			for _, link := range links {
				errorText += fmt.Sprintf("\n%s/%s:%s", link.Oem, link.Handle, link.Version)
			}
		} else {
			state.Logger.Errorf("No data links for [%s/%s] could be retrieved from the broker", params.Oem, params.Handle)
		}

		return errors.New(errorText)
	}

	return nil
}

func handleDataLinkFindContext(params *DataLinkParams, state *task.State) error {
	if state.Internal == nil {
		if !params.isValid() {
			return errors.New("not a valid data link identity")
		}
	}

	return nil
}

func handleDataLinkFindComplete(params *DataLinkParams, state *task.State) error {
	var brokerClient Client
	var err error

	if params.NoValidation {
		state.Logger.Debugf("Skipping data link validation for [%s/%s:%s]", params.Oem, params.Handle, params.Version)
		state.Internal = DataLink{
			Oem:     params.Oem,
			Handle:  params.Handle,
			Version: params.Version,
		}
	} else if brokerClient, err = params.GetClient(); err == nil {
		var dataLinks []DataLink

		state.Logger.Debugf("Finding a data link corresponding to [%s/%s]", params.Oem, params.Handle)

		if dataLinks, err = brokerClient.ListDataLinks(params.Oem, params.Handle, DataLinkFlags.Active); err == nil {
			if i := slices.IndexFunc(dataLinks, func(link DataLink) bool {
				return link.IsActive() && params.isEqual(&link)
			}); i >= 0 {
				state.Logger.Debugf("Found data link [%d]", dataLinks[i].Id)
				state.Internal = dataLinks[i]
				return nil
			}

			state.Logger.Debugf("Could not find a data link for [%s/%s]", params.Oem, params.Handle)
			state.Internal = dataLinks
			return handleDataLinkAvailableError(params, state)
		}
	}

	return err
}

func handleDataLinkFindPretend(params *DataLinkParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		state.Logger.Debugf("Pretending to list data links for [%s/%s]", params.Oem, params.Handle)
		fmt.Printf("curl -X GET -H \"Accept: application/json\" \\\n")
		fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
		fmt.Printf("  -d oem=%s\\\n", params.Oem)
		fmt.Printf("  -d handle=%s\\\n", params.Handle)
		fmt.Printf("  -d flags=%d\\\n", DataLinkFlags.Active)
		fmt.Printf("%s\n", brokerClient.ListDataLinksUrl())
		return nil
	}

	return err
}

func handleDataLinkPublishContext(params *DataLinkParams, state *task.State) error {
	if state.Internal == nil {
		var err error

		if err = handleDataLinkFindContext(params, state); err == nil {
			if params.Name == "" {
				err = errors.New("invalid data link name")
			}
		}

		return err
	}

	return nil
}

func handleDataLinkPublishComplete(params *DataLinkParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var result *DataLink

		state.Logger.Debugf("Publishing data link [%s/%s:%s] to broker [%s]",
			params.Oem, params.Handle, params.Version, brokerClient.GetHostAddr())

		if result, err = brokerClient.PublishDataLink(params.ToDataLink()); err == nil {
			state.Reportf("Published data link [%s/%s:%s], id [%d]",
				result.Oem, result.Handle, result.Version, result.Id)
			state.Internal = result
			return nil
		}
	}

	return err
}

func handleDataLinkPublishPretend(params *DataLinkParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var linkModel, _ = json.Marshal(params.ToDataLink())

		state.Logger.Debugf("Pretending to list data links for [%s/%s]", params.Oem, params.Handle)
		fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
		fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
		fmt.Printf("  -d model=\"%s\"\\\n", url.QueryEscape(string(linkModel)))
		fmt.Printf("%s\n", brokerClient.ListDataLinksUrl())
		return nil
	}

	return err
}
