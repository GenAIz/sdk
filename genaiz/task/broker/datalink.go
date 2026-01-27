package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

var (
	errDataLinkConflict = errors.New("data link can not be updated to existing version")
	errDataLinkExists   = errors.New("data link exists")
	errDataLinkInvalid  = errors.New("data link fqdn is invalid")
	errDataLinkNotFound = errors.New("data link not found")
)

type DataLinkWriter interface {
	BuildDataLinks() (string, []DataLink)

	GetDataLink(string, string, string) *DataLink

	GetLatest(string, string) *DataLink

	SyncDataLinks() []*DataLink

	WithDataLink(*DataLink) DataLinkWriter

	Write(string) error
}

type DataLinkParams struct {
	Broker
	shared.ConfigParams
	*DataLink
	NewVersion   string
	NoValidation bool
}

func (dlp DataLinkParams) ToFqdn() (string, string, string) {
	if dlp.DataLink != nil {
		return dlp.Oem, dlp.Handle, dlp.Version
	}

	return "", "", ""
}

func (dlp DataLinkParams) ToString() string {
	if dlp.DataLink != nil {
		return fmt.Sprintf("%s/%s:%s", dlp.Oem, dlp.Handle, dlp.Version)
	}

	return ""
}

func (dlp DataLinkParams) ToPublished() string {
	var oem, handle, ver = dlp.publishedFqdn()

	return fmt.Sprintf("%s/%s:%s", oem, handle, ver)
}

func (dlp DataLinkParams) findDataLink(writer DataLinkWriter) (*DataLink, error) {
	if dlp.DataLink != nil {
		var result *DataLink

		if dlp.Oem != "" && dlp.Handle != "" {
			if dlp.Version == "" {
				result = writer.GetLatest(dlp.Oem, dlp.Handle)
			} else {
				result = writer.GetDataLink(dlp.Oem, dlp.Handle, dlp.Version)
			}

			if result == nil {
				return nil, errDataLinkNotFound
			}

			return result, nil
		}
	}

	return nil, errDataLinkInvalid
}

func (dlp DataLinkParams) exists(writer DataLinkWriter) bool {
	if dlp.DataLink != nil {
		var result = writer.GetDataLink(dlp.Oem, dlp.Handle, dlp.Version)

		return result != nil
	}

	return false
}

func (dlp DataLinkParams) isEqual(link *DataLink) bool {
	if link == nil {
		return dlp.DataLink == nil
	} else if dlp.DataLink != nil {
		return link.IsEqual(dlp.Oem, dlp.Handle, dlp.Version)
	}

	return false
}

func (dlp DataLinkParams) isValid() bool {
	if dlp.DataLink != nil {
		return dlp.Oem != "" && dlp.Handle != "" && dlp.Version != ""
	}

	return false
}

func (dlp DataLinkParams) publishedFqdn() (string, string, string) {
	if dlp.DataLink != nil {
		var resultVersion = dlp.NewVersion

		if resultVersion == "" {
			resultVersion = dlp.Version
		}

		return dlp.Oem, dlp.Handle, resultVersion
	}

	return "", "", ""
}

func NewDataLinkCreateTask(writer DataLinkWriter) *task.Task[DataLinkParams] {
	return &task.Task[DataLinkParams]{
		Name:         "data-link-create",
		OnPrepare:    lang.Assists(writer, handleDataLinkCreateContext),
		OnComplete:   lang.Assists(writer, handleDataLinkCreateComplete),
		OnIncomplete: handleDataLinkCreateIncomplete,
		OnPretend:    lang.Assists(writer, handleDataLinkCreatePretend),
	}
}

func NewDataLinkEditTask(writer DataLinkWriter) *task.Task[DataLinkParams] {
	return &task.Task[DataLinkParams]{
		Name:       "data-link-edit",
		OnPrepare:  lang.Assists(writer, handleDataLinkEditContext),
		OnComplete: lang.Assists(writer, handleDataLinkEditComplete),
		OnPretend:  lang.Assists(writer, handleDataLinkEditPretend),
	}
}

func NewDataLinkExportTask(writer DataLinkWriter) *task.Task[DataLinkParams] {
	return &task.Task[DataLinkParams]{
		Name:         "data-link-export",
		OnPrepare:    handleDataLinkExportContext,
		OnComplete:   lang.Assists(writer, handleDataLinkExportComplete),
		OnIncomplete: handleDataLinkCreateIncomplete,
		OnPretend:    handleDataLinkExportPretend,
	}
}

func NewDataLinkFindTask() *task.Task[DataLinkParams] {
	return &task.Task[DataLinkParams]{
		Name:       "data-link-find",
		OnPrepare:  handleDataLinkFindContext,
		OnComplete: handleDataLinkFindComplete,
		OnPretend:  handleDataLinkFindPretend,
	}
}

func NewDataLinkPublishTask(writer DataLinkWriter) *task.Task[DataLinkParams] {
	return &task.Task[DataLinkParams]{
		Name:       "data-link-publish",
		OnPrepare:  lang.Assists(writer, handleDataLinkPublishContext),
		OnComplete: lang.Assists(writer, handleDataLinkPublishComplete),
		OnPretend:  lang.Assists(writer, handleDataLinkPublishPretend),
	}
}

func handleDataLinkAvailableError(params *DataLinkParams, state *task.State) error {
	if state.Internal != nil {
		var oem, handle, _ = params.ToFqdn()
		var errorText = fmt.Sprintf("Data Link for [%s/%s] is unavailable", oem, handle)

		if links, ok := state.Internal.([]DataLink); ok && len(links) > 0 {
			errorText += ", the following are available:"

			for _, link := range links {
				errorText += fmt.Sprintf("\n%s/%s:%s", link.Oem, link.Handle, link.Version)
			}
		} else {
			state.Logger.Errorf("No data links for [%s/%s] could be retrieved from the broker", oem, handle)
		}

		return errors.New(errorText)
	}

	return nil
}

func handleDataLinkCreateContext(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output == "" {
		var oem, handle, _ = params.ToFqdn()

		state.Logger.Debugf("Looking for local data links for [%s/%s]", oem, handle)

		if !params.exists(writer) {
			var err error

			if state.Output, err = params.ResolveConfigPath(); err == nil {
				state.Logger.Debugf("Validating configuration file [%s]", params.GetConfigFile())

				if _, err = os.Stat(state.Output); err != nil {
					return err
				}
			} else if errors.Is(err, shared.ErrorConfigFileExists) {
				state.Logger.Debugf("Found configuration file [%s]", state.Output)
				return nil
			}

			return err
		}

		return errDataLinkExists
	}

	return nil
}

func handleDataLinkCreateComplete(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output != "" {
		var removed = writer.WithDataLink(params.DataLink).SyncDataLinks()

		for _, link := range removed {
			if link != nil {
				state.Logger.Debugf("Pruned old link [%s,%s:%s]", link.Oem, link.Handle, link.Version)
			}
		}

		state.Logger.Debugf("Updating data links under[%s]", state.Output)

		if err := writer.Write(state.Output); err != nil {
			return err
		}

		state.Reportf("Added link definition [%s]", params.ToString())
		return nil
	}

	return shared.ErrorConfigFileInvalid
}

func handleDataLinkCreateIncomplete(params *DataLinkParams, state *task.State) error {
	_ = params

	if state.Output != "" &&
		errorz.IsPathError(state.Error) {
		var dir = filepath.Dir(state.Output)
		var file = filepath.Base(state.Output)

		state.Logger.Debugf("Creating path [%s], file [%s]", dir, file)

		if fd, err := filez.CreateRecursive(dir, file); err == nil {
			filez.CloseSilently(fd)
			return nil
		} else {
			state.Error = err
		}
	}

	state.Completed = true
	return state.Error
}

func handleDataLinkCreatePretend(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output != "" &&
		!errors.Is(state.Error, errDataLinkExists) {
		var pretender = shared.NewConfigPretender(state.Output)
		var removed = writer.WithDataLink(params.DataLink).SyncDataLinks()
		var rootKey, dataLinks = writer.BuildDataLinks()
		var linksSize = len(dataLinks)

		for index, link := range removed {
			if link != nil {
				linksSize -= 1
				shared.PretendDelete(pretender, func() string {
					return fmt.Sprintf("%s[%d]", rootKey, index)
				})
			}
		}

		if linksSize > 0 {
			var index = linksSize - 1

			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].handle", rootKey, index), dataLinks[index].Handle
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].oem", rootKey, index), dataLinks[index].Oem
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].version", rootKey, index), dataLinks[index].Version
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].description", rootKey, index), dataLinks[index].Description
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].name", rootKey, index), dataLinks[index].Name
			})
		}

		return nil
	}

	return state.Error
}

func handleDataLinkEditContext(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output == "" {
		var oem, handle, _ = params.ToFqdn()

		state.Logger.Debugf("Looking for local data links for [%s/%s]", oem, handle)

		if params.exists(writer) {
			var err error

			if state.Output, err = params.ResolveConfigPath(); err == nil ||
				errors.Is(err, shared.ErrorConfigFileExists) {
				state.Logger.Debugf("Validating configuration file [%s]", params.GetConfigFile())
				return nil
			}

			return err
		}

		return errDataLinkNotFound
	}

	return nil
}

func handleDataLinkEditComplete(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output != "" {
		var oem, handle, ver = params.ToFqdn()
		var removed = writer.WithDataLink(params.DataLink).SyncDataLinks()

		for _, link := range removed {
			if link != nil && link.Version != params.Version {
				state.Logger.Debugf("Pruning old link [%s,%s:%s]", link.Oem, link.Handle, link.Version)
			}
		}

		state.Logger.Debugf("Updating data link [%s,%s:%s] under[%s]", oem, handle, ver, state.Output)

		if err := writer.Write(state.Output); err != nil {
			return err
		}

		state.Reportf("Edited link definition [%s]", params.ToString())
		return nil

	}

	return shared.ErrorConfigFileInvalid
}

func handleDataLinkEditPretend(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output != "" &&
		!errors.Is(state.Error, errDataLinkNotFound) {
		var pretender = shared.NewConfigPretender(state.Output)
		var removed = writer.WithDataLink(params.DataLink).SyncDataLinks()
		var rootKey, dataLinks = writer.BuildDataLinks()
		var linksSize = len(dataLinks)

		for index, link := range removed {
			if link != nil {
				linksSize -= 1
				shared.PretendDelete(pretender, func() string {
					return fmt.Sprintf("%s[%d]", rootKey, index)
				})
			}
		}

		if i := slices.IndexFunc(dataLinks, func(link DataLink) bool {
			return params.isEqual(&link)
		}); i >= 0 {
			var index = linksSize - 1

			shared.PretendDelete(pretender, func() string {
				return fmt.Sprintf("%s[%d].PropSpecs", rootKey, index)
			})
			pretendPropSpec(pretender, rootKey, index, dataLinks[index].PropSpecs)
			shared.PretendDelete(pretender, func() string {
				return fmt.Sprintf("%s[%d].SecretSpecs", rootKey, i)
			})
			pretendPropSpec(pretender, rootKey, index, dataLinks[index].SecretSpecs)

			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].handle", rootKey, index), dataLinks[index].Handle
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].oem", rootKey, index), dataLinks[index].Oem
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].version", rootKey, index), dataLinks[index].Version
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].description", rootKey, index), dataLinks[index].Description
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].name", rootKey, index), dataLinks[index].Name
			})
		}

		return nil
	}

	return state.Error
}

func handleDataLinkExportContext(params *DataLinkParams, state *task.State) error {
	if state.Output == "" {
		if params.isValid() {
			var err error

			if state.Output, err = params.ResolveOptionalType(shared.ConfigTypeYaml); err == nil {
				state.Logger.Debugf("Validating configuration file [%s]", state.Output)

				if _, err = os.Stat(state.Output); err != nil {
					return err
				}
			} else if errors.Is(err, shared.ErrorConfigFileExists) {
				state.Logger.Debugf("Found configuration file [%s]", state.Output)
				return nil
			}

			return err
		}

		return errDataLinkInvalid
	}

	return nil
}

func handleDataLinkExportComplete(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var sequence string
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var oem, handle, ver = params.ToFqdn()
			var fqdnString = params.ToString()
			var remote *DataLink

			state.Logger.Debugf("Looking up for data link [%s]", fqdnString)

			if params.Seq != 0 {
				sequence = cast.ToString(params.Seq)
				state.Logger.Debugf("Clamping on sequence version [%s]", sequence)
			}

			if remote, err = brokerClient.ExportDataLink(oem, handle, ver, sequence); err == nil {
				writer.WithDataLink(remote).SyncDataLinks()

				if err = writer.Write(state.Output); err == nil {
					if sequence == "" {
						state.Reportf("Imported data link [%s]", fqdnString)
					} else {
						state.Reportf("Imported data link [%s-rc%s]", fqdnString, sequence)
					}

					// Here: We'll need to create a shared.PropSpecState to be able to buffer Env vars for other tasks
					return nil
				}
			}
		}

		return err
	}

	return shared.ErrorConfigFileInvalid
}

func handleDataLinkExportPretend(params *DataLinkParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var oem, handle, ver = params.ToFqdn()
			var fqdnString = params.ToString()
			var sequence string

			if params.Seq == 0 {
				state.Logger.Debugf("Pretending to export data link [%s]", fqdnString)
			} else {
				sequence = cast.ToString(params.Seq)
				state.Logger.Debugf("Pretending to export data link [%s-rc%s]", fqdnString, sequence)
			}

			fmt.Printf("curl -X GET -H \"Accept: application/json\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -d oem=%s\\\n", oem)
			fmt.Printf("  -d handle=%s\\\n", handle)
			fmt.Printf("  -d version=%s\\\n", ver)

			if sequence != "" {
				fmt.Printf("  -d sequence=%s\\\n", sequence)
			}

			fmt.Printf("%s\n", brokerClient.ListDataLinksUrl())
			return nil
		}

		return err
	}

	return state.Error
}

func handleDataLinkFindContext(params *DataLinkParams, state *task.State) error {
	if state.Internal == nil {
		if !params.isValid() {
			return errDataLinkInvalid
		}
	}

	return nil
}

func handleDataLinkFindComplete(params *DataLinkParams, state *task.State) error {
	var oem, handle, ver = params.ToFqdn()
	var brokerClient Client
	var err error

	if params.NoValidation {
		state.Logger.Debugf("Skipping data link validation for [%s/%s:%s]", oem, handle, ver)
		state.Internal = params.DataLink
	} else if brokerClient, err = params.GetClient(); err == nil {
		var dataLinks []DataLink

		state.Logger.Debugf("Finding a data link corresponding to [%s/%s]", oem, handle)

		if dataLinks, err = brokerClient.ListDataLinks(oem, handle, DataLinkFlags.Active); err == nil {
			if i := slices.IndexFunc(dataLinks, func(link DataLink) bool {
				return link.IsActive() && params.isEqual(&link)
			}); i >= 0 {
				state.Logger.Debugf("Found data link [%d]", dataLinks[i].Id)
				state.Internal = dataLinks[i]
				return nil
			}

			state.Logger.Debugf("Could not find a data link for [%s/%s]", oem, handle)
			state.Internal = dataLinks
			return handleDataLinkAvailableError(params, state)
		}
	}

	return err
}

func handleDataLinkFindPretend(params *DataLinkParams, state *task.State) error {
	var oem, handle, _ = params.ToFqdn()
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		state.Logger.Debugf("Pretending to list data links for [%s/%s]", oem, handle)
		fmt.Printf("curl -X GET -H \"Accept: application/json\" \\\n")
		fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
		fmt.Printf("  -d oem=%s\\\n", oem)
		fmt.Printf("  -d handle=%s\\\n", handle)
		fmt.Printf("  -d flags=%d\\\n", DataLinkFlags.Active)
		fmt.Printf("%s\n", brokerClient.ListDataLinksUrl())
		return nil
	}

	return err
}

func handleDataLinkPublishContext(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output == "" {
		var fqdn = params.ToString()
		var err error

		state.Logger.Debugf("Finding local link definition for [%s]", fqdn)

		if _, err = params.findDataLink(writer); err == nil {
			var pubOem, pubHandle, pubVersion = params.publishedFqdn()
			var brokerClient Client

			state.Logger.Debugf("Finding remote link definitions for [%s/%s:%s]", pubOem, pubHandle, pubVersion)

			if brokerClient, err = params.GetClient(); err == nil {
				var remoteLink *DataLink

				if remoteLink, err = brokerClient.FindDataLink(pubOem, pubHandle, pubVersion); err == nil {
					state.Output = cast.ToString(remoteLink.Id)
					return errDataLinkConflict
				}

				return nil
			}
		}

		return err
	}

	return nil
}

func handleDataLinkPublishComplete(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output == "" {
		var publishing *DataLink
		var brokerClient Client
		var err error

		if publishing, err = params.findDataLink(writer); err == nil {
			if brokerClient, err = params.GetClient(); err == nil {
				var sanitized = publishing.Sanitize()
				var result *DataLink

				if params.NewVersion != "" {
					sanitized.Version = params.NewVersion
				}

				state.Logger.Debugf("Publishing data link [%s] to broker [%s]",
					params.ToPublished(), brokerClient.GetHostAddr())

				if result, err = brokerClient.PublishDataLink(sanitized); err == nil {
					state.Reportf("Published data link [%s/%s:%s], id [%d]",
						result.Oem, result.Handle, result.Version, result.Id)
					state.Internal = result
					return nil
				}
			}
		}

		return err
	}

	return errDataLinkExists
}

func handleDataLinkPublishPretend(writer DataLinkWriter, params *DataLinkParams, state *task.State) error {
	if state.Output == "" {
		var localLink *DataLink
		var brokerClient Client
		var err error

		if localLink, err = params.findDataLink(writer); err == nil {
			if brokerClient, err = params.GetClient(); err == nil {
				var linkModel []byte

				if params.NewVersion != "" {
					localLink.Version = params.NewVersion
				}

				linkModel, _ = json.Marshal(localLink)
				state.Logger.Debugf("Pretending to publish data link [%s]", params.ToString())
				fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
				fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
				fmt.Printf("  -d model=\"%s\"\\\n", url.QueryEscape(string(linkModel)))
				fmt.Printf("%s\n", brokerClient.ListDataLinksUrl())
				return nil
			}
		}

		return err
	}

	return state.Error
}

func pretendPropSpec(pretender shared.ConfigPretender, rootKey string, index int, specs []PropSpec) {
	for i, spec := range specs {
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].PropSpecs[%d].key", rootKey, index, i), spec.Key
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].PropSpecs[%d].type", rootKey, index, i), spec.Type
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].PropSpecs[%d].name", rootKey, index, i), spec.Name
		})

		if spec.Description != "" {
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].PropSpecs[%d].description", rootKey, index, i), spec.Description
			})
		}

		if spec.Value != "" {
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].PropSpecs[%d].value", rootKey, index, i), spec.Value
			})
		} else {
			shared.PretendSlice(pretender, func() (string, []string) {
				return fmt.Sprintf("%s[%d].PropSpecs[%d].values", rootKey, index, i), spec.Values
			})
		}
	}
}
