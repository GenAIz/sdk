package config

import (
	"strings"

	"genaiz.com/genaiz/task/broker"
)

type AccountParametric interface {
	BrokerParams() *broker.Broker
}

type accountParams struct {
	ledger *Ledger
	option *StringOption
}

func (ap accountParams) BrokerParams() *broker.Broker {
	if value := ap.ledger.GetString(ap.option); value != "" {
		var tokens = strings.Split(value, "@")

		if len(tokens) == 2 {
			return &broker.Broker{
				AuthFile: ap.ledger.AuthFile,
				HostAddr: tokens[1],
				Username: tokens[0],
			}
		}

		return &broker.Broker{
			AuthFile: ap.ledger.AuthFile,
			HostAddr: value,
		}
	}

	return &broker.Broker{
		AuthFile: ap.ledger.AuthFile,
	}
}

func NewAccountParams(ledger *Ledger, option *StringOption) AccountParametric {
	return &accountParams{
		ledger: ledger,
		option: option,
	}
}
