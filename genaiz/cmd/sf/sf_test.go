package sf

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-it/mock"
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
	var testCli = NewSfCli(testInteractive, testDry, testPretend)
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
	var testCli = NewSfCli(testInteractive, testDry, testPretend)
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
	var testCli = NewSfCli(testInteractive, testDry, testPretend)
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
	var testCli = NewSfCli(testInteractive, testDry, testPretend)
	var actualExec = &ExecutorStub{}

	defer patch.Unpatch()
	testCli.Exec(testLedger, actualExec)
	assert.False(t, actualExec.calledProceed)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 0, patch.CalledWith)
}

func TestCli_allDefiners(t *testing.T) {
	var testSfCli = NewSfCli(nil, nil, nil)
	var testCliDefiners = testSfCli.allDefiners()

	assert.NotEmpty(t, testSfCli.optionDockerContext)
	assert.NotEmpty(t, testSfCli.optionDockerFile)
	assert.NotEmpty(t, testSfCli.optionDockerTag)
	assert.NotEmpty(t, testSfCli.optionDockerVersion)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerContext)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerFile)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerTag)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerVersion)
}

func TestNewSf(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testSf = NewSf(testLedger, nil, nil, nil)
	var testSubCommand = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	var testDockerContextOption = newOptionDockerContext()
	var testDockerFileOption = newOptionDockerFile()
	var expectedFile = "dockerFile"

	assert.NotEmpty(t, testSf.Commands())
	assert.NotEmpty(t, testSf.PersistentFlags())
	assert.NoError(t, testSf.PersistentFlags().Lookup(testDockerFileOption.Param).Value.Set(expectedFile))
	testSubCommand.Flags().AddFlagSet(testSf.PersistentFlags())
	testLedger.WorkDir = "/tmp"
	testSf.PersistentPreRun(testSubCommand, []string{})
	assert.EqualValues(t, "/tmp", testSf.PersistentFlags().Lookup(testDockerContextOption.Param).Value.String())
	assert.EqualValues(t, "/tmp/dockerFile", testSf.PersistentFlags().Lookup(testDockerFileOption.Param).Value.String())
}

func Test_newOptionDockerTag_DefaultSetter(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = newOptionDockerTag()
	var expectedTag = "test"

	testLedger.WorkDir = filepath.Join("/tmp", "genaiz", expectedTag)
	assert.EqualValues(t, expectedTag, testOptions.DefaultSetter(testLedger))
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
