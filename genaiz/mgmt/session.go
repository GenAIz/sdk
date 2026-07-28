package mgmt

import (
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
