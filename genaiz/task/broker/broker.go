package broker

import (
	"strconv"

	"genaiz.com/genaiz/task/shared"
)

type Broker struct {
	AuthFile string
	HostAddr string
}

type Function struct {
	Id          int // Id is assigned by a publishing Broker and refers to the Smart Function release cycle
	Arches      []string
	Description string
	Fqdn        string
	Handle      string
	Img         string
	Digest      string
	Name        string
	Oem         string
	Type        string
	Version     string
}

func (f Function) asIdentity() *shared.Identity {
	return &shared.Identity{
		Id:      strconv.Itoa(f.Id),
		Hash:    f.Digest,
		Path:    f.Img,
		Version: f.Version,
	}
}

type Provision struct {
	Auth string
	Sf   Function
}

type Session struct {
	Id     int64
	Nco    int64
	Nms    int64
	Flags  int
	UserId int
	Expiry int64
}
