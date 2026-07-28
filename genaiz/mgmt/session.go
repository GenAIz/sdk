package mgmt

import (
	"encoding/json"
	"time"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task/broker"
)

type UserSession struct {
	Id      int64 `cli:"Id"`
	UserId  int   `cli:"UserId"`
	Created int64 `cli:"Created"`
	Expiry  int64 `cli:"Expires" redGreen:"Expired"`
	Expired bool  `cli:"Expired,noShow"`
	Flags   *int
}

func (us UserSession) MarshalJSON() ([]byte, error) {
	var created, expiry string

	if us.Created > 0 {
		created = createdFormatter.FormatMillis(us.Created)
	}

	if us.Expiry > 0 {
		expiry = expiryFormatter.FormatMillis(us.Expiry)
	}

	return json.Marshal(&struct {
		Id      int64  `json:"id,omitempty"`
		UserId  int    `json:"userId,omitempty"`
		Created string `json:"created,omitempty"`
		Expiry  string `json:"expiry,omitempty"`
		Expired bool   `json:"expired"`
		Flags   *int   `json:"flags,omitempty"`
	}{
		Id:      us.Id,
		UserId:  us.UserId,
		Created: created,
		Expiry:  expiry,
		Expired: us.Expired,
		Flags:   us.Flags,
	})
}

func (us UserSession) MarshalSlice() ([]string, error) {
	var created string

	if us.Created > 0 {
		created = createdFormatter.FormatMillis(us.Created)
	} else {
		created = "-"
	}

	return []string{
		cast.ToString(us.Id),
		cast.ToString(us.UserId),
		created,
		expiryFormatter.FormatMillis(us.Expiry),
		stringz.YesOrNo(us.Expired),
	}, nil
}

func ToUserSession(session *broker.Session) *UserSession {
	return &UserSession{
		Id:      session.Id,
		UserId:  session.UserId,
		Created: session.Nco,
		Expiry:  session.Expiry,
		Expired: session.Expiry != -1 && session.Expiry <= time.Now().UTC().UnixMilli(),
		Flags:   &session.Flags,
	}
}
