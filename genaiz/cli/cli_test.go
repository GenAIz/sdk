package cli

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
)

type ExecutorStub struct {
	calledDisplay bool
	calledPretend bool
	calledProceed bool
}

func (es *ExecutorStub) Display() {
	es.calledDisplay = true
}

func (es *ExecutorStub) Pretend() {
	es.calledPretend = true
}

func (es *ExecutorStub) Proceed() {
	es.calledProceed = true
}

func TestCli_ExecDry(t *testing.T) {
	var testDry = newDecisiveStub(true)
	var testPretend = newDecisiveStub(false)
	var testInteractive = newInteractiveStub(false)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = newBaseCli(testInteractive, testDry, testPretend)
	var actualExec = &ExecutorStub{}

	testCli.Exec(testLedger, actualExec)
	assert.True(t, actualExec.calledDisplay)
}

func TestCli_ExecPretend(t *testing.T) {
	var testDry = newDecisiveStub(false)
	var testPretend = newDecisiveStub(true)
	var testInteractive = newInteractiveStub(false)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = newBaseCli(testInteractive, testDry, testPretend)
	var actualExec = &ExecutorStub{}

	testCli.Exec(testLedger, actualExec)
	assert.True(t, actualExec.calledPretend)
}

func TestCli_ExecProceed_Confirmed(t *testing.T) {
	var testDry = newDecisiveStub(false)
	var testPretend = newDecisiveStub(false)
	var testInteractive = newInteractiveStub(true)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = newBaseCli(testInteractive, testDry, testPretend)
	var actualExec = &ExecutorStub{}

	testCli.Exec(testLedger, actualExec)
	assert.True(t, actualExec.calledProceed)
}

func TestCli_ExecProceed_Rejected(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testDry = newDecisiveStub(false)
	var testPretend = newDecisiveStub(false)
	var testInteractive = newInteractiveStub(false)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = newBaseCli(testInteractive, testDry, testPretend)
	var actualExec = &ExecutorStub{}

	defer patch.Unpatch()
	testCli.Exec(testLedger, actualExec)
	assert.False(t, actualExec.calledProceed)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 0, patch.CalledWith)
}

func newBaseCli(confirm Interactive, dry, pretend Decisive) *BaseCli {
	return &BaseCli{
		Confirm: confirm,
		Dry:     dry,
		Pretend: pretend,
	}
}

func newDecisiveStub(decision bool) Decisive {
	return func(l *config.Ledger) bool {
		return decision
	}
}

func newInteractiveStub(decision bool) Interactive {
	return func(l *config.Ledger, f ...func()) bool {
		return decision
	}
}
