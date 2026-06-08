package mgmt

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserAccountFacade Facade[[]UserAccount, broker.AuthParams]

type SessionListTaskFactory func() *task.Task[broker.AuthParams]

type UserAccount struct {
	Active   bool `cli:"Active" selected:"*"`
	Name     string
	Username string `cli:"User"`
	HostAddr string `cli:"Host"`
	Created  int64  `cli:"Created"`
	Expiry   int64  `cli:"Expires" redGreen:"Expired"`
	Expired  bool   `cli:"Expired,noShow"`

	matched string
}

func (ua UserAccount) MarshalJSON() ([]byte, error) {
	var created string

	if ua.Created > 0 {
		created = createdFormatter.FormatMillis(ua.Created)
	}

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
		Created:  created,
		Expiry:   expiryFormatter.FormatMillis(ua.Expiry),
	})
}

func (ua UserAccount) MarshalSlice() ([]string, error) {
	var created string

	if ua.Created > 0 {
		created = createdFormatter.FormatMillis(ua.Created)
	} else {
		created = "-"
	}

	return []string{
		stringz.YesOrNo(ua.Active),
		ua.Username,
		ua.HostAddr,
		created,
		expiryFormatter.FormatMillis(ua.Expiry),
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

type userAccountsFacade struct {
	baseLoggingFacade
	params *broker.AuthParams
}

func (uaf *userAccountsFacade) Filtering(filter string) Provider[[]UserAccount] {
	return &userAccountsProvider{
		Plan: task.Plan{
			Logger: uaf.logger,
		},
		filter:                 filter,
		params:                 uaf.params,
		sessionListTaskFactory: broker.NewAuthTask,
	}
}

func (uaf *userAccountsFacade) Provider() Provider[[]UserAccount] {
	return uaf.Filtering("")
}

func (uaf *userAccountsFacade) WithLogger(logger *logrus.Logger) Facade[[]UserAccount, broker.AuthParams] {
	uaf.logger = logger
	return uaf
}

func (uaf *userAccountsFacade) WithParams(params *broker.AuthParams) Facade[[]UserAccount, broker.AuthParams] {
	uaf.params = params
	return uaf
}

type userAccountsProvider struct {
	task.Plan
	filter                 string
	params                 *broker.AuthParams
	sessionListTaskFactory SessionListTaskFactory
}

func (uap userAccountsProvider) Get() ([]UserAccount, task.Error) {
	var authAccounts *broker.AuthData
	var failure interface{}

	uap.OnReturn = func(i interface{}) { authAccounts = i.(*broker.AuthData) }
	uap.OnFailure = func(i interface{}) { failure = i }
	uap.Sequence(task.NewWorker(uap.params, uap.sessionListTaskFactory()))

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
				Expiry:   aa.Expiry,
				Created:  aa.Created,
				Expired:  aa.IsExpired(),
			}

			if uap.filter == "" {
				result = append(result, *account)
			} else if matched := account.Match(uap.filter); matched != nil {
				result = append(result, *matched)
			}
		}

		return result, nil
	}

	return nil, failure.(task.Error)
}

func NewUserAccountFacade() UserAccountFacade {
	return &userAccountsFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}
