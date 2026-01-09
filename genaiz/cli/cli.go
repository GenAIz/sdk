// Package cli represents functionality to help gather and display information to the user
package cli

import (
	"fmt"
	"os"

	"genaiz.com/genaiz/config"
)

const (
	DefaultValueForNil = "__internal_nil"
)

type Decisive func(*config.Ledger) bool

type Interactive func(*config.Ledger, ...func()) bool

type Executor interface {
	Display()

	Pretend()

	Proceed()
}

type BaseCli struct {
	Confirm Interactive
	Dry     Decisive
	Pretend Decisive
}

func (c BaseCli) Exec(ledger *config.Ledger, executor Executor) {
	if !c.isDry(ledger, executor.Display) {
		if c.isPretend(ledger) {
			executor.Pretend()
		} else {
			c.execConfirm(ledger, executor.Proceed, executor.Display)
		}
	}
}

func (c BaseCli) execConfirm(ledger *config.Ledger, exec func(), display ...func()) {
	if c.Confirm != nil && c.Confirm(ledger, display...) {
		exec()
	} else {
		fmt.Println("Cancelled, exiting")
		os.Exit(0)
	}
}

func (c BaseCli) isDecisive(ledger *config.Ledger, decisive Decisive, display ...func()) bool {
	if decisive != nil && decisive(ledger) {
		for _, d := range display {
			d()
		}

		return true
	}

	return false
}

func (c BaseCli) isDry(ledger *config.Ledger, display ...func()) bool {
	return c.isDecisive(ledger, c.Dry, display...)
}

func (c BaseCli) isPretend(ledger *config.Ledger) bool {
	return c.isDecisive(ledger, c.Pretend)
}
