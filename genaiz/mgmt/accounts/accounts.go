package accounts

import (
	"encoding/json"
	"fmt"
	"strings"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserAccountsFactory interface {
	mgmt.Factory[broker.AccountListParams, []UserAccount]
}

type UserAccountsProvider = mgmt.Provider[[]UserAccount]
type UserAccountsProviderFactory = mgmt.ProviderFactory[broker.AccountListParams, []UserAccount]

type UserAccount struct {
	Active   bool `cli:"Active" selected:"*"`
	Name     string
	Username string `cli:"User"`
	HostAddr string `cli:"Host"`
	Created  string `cli:"Created"`
	Expiry   string `cli:"Expires" redGreen:"Expired"`
	Expired  bool   `cli:"Expired,noShow"`

	matched string
}

func (ua UserAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Active   bool   `json:"active"`
		Username string `json:"username"`
		HostAddr string `json:"hostAddr"`
		Created  string `json:"created,omitempty"`
		Expiry   string `json:"expiry"`
	}{
		Active:   ua.Active,
		Username: ua.Username,
		HostAddr: ua.HostAddr,
		Created:  ua.Created,
		Expiry:   ua.Expiry,
	})
}

func (ua UserAccount) MarshalSlice() ([]string, error) {
	return []string{
		stringz.YesOrNo(ua.Active),
		ua.Username,
		ua.HostAddr,
		stringz.FirstNonEmpty(ua.Created, "-"),
		stringz.FirstNonEmpty(ua.Expiry, "-"),
		stringz.YesOrNo(ua.Expired),
	}, nil
}

func (ua UserAccount) Match(filter string) *UserAccount {
	var lowFilter = strings.ToLower(filter)
	var matched string

	if strings.EqualFold(ua.Name, filter) ||
		strings.HasPrefix(ua.Name, filter) ||
		strings.HasSuffix(ua.Name, filter) {
		matched = ua.Name
	}

	if strings.EqualFold(ua.HostAddr, lowFilter) ||
		strings.HasPrefix(ua.HostAddr, lowFilter) ||
		strings.HasSuffix(ua.HostAddr, lowFilter) {
		matched = ua.HostAddr
	}

	if matched != "" {
		return &UserAccount{
			Active:   ua.Active,
			Name:     ua.Name,
			Username: ua.Username,
			HostAddr: ua.HostAddr,
			Created:  ua.Created,
			Expiry:   ua.Expiry,
			Expired:  ua.Expired,
			matched:  matched,
		}
	}

	return nil
}

func (ua UserAccount) Matched() string {
	if ua.matched == "" {
		return ua.HostAddr
	}

	return ua.matched
}

type userAccountsFactory struct {
	mgmt.ProviderFactory[broker.AccountListParams, []UserAccount]
}

func (uf userAccountsFactory) provider(factory mgmt.ProviderFactory[broker.AccountListParams, []UserAccount]) mgmt.Provider[[]UserAccount] {
	return &userAccountsProvider{
		Plan: task.Plan{
			Logger: factory.Logger,
		},

		filter: factory.Filter,
		format: factory.Format,
		params: factory.Params,

		accountListTaskFactory: broker.AccountTasks.ListTaskFactory,
	}
}

type userAccountsProvider struct {
	task.Plan

	filter *string
	format timez.Formatter
	params *broker.AccountListParams

	accountListTaskFactory broker.AccountListTaskFactory
}

func (up userAccountsProvider) Get() ([]UserAccount, task.Error) {
	var authAccounts *broker.AuthData
	var failure interface{}

	up.OnReturn = func(i interface{}) { authAccounts = i.(*broker.AuthData) }
	up.OnFailure = func(i interface{}) { failure = i }
	up.Sequence(task.NewWorker(up.params, up.accountListTaskFactory()))

	if failure == nil {
		var result = make([]UserAccount, 0)

		for i, aa := range authAccounts.Accounts {
			var username string

			if strings.EqualFold("token", aa.Username) && len(aa.Token) >= 10 {
				username = aa.Token[:10]
			} else if strings.Contains(aa.Username, "@") {
				username = strings.Split(aa.Username, "@")[0]
			} else {
				username = aa.Username
			}

			account := &UserAccount{
				Active:   i == authAccounts.Active,
				Name:     fmt.Sprintf("%s@%s", username, aa.HostAddr),
				HostAddr: aa.HostAddr,
				Username: username,
				Expiry:   up.format.FormatMillis(aa.Expiry),
				Created:  up.format.FormatMillis(aa.Created),
				Expired:  aa.IsExpired(),
			}

			if up.filter == nil {
				result = append(result, *account)
			} else if matched := account.Match(*up.filter); matched != nil {
				result = append(result, *matched)
			}
		}

		return result, nil
	}

	return nil, failure.(task.Error)
}

func NewUserAccountsFactory() UserAccountsFactory {
	var factory = &userAccountsFactory{
		ProviderFactory: mgmt.ProviderFactory[broker.AccountListParams, []UserAccount]{
			Format: timez.NewMidnightFormatter(),
		},
	}

	return factory.WithProvider(factory.provider)
}
