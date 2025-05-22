package broker

import "time"

type Session struct {
	Expiry int64
	Token  string
}

func (s *Session) IsExpired() bool {
	return s.Expiry != -1 && s.Expiry <= time.Now().UTC().UnixMilli()
}
