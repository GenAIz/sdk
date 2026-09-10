package locker

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorLockerDataLinkEmpty      = task.NewError("datalink property set is empty, is it incomplete?")
	errorLockerDataLinkConflict   = task.NewError("datalink is used in multiple sources")
	errorLockerDataSourceConflict = task.NewError("datasource name is used in multiple sources")
	errorDataSourceExist          = task.NewError("data source exists")
	errorDataSourceUnknown        = task.NewError("update called with no data source")
	errorDataSourceLinkUnknown    = task.NewError("data source data link is unknown")
)

type SourceAddParams struct {
	BaseParams
	LinkParams
	broker.Broker
	SourceHandle string
}

type SourceFindParams struct {
	BaseParams
	*broker.DataLinkParams
	SourceHandle string
	SourceName   string

	client broker.Client
}

func (sfp SourceFindParams) GetClient() (broker.Client, error) {
	if sfp.client == nil {
		return sfp.Broker.GetClient()
	}

	return sfp.client, nil
}

func (sfp SourceFindParams) hasValidDataLinkParams() bool {
	return sfp.DataLinkParams != nil && sfp.DataLinkParams.IsValid()
}

type SourcePublishParams struct {
	BaseParams
	broker.Broker
	SourceHandle string
	Name         string
	Description  string
	Visibility   string

	client broker.Client
}

func (spp SourcePublishParams) GetClient() (broker.Client, error) {
	if spp.client == nil {
		return spp.Broker.GetClient()
	}

	return spp.client, nil
}

type SourceUpdateParams struct {
	*SourceFindParams
	PropertyParams
}

func NewSourceAddTask() *task.Task[SourceAddParams] {
	return &task.Task[SourceAddParams]{
		Name:       "source-add",
		OnPrepare:  handleSourceAddContext,
		OnComplete: handleSourceAddComplete,
		OnPretend:  handleSourceAddPretend,
	}
}

func NewSourceFindTask() *task.Task[SourceFindParams] {
	return &task.Task[SourceFindParams]{
		Name:         "source-find",
		OnPrepare:    handleSourceFindContext,
		OnComplete:   handleSourceFindComplete,
		OnIncomplete: handleSourceFindIncomplete,
		OnPretend:    handleSourceFindPretend,
	}
}

func NewSourceSyncTask() *task.Task[SourceFindParams] {
	return &task.Task[SourceFindParams]{
		Name:       "source-sync",
		OnPrepare:  handleSourceSyncContext,
		OnComplete: handleSourceSyncComplete,
		OnPretend:  handleSourceSyncPretend,
	}
}

func NewSourcePublishTask() *task.Task[SourcePublishParams] {
	return &task.Task[SourcePublishParams]{
		Name:         "source-publish",
		OnPrepare:    handleSourcePublishContext,
		OnComplete:   handleSourcePublishCreate,
		OnIncomplete: handleSourcePublishUpdate,
		OnPretend:    handleSourcePublishPretend,
	}
}

func NewSourceUpdateTask() *task.Task[SourceUpdateParams] {
	return &task.Task[SourceUpdateParams]{
		Name:       "source-update",
		OnPrepare:  handleSourceUpdateContext,
		OnComplete: handleSourceUpdateComplete,
		OnPretend:  handleSourceUpdatePretend,
	}
}

func handleSourceAddComplete(params *SourceAddParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		state.Logger.Debugf("Adding source [%s] to locker [%s]", params.SourceHandle, params.LockerPath)

		if brokerClient, err = params.GetClient(); err == nil {
			var lockerState = NewSecuredLockerState(state)
			var accountUrl = brokerClient.GetHostAddr()

			state.Logger.Debugf("Adding data source to account [%s]", accountUrl)

			if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
				var link = &lockerLink{
					LockerHandle: params.SourceHandle,
					LinkOem:      params.Oem,
					LinkHandle:   params.Handle,
					LinkVersion:  params.Version,
				}

				if err = lockerState.addSource(accountUrl, link); err == nil {
					if err = lockerState.Write(params.LockerPath, params.Passphrase); err == nil {
						state.Reportf("Added data source %s to locker %s", params.SourceHandle, params.LockerPath)
						state.Output = ""
						return nil
					}
				}
			}

			if strings.HasPrefix(err.Error(), "chacha20poly1305") {
				return errorLockerPassFailed
			}
		}

		return err
	}

	return errorLockerPathInvalid
}

func handleSourceAddContext(params *SourceAddParams, state *task.State) error {
	return handleSourceBaseContext(&params.BaseParams, state)
}

func handleSourceAddPretend(params *SourceAddParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var accountUrl = brokerClient.GetHostAddr()

			state.Logger.Debugf("Pretending to add data source [%s] to account [%s]", params.SourceHandle, accountUrl)
			state.Logger.Debugf("Data source added for link [%s/%s:%s]", params.Oem, params.Handle, params.Version)
			state.Output = ""
			return nil
		}

		return err
	}

	return errorLockerPathInvalid
}

func handleSourceBaseContext(params *BaseParams, state *task.State) error {
	if state.Output == "" {
		var varSpecState = shared.NewVarSpecState(state)
		var err error

		if len(varSpecState.VarSpecs) == 0 {
			// We expect to have all the data on the dataLink we need collected or exported and to
			// have specs available for the add to be valid
			return errorLockerDataLinkEmpty
		}

		if err = filez.IsReadable(params.LockerPath); errorz.IsPathError(err) {
			return fmt.Errorf("locker [%s] can not be read", params.LockerPath)
		}

		if params.Passphrase == nil {
			return errorLockerPassFailed
		}

		state.Output = params.LockerPath
	}

	return nil
}

func handleSourceFindComplete(params *SourceFindParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var lockerState = NewSecuredLockerState(state)

			state.Logger.Debugf("Finding source from locker [%s]", params.LockerPath)

			if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
				var link RemoteLink

				if link, err = lockerState.LookupSource(brokerClient.GetHostAddr(), state.Output); err == nil {
					var oem, handle, ver = link.GetPublishing()

					params.DataLinkParams.DataLink = &broker.DataLink{
						Oem:     oem,
						Handle:  handle,
						Version: ver,
					}
					state.Reportf("Found datalink [%s]", params.ToPublished())
					state.Output = ""
					return nil
				}
			}

			if strings.HasPrefix(err.Error(), "chacha20poly1305") {
				return errorLockerPassFailed
			}
		}

		return err
	}

	return errorLockerDataLinkInvalid
}

func handleSourceFindContext(params *SourceFindParams, state *task.State) error {
	if state.Output == "" {
		state.Logger.Debugf("Need datalink fqdn for handle [%s]", params.SourceHandle)

		if params.hasValidDataLinkParams() {
			state.Output = params.ToPublished()
			return errorLockerDataLinkFound
		}

		if params.Passphrase == nil {
			return errorLockerPassFailed
		}

		state.Output = params.SourceHandle
	}

	return nil
}

func handleSourceFindIncomplete(params *SourceFindParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, errorLockerDataLinkFound) {
		state.Reportf("Already know datalink [%s]", params.ToPublished())
		return nil
	}

	return state.Error
}

func handleSourceFindPretend(params *SourceFindParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var lockerState = NewSecuredLockerState(state)

			state.Logger.Debugf("Finding source from locker [%s]", params.LockerPath)

			if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
				state.Logger.Debugf("Pretending to lookup a source [%s] under account [%s]",
					params.SourceHandle, brokerClient.GetHostAddr())
				state.Output = ""
				return nil
			}

			if strings.HasPrefix(err.Error(), "chacha20poly1305") {
				return errorLockerPassFailed
			}
		}

		return err
	}

	return errorLockerDataLinkInvalid
}

func handleSourcePublishContext(params *SourcePublishParams, state *task.State) error {
	if state.Output == "" {
		var ok bool

		if params.Passphrase == nil {
			return errorLockerPassFailed
		}

		state.Logger.Debugf("Publishing [%s] from locker [%s]", params.SourceHandle, params.LockerPath)

		if state.Internal != nil {
			if _, ok = state.Internal.(broker.DataLinkInstance); ok {
				return errorDataSourceExist
			} else if _, ok = state.Internal.(broker.DataLink); ok {
				return nil
			}
		}

		return errorDataSourceLinkUnknown
	}

	return nil
}

func handleSourcePublishCreate(params *SourcePublishParams, state *task.State) error {
	if state.Internal != nil {
		var dataLink = state.Internal.(broker.DataLink)
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var lockerState = NewSecuredLockerState(state)
			var account = brokerClient.GetHostAddr()

			state.Logger.Debugf("Creating data source for data link [%d] on account [%s]", *dataLink.Id, account)

			if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
				var props map[string]string

				defer lockerState.Destroy()
				state.Logger.Debugf("Pushing data source handle [%s] onto name [%s]", params.SourceHandle, params.Name)

				if props, err = lockerState.GetSourceProps(account, params.SourceHandle, params.Passphrase); err == nil {
					var updated *broker.DataLinkInstance
					var instance = &broker.DataLinkInstance{
						Name:        params.Name,
						Description: params.Description,
						Visibility:  params.Visibility,
						DataLinkId:  dataLink.Id,
					}

					if updated, err = brokerClient.CreateDataSource(instance, props); err == nil {
						state.Reportf("Created data source [%s], id: [%d]", params.SourceHandle, *updated.Id)
						return nil
					}
				}
			}
		}

		return err
	}

	return errorDataSourceUnknown
}

func handleSourcePublishUpdate(params *SourcePublishParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, errorDataSourceExist) {
		if state.Internal != nil {
			var instance = state.Internal.(broker.DataLinkInstance)
			var brokerClient broker.Client
			var err error

			if brokerClient, err = params.GetClient(); err == nil {
				var lockerState = NewSecuredLockerState(state)
				var account = brokerClient.GetHostAddr()

				state.Logger.Debugf("Updating data source [%d] from locker [%s]", *instance.Id, params.LockerPath)

				if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
					var props map[string]string

					defer lockerState.Destroy()
					state.Logger.Debugf("Pushing data source handle [%s] onto name [%s]", params.SourceHandle, instance.Name)

					if props, err = lockerState.GetSourceProps(account, params.SourceHandle, params.Passphrase); err == nil {
						var updated *broker.DataLinkInstance

						if updated, err = brokerClient.UpdateDataSource(&instance, props); err == nil {
							state.Reportf("Updated data source [%s], id: [%d]", params.SourceHandle, *updated.Id)
							return nil
						}
					}
				}
			}

			return err
		}

		return errorDataSourceUnknown
	}

	return state.Error
}

func handleSourcePublishPretend(params *SourcePublishParams, state *task.State) error {
	if errors.Is(state.Error, errorDataSourceExist) {
		var instance = state.Internal.(broker.DataLinkInstance)
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending to update a data source with name [%s]", instance.Name)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d id=\"%d\"\\\n", *instance.Id)
			fmt.Printf("  -d name=\"%s\"\\\n", url.QueryEscape(params.Name))
			fmt.Printf("  -d description=\"%s\"\\\n", url.QueryEscape(params.Description))
			fmt.Printf("  -d dataLinkId=\"%d\"\\\n", *instance.DataLinkId)
			fmt.Printf("  -d visibility=\"%s\"\\\n", params.Visibility)
			fmt.Printf("  -d active=\"%s\"\\\n", cast.ToString(true))
			fmt.Println("  -d props=\"{...}\"")
			fmt.Printf("%s\n", brokerClient.UpdateDataSourceUrl())
			return nil
		}

		return err
	} else if state.Error == nil {
		var link = state.Internal.(broker.DataLink)
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending to create a data source with name [%s]", params.Name)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d name=\"%s\"\\\n", url.QueryEscape(params.Name))
			fmt.Printf("  -d description=\"%s\"\\\n", url.QueryEscape(params.Description))
			fmt.Printf("  -d dataLinkId=\"%d\"\\\n", *link.Id)
			fmt.Printf("  -d visibility=\"%s\"\\\n", params.Visibility)
			fmt.Printf("  -d active=\"%s\"\\\n", cast.ToString(true))
			fmt.Println("  -d props=\"{...}\"")
			fmt.Printf("%s\n", brokerClient.CreateDataSourceUrl())
			return nil
		}

		return err
	}

	return state.Error
}

func handleSourceSyncComplete(params *SourceFindParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		state.Logger.Debugf("Retrieving data source [%s] for data link [%s]", params.SourceHandle, state.Output)

		if brokerClient, err = params.GetClient(); err == nil {
			var sources []broker.DataLinkInstance

			if sources, err = brokerClient.ListDataSources(); err == nil {
				state.Output = ""

				if len(sources) > 0 {
					var filtered = slices.DeleteFunc(sources, func(instance broker.DataLinkInstance) bool {
						return !instance.HasLink(params.DataLink)
					})

					if params.SourceName != "" {
						var named = slices.DeleteFunc(filtered, func(instance broker.DataLinkInstance) bool {
							return !strings.EqualFold(params.SourceName, instance.Name)
						})

						if len(named) > 1 {
							return errorLockerDataSourceConflict
						} else if len(named) == 1 {
							state.Internal = named[0]
						}

						return nil
					} else if len(filtered) == 1 {
						state.Internal = filtered[0]
						return nil
					}

					return errorLockerDataLinkConflict
				}

				return nil
			}
		}

		return err
	}

	return errorLockerDataLinkInvalid
}

func handleSourceSyncContext(params *SourceFindParams, state *task.State) error {
	if state.Output == "" {
		if !params.hasValidDataLinkParams() {
			return errorLockerDataLinkInvalid
		}

		state.Output = params.DataLinkParams.ToPublished()
	}

	return nil
}

func handleSourceSyncPretend(params *SourceFindParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending finding data source [%s] for account [%s]",
				params.SourceHandle, brokerClient.GetHostAddr())

			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("%s\n", brokerClient.ListDataSourcesUrl())
			state.Logger.Debugf("Filtering data sources for data link [%s]", params.DataLinkParams.ToPublished())
			return nil
		}

		return err
	}

	return errorLockerDataLinkInvalid
}

func handleSourceUpdateComplete(params *SourceUpdateParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		state.Logger.Debugf("Updating source [%s] from locker [%s]", params.SourceHandle, params.LockerPath)

		if brokerClient, err = params.GetClient(); err == nil {
			var lockerState = NewSecuredLockerState(state)

			if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
				var accountUrl = brokerClient.GetHostAddr()
				var valueEnclave Enclave

				if params.Secret != nil {
					valueEnclave = params.Secret
				} else if params.Value != "" {
					valueEnclave = memguard.NewEnclave([]byte(params.Value))
				} else {
					state.Reportf("Property key [%s] set to the empty string", params.Key)
					valueEnclave = newEmptyEnclave()
				}

				if err = lockerState.updateSource(accountUrl, params.SourceHandle, params.Key, valueEnclave, params.Passphrase); err == nil {
					if err = lockerState.Write(params.LockerPath, params.Passphrase); err == nil {
						state.Reportf("Updated property key [%s] for source [%s] on account [%s]",
							params.Key, params.SourceHandle, accountUrl)
						return nil
					}
				}
			}

			if strings.HasPrefix(err.Error(), "chacha20poly1305") {
				return errorLockerPassFailed
			}
		}

		return err
	}

	return errorLockerPathInvalid
}

func handleSourceUpdateContext(params *SourceUpdateParams, state *task.State) error {
	if state.Output == "" {
		var err error

		if err = handleSourceBaseContext(&params.BaseParams, state); err == nil {
			var varSpecState = shared.NewVarSpecState(state)

			if i := slices.IndexFunc(varSpecState.VarSpecs, func(spec shared.VarSpec) bool {
				return strings.EqualFold(spec.GetKey(), params.Key)
			}); i >= 0 {
				var spec = varSpecState.VarSpecs[i]

				if spec.IsSecret() && params.Secret == nil && params.Value != "" {
					return fmt.Errorf("property [%s] is a secret key for datalink [%s]", params.Key, params.ToPublished())

				}

				if err = spec.Validate(params); err != nil {
					return fmt.Errorf("property value type for key [%s] is invalid", params.Key)
				}

				return nil
			}

			return fmt.Errorf("property [%s] is not a valid key for datalink [%s]", params.Key, params.ToPublished())
		}

		return err
	}

	return nil
}

func handleSourceUpdatePretend(params *SourceUpdateParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var accountUrl = brokerClient.GetHostAddr()

			state.Logger.Debugf("Pretending to update data source [%s] from account [%s]", params.SourceHandle, accountUrl)
			state.Logger.Debugf("Data source property [%s] updated for link [%s/%s:%s]", params.Key, params.Oem, params.Handle, params.Version)
			state.Output = ""
			return nil
		}

		return err
	}

	return errorLockerPathInvalid
}
