package locker

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/slicez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorLockerDataStoreConflict = task.NewError("datastore name is used in multiple stores")
	errorDataStoreExist          = task.NewError("data store exists")
	errorDataStoreLinkUnknown    = task.NewError("update called with no data store")
	errorDataStoreUnknown        = task.NewError("update called with no data store")
)

type StoreAddParams struct {
	BaseParams
	LinkParams
	broker.Broker
}

type StoreFindParams struct {
	BaseParams
	*broker.DataLinkParams
	StoreName string

	client broker.Client
}

func (sfp StoreFindParams) GetClient() (broker.Client, error) {
	if sfp.client == nil {
		return sfp.Broker.GetClient()
	}

	return sfp.client, nil
}

type StorePublishParams struct {
	BaseParams
	broker.Broker
	Name        string
	Description string
	Visibility  string

	client broker.Client
}

func (spp StorePublishParams) GetClient() (broker.Client, error) {
	if spp.client == nil {
		return spp.Broker.GetClient()
	}

	return spp.client, nil
}

type StoreUpdateParams struct {
	*StoreFindParams
	PropertyParams
}

func NewStoreAddTask() *task.Task[StoreAddParams] {
	return &task.Task[StoreAddParams]{
		Name:       "store-add",
		OnPrepare:  handleStoreAddContext,
		OnComplete: handleStoreAddComplete,
		OnPretend:  handleStoreAddPretend,
	}
}

func NewStoreFindTask() *task.Task[StoreFindParams] {
	return &task.Task[StoreFindParams]{
		Name:         "store-find",
		OnPrepare:    handleStoreFindContext,
		OnComplete:   handleStoreFindComplete,
		OnIncomplete: handleStoreFindIncomplete,
		OnPretend:    handleStoreFindPretend,
	}
}

func NewStorePublishTask() *task.Task[StorePublishParams] {
	return &task.Task[StorePublishParams]{
		Name:         "store-publish",
		OnPrepare:    handleStorePublishContext,
		OnComplete:   handleStorePublishCreate,
		OnIncomplete: handleStorePublishUpdate,
		OnPretend:    handleStorePublishPretend,
	}
}

func NewStoreSyncTask() *task.Task[StoreFindParams] {
	return &task.Task[StoreFindParams]{
		Name:       "store-sync",
		OnPrepare:  handleStoreSyncContext,
		OnComplete: handleStoreSyncComplete,
		OnPretend:  handleStoreSyncPretend,
	}
}

func NewStoreUpdateTask() *task.Task[StoreUpdateParams] {
	return &task.Task[StoreUpdateParams]{
		Name:       "store-update",
		OnPrepare:  handleStoreUpdateContext,
		OnComplete: handleStoreUpdateComplete,
		OnPretend:  handleStoreUpdatePretend,
	}
}

func handleStateStoreInstanceList(links []broker.DataLinkInstance, state *task.State) error {
	if len(links) > 1 {
		return errorLockerDataStoreConflict
	} else if len(links) == 1 {
		state.Internal = links[0]
	}

	return nil
}

func handleStoreAddComplete(params *StoreAddParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		state.Logger.Debugf("Adding store [%s] to locker [%s]", params.LockerHandle, params.LockerPath)

		if brokerClient, err = params.GetClient(); err == nil {
			var lockerState = NewSecuredLockerState(state)
			var accountUrl = brokerClient.GetHostAddr()

			state.Logger.Debugf("Adding data Store to account [%s]", accountUrl)

			if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
				var link = &lockerLink{
					LockerHandle: params.LockerHandle,
					LinkOem:      params.Oem,
					LinkHandle:   params.Handle,
					LinkVersion:  params.Version,
				}

				if err = lockerState.addStore(accountUrl, link); err == nil {
					if err = lockerState.Write(params.LockerPath, params.Passphrase); err == nil {
						state.Reportf("Added data Store %s to locker %s", params.LockerHandle, params.LockerPath)
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

func handleStoreAddContext(params *StoreAddParams, state *task.State) error {
	return handleBaseAddContext(&params.BaseParams, state)
}

func handleStoreAddPretend(params *StoreAddParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var accountUrl = brokerClient.GetHostAddr()

			state.Logger.Debugf("Pretending to add data store [%s] to account [%s]", params.LockerHandle, accountUrl)
			state.Logger.Debugf("Data store added for link [%s/%s:%s]", params.Oem, params.Handle, params.Version)
			state.Output = ""
			return nil
		}

		return err
	}

	return errorLockerPathInvalid
}

func handleStoreFindComplete(params *StoreFindParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var lockerState = NewSecuredLockerState(state)

			state.Logger.Debugf("Finding store from locker [%s]", params.LockerPath)

			if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
				var link RemoteLink

				if link, err = lockerState.LookupStore(brokerClient.GetHostAddr(), state.Output); err == nil {
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

func handleStoreFindContext(params *StoreFindParams, state *task.State) error {
	return handleBaseFindContext(&params.BaseParams, params.DataLinkParams, state)
}

func handleStoreFindIncomplete(params *StoreFindParams, state *task.State) error {
	return handleBaseFindIncomplete(params.DataLinkParams, state)
}

func handleStoreFindPretend(params *StoreFindParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var lockerState = NewSecuredLockerState(state)

			state.Logger.Debugf("Finding store from locker [%s]", params.LockerPath)

			if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
				state.Logger.Debugf("Pretending to lookup a store [%s] under account [%s]",
					params.LockerHandle, brokerClient.GetHostAddr())
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

func handleStorePublishContext(params *StorePublishParams, state *task.State) error {
	if state.Output == "" {
		var ok bool

		if params.Passphrase == nil {
			return errorLockerPassFailed
		}

		state.Logger.Debugf("Publishing [%s] from locker [%s]", params.LockerHandle, params.LockerPath)

		if state.Internal != nil {
			if _, ok = state.Internal.(broker.DataLinkInstance); ok {
				return errorDataStoreExist
			} else if _, ok = state.Internal.(broker.DataLink); ok {
				return nil
			}
		}

		return errorDataStoreLinkUnknown
	}

	return nil
}

func handleStorePublishCreate(params *StorePublishParams, state *task.State) error {
	if state.Internal != nil {
		var dataLink = state.Internal.(broker.DataLink)
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var lockerState = NewSecuredLockerState(state)
			var account = brokerClient.GetHostAddr()

			state.Logger.Debugf("Creating data store for data link [%d] on account [%s]", *dataLink.Id, account)

			if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
				var props map[string]string

				defer lockerState.Destroy()
				state.Logger.Debugf("Pushing data store handle [%s] onto name [%s]", params.LockerHandle, params.Name)

				if props, err = lockerState.GetStoreProps(account, params.LockerHandle, params.Passphrase); err == nil {
					var updated *broker.DataLinkInstance
					var instance = &broker.DataLinkInstance{
						Name:        params.Name,
						Description: params.Description,
						Visibility:  params.Visibility,
						DataLinkId:  dataLink.Id,
					}

					if updated, err = brokerClient.CreateDataStore(instance, props); err == nil {
						state.Reportf("Created data store [%s], id: [%d]", params.LockerHandle, *updated.Id)
						return nil
					}
				}
			}
		}

		return err
	}

	return errorDataStoreUnknown
}

func handleStorePublishPretend(params *StorePublishParams, state *task.State) error {
	if errors.Is(state.Error, errorDataStoreExist) {
		var instance = state.Internal.(broker.DataLinkInstance)
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending to update a data store with name [%s]", instance.Name)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d id=\"%d\"\\\n", *instance.Id)
			fmt.Printf("  -d name=\"%s\"\\\n", url.QueryEscape(params.Name))
			fmt.Printf("  -d description=\"%s\"\\\n", url.QueryEscape(params.Description))
			fmt.Printf("  -d dataLinkId=\"%d\"\\\n", *instance.DataLinkId)
			fmt.Printf("  -d visibility=\"%s\"\\\n", params.Visibility)
			fmt.Printf("  -d active=\"%s\"\\\n", cast.ToString(true))
			fmt.Println("  -d props=\"{...}\"")
			fmt.Printf("%s\n", brokerClient.UpdateDataStoreUrl())
			return nil
		}

		return err
	} else if state.Error == nil {
		var link = state.Internal.(broker.DataLink)
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending to create a data store with name [%s]", params.Name)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d name=\"%s\"\\\n", url.QueryEscape(params.Name))
			fmt.Printf("  -d description=\"%s\"\\\n", url.QueryEscape(params.Description))
			fmt.Printf("  -d dataLinkId=\"%d\"\\\n", *link.Id)
			fmt.Printf("  -d visibility=\"%s\"\\\n", params.Visibility)
			fmt.Printf("  -d active=\"%s\"\\\n", cast.ToString(true))
			fmt.Println("  -d props=\"{...}\"")
			fmt.Printf("%s\n", brokerClient.CreateDataStoreUrl())
			return nil
		}

		return err
	}

	return state.Error
}

func handleStorePublishUpdate(params *StorePublishParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, errorDataStoreExist) {
		if state.Internal != nil {
			var instance = state.Internal.(broker.DataLinkInstance)
			var brokerClient broker.Client
			var err error

			if brokerClient, err = params.GetClient(); err == nil {
				var lockerState = NewSecuredLockerState(state)
				var account = brokerClient.GetHostAddr()

				state.Logger.Debugf("Updating data store [%d] from locker [%s]", *instance.Id, params.LockerPath)

				if err = lockerState.Read(params.LockerPath, params.Passphrase); err == nil {
					var props map[string]string

					defer lockerState.Destroy()
					state.Logger.Debugf("Pushing data store handle [%s] onto name [%s]", params.LockerHandle, instance.Name)

					if props, err = lockerState.GetStoreProps(account, params.LockerHandle, params.Passphrase); err == nil {
						var updated *broker.DataLinkInstance

						if updated, err = brokerClient.UpdateDataStore(&instance, props); err == nil {
							state.Reportf("Updated data store [%s], id: [%d]", params.LockerHandle, *updated.Id)
							return nil
						}
					}
				}
			}

			return err
		}

		return errorDataStoreUnknown
	}

	return state.Error
}

func handleStoreSyncComplete(params *StoreFindParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		state.Logger.Debugf("Retrieving data store [%s] for data link [%s]", params.LockerHandle, state.Output)

		if brokerClient, err = params.GetClient(); err == nil {
			var stores []broker.DataLinkInstance

			if stores, err = brokerClient.ListDataStores(); err == nil {
				state.Output = ""

				if len(stores) > 0 {
					var filtered = slicez.Filter(stores, func(instance broker.DataLinkInstance) bool {
						return instance.HasLink(params.DataLink)
					})

					if params.StoreName != "" {
						var named = slicez.Filter(filtered, func(instance broker.DataLinkInstance) bool {
							return strings.EqualFold(params.StoreName, instance.Name)
						})

						return handleStateStoreInstanceList(named, state)
					}

					return handleStateStoreInstanceList(filtered, state)
				}

				return nil
			}
		}

		return err
	}

	return errorLockerDataLinkInvalid
}

func handleStoreSyncContext(params *StoreFindParams, state *task.State) error {
	return handleBaseSyncContext(params.DataLinkParams, state)
}

func handleStoreSyncPretend(params *StoreFindParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending finding data store [%s] for account [%s]",
				params.LockerHandle, brokerClient.GetHostAddr())

			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("%s\n", brokerClient.ListDataStoresUrl())
			state.Logger.Debugf("Filtering data stores for data link [%s]", params.DataLinkParams.ToPublished())
			return nil
		}

		return err
	}

	return errorLockerDataLinkInvalid
}

func handleStoreUpdateComplete(params *StoreUpdateParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		state.Logger.Debugf("Updating store [%s] from locker [%s]", params.LockerHandle, params.LockerPath)

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

				if err = lockerState.updateStore(accountUrl, params.LockerHandle, params.Key, valueEnclave, params.Passphrase); err == nil {
					if err = lockerState.Write(params.LockerPath, params.Passphrase); err == nil {
						state.Reportf("Updated property key [%s] for store [%s] on account [%s]",
							params.Key, params.LockerHandle, accountUrl)
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

func handleStoreUpdateContext(params *StoreUpdateParams, state *task.State) error {
	if state.Output == "" {
		var err error

		if err = handleBaseAddContext(&params.BaseParams, state); err == nil {
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

func handleStoreUpdatePretend(params *StoreUpdateParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient broker.Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var accountUrl = brokerClient.GetHostAddr()

			state.Logger.Debugf("Pretending to update data store [%s] from account [%s]", params.LockerHandle, accountUrl)
			state.Logger.Debugf("Data store property [%s] updated for link [%s/%s:%s]", params.Key, params.Oem, params.Handle, params.Version)
			state.Output = ""
			return nil
		}

		return err
	}

	return errorLockerPathInvalid
}
