package cmd

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/version"
)

type errorWriter struct {
}

func (ew errorWriter) Write(p []byte) (int, error) {
	return -1, errors.New("test")
}

func TestRunnerOptions_Confirm_Error(t *testing.T) {
	var calledDisplay bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testInput = bytes.NewReader([]byte("Y\n"))
	var testRunner = &RunnerOptions{
		runConfirm: newOptionRunConfirm(),
		stdIn:      testInput,
		stdOut:     &errorWriter{},
	}
	var testDisplay = func() {
		calledDisplay = true
	}

	testViper.Set(testRunner.runConfirm.Param, true)
	assert.False(t, testRunner.Confirm(testLedger, testDisplay))
	assert.True(t, calledDisplay)
}

func TestRunnerOptions_Confirm_InvalidButNo(t *testing.T) {
	var calledDisplay bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testInput = bytes.NewReader([]byte("\ntest\nno\n"))
	var testOutput = bytes.Buffer{}
	var testRunner = &RunnerOptions{
		runConfirm: newOptionRunConfirm(),
		stdIn:      testInput,
		stdOut:     io.Writer(&testOutput),
	}
	var testDisplay = func() {
		calledDisplay = true
	}

	testViper.Set(testRunner.runConfirm.Param, true)
	assert.False(t, testRunner.Confirm(testLedger, testDisplay))
	assert.True(t, calledDisplay)
}

func TestRunnerOptions_Confirm_NoDisplay(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testInput = bytes.NewReader([]byte("Yes\n"))
	var testOutput = bytes.Buffer{}
	var testRunner = &RunnerOptions{
		runConfirm: newOptionRunConfirm(),
		stdIn:      testInput,
		stdOut:     io.Writer(&testOutput),
	}

	testViper.Set(testRunner.runConfirm.Param, true)
	assert.True(t, testRunner.Confirm(testLedger))
}

func TestRunnerOptions_Confirm_No(t *testing.T) {
	var calledDisplay bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testInput = bytes.NewReader([]byte("n\n"))
	var testOutput = bytes.Buffer{}
	var testRunner = &RunnerOptions{
		runConfirm: newOptionRunConfirm(),
		stdIn:      testInput,
		stdOut:     io.Writer(&testOutput),
	}
	var testDisplay = func() {
		calledDisplay = true
	}

	testViper.Set(testRunner.runConfirm.Param, true)
	assert.False(t, testRunner.Confirm(testLedger, testDisplay))
	assert.True(t, calledDisplay)
}

func TestRunnerOptions_Confirm_Omit(t *testing.T) {
	var calledDisplay bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testRunner = &RunnerOptions{
		runConfirm: newOptionRunConfirm(),
	}
	var testDisplay = func() {
		calledDisplay = true
	}

	testViper.Set(testRunner.runConfirm.Param, false)
	assert.True(t, testRunner.Confirm(testLedger, testDisplay))
	assert.False(t, calledDisplay)
}

func TestRunnerOptions_Confirm_Yes(t *testing.T) {
	var calledDisplay bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testInput = bytes.NewReader([]byte("Y\n"))
	var testOutput = bytes.Buffer{}
	var testRunner = &RunnerOptions{
		runConfirm: newOptionRunConfirm(),
		stdIn:      testInput,
		stdOut:     io.Writer(&testOutput),
	}
	var testDisplay = func() {
		calledDisplay = true
	}

	testViper.Set(testRunner.runConfirm.Param, true)
	assert.True(t, testRunner.Confirm(testLedger, testDisplay))
	assert.True(t, calledDisplay)
}

func TestRunnerOptions_Dry(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testRunner = &RunnerOptions{
		runDry: newOptionRunDry(),
	}

	testViper.Set(testRunner.runDry.Param, true)
	assert.True(t, testRunner.Dry(testLedger))
}

func TestRunnerOptions_Pretend(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testRunner = &RunnerOptions{
		runPretend: newOptionRunPretend(),
	}

	testViper.Set(testRunner.runPretend.Param, true)
	assert.True(t, testRunner.Pretend(testLedger))
}

func TestRunnerOptions_allDefiners(t *testing.T) {
	var testOptions = NewRunnerOptions()
	var testDefiners = testOptions.allDefiners()

	assert.NotEmpty(t, testOptions.logLevel)
	assert.NotEmpty(t, testOptions.logFormat)
	assert.NotEmpty(t, testOptions.runConfig)
	assert.NotEmpty(t, testOptions.runConfirm)
	assert.NotEmpty(t, testOptions.runDry)
	assert.NotEmpty(t, testOptions.runPretend)
	assert.Contains(t, testDefiners, testOptions.logLevel)
	assert.Contains(t, testDefiners, testOptions.logFormat)
	assert.Contains(t, testDefiners, testOptions.runConfig)
	assert.Contains(t, testDefiners, testOptions.runConfirm)
	assert.Contains(t, testDefiners, testOptions.runDry)
	assert.Contains(t, testDefiners, testOptions.runPretend)
}

func TestNew(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = New(testLedger)

	testLedger.LoggerFactory(testLedger)
	assert.EqualValues(t, version.GetVersion(), testCmd.Version)
	assert.Equal(t, 4, len(testCmd.Commands()))
}

func TestNew_ErrorLogFactory(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = New(testLedger)
	var testLogLevel = cli.Options.Solutions.LogLevel().BuildStringOption()

	testViper.Set(testLogLevel.Key, "invalid")
	testLedger.InitLogging()
	assert.EqualValues(t, version.GetVersion(), testCmd.Version)
}

func Test_getFormatter(t *testing.T) {
	var testLogger = logrus.New()
	var testEntry = logrus.NewEntry(testLogger)
	var defFormatter, jsonFormatter, customFormater logrus.Formatter
	var expected = []byte("message")
	var actual []byte
	var err error

	testEntry.Message = string(expected)
	defFormatter = getFormatter("")
	assert.NotEmpty(t, defFormatter)
	actual, err = defFormatter.Format(testEntry)
	assert.NoError(t, err)
	assert.Contains(t, string(actual), testEntry.Message)

	jsonFormatter = getFormatter("json")
	assert.NotEmpty(t, jsonFormatter)
	actual, err = jsonFormatter.Format(testEntry)
	assert.NoError(t, err)
	assert.Contains(t, string(actual), testEntry.Message)

	customFormater = getFormatter("%msg%")
	assert.NotEmpty(t, customFormater)
	actual, err = customFormater.Format(testEntry)
	assert.NoError(t, err)
	assert.Contains(t, string(actual), testEntry.Message)
}

func Test_getLevel(t *testing.T) {
	var actual logrus.Level
	var err error

	actual, err = getLevel("d")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.DebugLevel, actual)
	actual, err = getLevel("v")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.DebugLevel, actual)
	actual, err = getLevel("debug")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.DebugLevel, actual)
	actual, err = getLevel("verbose")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.DebugLevel, actual)

	actual, err = getLevel("e")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.ErrorLevel, actual)
	actual, err = getLevel("err")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.ErrorLevel, actual)
	actual, err = getLevel("error")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.ErrorLevel, actual)

	actual, err = getLevel("")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.InfoLevel, actual)
	actual, err = getLevel("i")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.InfoLevel, actual)
	actual, err = getLevel("nfo")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.InfoLevel, actual)
	actual, err = getLevel("info")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.InfoLevel, actual)

	actual, err = getLevel("q")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.FatalLevel, actual)
	actual, err = getLevel("qq")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.FatalLevel, actual)
	actual, err = getLevel("quiet")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.FatalLevel, actual)

	actual, err = getLevel("t")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.TraceLevel, actual)
	actual, err = getLevel("trc")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.TraceLevel, actual)
	actual, err = getLevel("trace")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.TraceLevel, actual)

	actual, err = getLevel("w")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.WarnLevel, actual)
	actual, err = getLevel("warn")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.WarnLevel, actual)
	actual, err = getLevel("warning")
	assert.NoError(t, err)
	assert.EqualValues(t, logrus.WarnLevel, actual)

	actual, err = getLevel("INVALID")
	assert.Error(t, err)
	assert.EqualValues(t, logrus.InfoLevel, actual)
}

func Test_newOptionOverrideConfig_defaultGetter(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = newOptionRunConfig()
	var expectedDir = "/tmp"

	testLedger.WorkDir = expectedDir
	assert.EqualValues(t, expectedDir, testOption.DefaultGetter(testLedger))
}
