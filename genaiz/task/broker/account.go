package broker

import "genaiz.com/genaiz/task"

var (
	AccountTasks struct {
		ListTaskFactory AccountListTaskFactory
	}
)

type AccountListTask = *task.Task[AccountListParams]
type AccountListTaskFactory = func() AccountListTask

type AccountListParams struct {
	Broker
	Expired bool
}

func handleAccountListContext(params *AccountListParams, state *task.State) error {
	var auth = NewAuthData(params.AuthFile)

	if len(auth.Accounts) > 0 {
		state.Logger.Debug("Accounts found")
		return nil
	}

	return ErrorSessionsEmpty
}

func handleAccountListComplete(params *AccountListParams, state *task.State) error {
	var auth = NewAuthData(params.AuthFile)
	var effectiveAuth *AuthData

	if auth.Active < 0 || auth.Active >= len(auth.Accounts) {
		effectiveAuth = &AuthData{
			Active:   0,
			Accounts: auth.Accounts,
		}
		state.Logger.Debug("Invalid activated index, defaulting to 0")
	} else {
		effectiveAuth = auth
	}

	if !params.Expired {
		var accountsFiltered []*AuthAccount
		var activeFiltered = effectiveAuth.Active

		state.Logger.Debug("Filtering expired accounts")

		for i, acc := range effectiveAuth.Accounts {
			if !acc.IsExpired() {
				accountsFiltered = append(accountsFiltered, acc)

				if i == activeFiltered {
					activeFiltered = len(accountsFiltered) - 1
				}
			} else if i == activeFiltered {
				activeFiltered = 0
			}
		}

		state.Internal = &AuthData{
			Active:   activeFiltered,
			Accounts: accountsFiltered,
		}
		return nil
	}

	state.Internal = effectiveAuth
	return nil
}

func init() {
	AccountTasks.ListTaskFactory = newAccountListTask
}

func newAccountListTask() *task.Task[AccountListParams] {
	return &task.Task[AccountListParams]{
		Name:       "auth-data",
		OnPrepare:  handleAccountListContext,
		OnComplete: handleAccountListComplete,
	}
}
