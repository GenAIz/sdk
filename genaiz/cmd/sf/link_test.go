package sf

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestDataLinkExecutor_makeDataLinkParams(t *testing.T) {
	var expectedOem = "expectedOem"
	var expectedHandle = "expectedHandle"
	var expectedVersion = "expectedVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataLinkExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		DataLinkOptions: newDataLinkOptions(),
	}

	testViper.Set(testExecutor.optionOem.Key, expectedOem)
	testViper.Set(testExecutor.optionHandle.Key, expectedHandle)
	testViper.Set(testExecutor.optionVersion.Key, expectedVersion)
	actual := testExecutor.makeDataLinkParams("")
	assert.Equal(t, actual.Oem, expectedOem)
	assert.Equal(t, actual.Handle, expectedHandle)
	assert.Equal(t, actual.Version, expectedVersion)
	assert.False(t, actual.NoValidation)
}

func TestDataLinkExecutor_makeDataLinkParams_Fqdnv(t *testing.T) {
	var expectedOem = "expectedOem"
	var expectedHandle = "expectedHandle"
	var expectedVersion = "expectedVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataLinkExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		DataLinkOptions: newDataLinkOptions(),
	}

	actual := testExecutor.makeDataLinkParams(fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion))
	assert.Equal(t, actual.Oem, expectedOem)
	assert.Equal(t, actual.Handle, expectedHandle)
	assert.Equal(t, actual.Version, expectedVersion)
	assert.False(t, actual.NoValidation)
}

func TestDataLinkExecutor_makeDataLinkParams_NoValidation(t *testing.T) {
	var expectedOem = "expectedOem"
	var expectedHandle = "expectedHandle"
	var expectedVersion = "expectedVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataLinkExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		DataLinkOptions: newDataLinkOptions(),
	}

	testExecutor.optionNoValidation = cli.NewOptionBuilder().WithKeys(&schema.Keys{Doc: "noVal"}).BuildBoolOption()
	testViper.Set(testExecutor.optionOem.Key, expectedOem)
	testViper.Set(testExecutor.optionNoValidation.Key, true)
	actual := testExecutor.makeDataLinkParams(fmt.Sprintf("%s:%s", expectedHandle, expectedVersion))
	assert.Equal(t, actual.Oem, expectedOem)
	assert.Equal(t, actual.Handle, expectedHandle)
	assert.Equal(t, actual.Version, expectedVersion)
	assert.True(t, actual.NoValidation)
}

func TestDataLinkExecutor_makeDataLinkParams_WithValidation(t *testing.T) {
	var expectedOem = "expectedOem"
	var expectedHandle = "expectedHandle"
	var expectedVersion = "expectedVersion"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataLinkExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		DataLinkOptions: newDataLinkOptions(),
	}

	testExecutor.optionNoValidation = cli.NewOptionBuilder().WithKeys(&schema.Keys{Doc: "noVal"}).BuildBoolOption()
	testViper.Set(testExecutor.optionOem.Key, expectedOem)
	testViper.Set(testExecutor.optionVersion.Key, expectedVersion)
	testViper.Set(testExecutor.optionNoValidation.Key, false)
	actual := testExecutor.makeDataLinkParams(expectedHandle)
	assert.Equal(t, actual.Oem, expectedOem)
	assert.Equal(t, actual.Handle, expectedHandle)
	assert.Equal(t, actual.Version, expectedVersion)
	assert.False(t, actual.NoValidation)
}

func TestDataLinkOptions_addDefiners(t *testing.T) {
	var testOptions = newDataLinkOptions()

	assert.Equal(t, 4, len(testOptions.addDefiners()))
}

func TestDataLinkOptions_removeDefiners(t *testing.T) {
	var testOptions = newDataLinkOptions()

	assert.Equal(t, 3, len(testOptions.removeDefiners()))
}

func newListLinksTaskCompleteStub(checks func(params *broker.DataLinkParams)) ListLinksTaskFactory {
	return func() *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			Name: "init_test",
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				checks(params)
				return nil
			},
		}
	}
}

func newListLinksTaskPretendStub(flag *bool) ListLinksTaskFactory {
	return func() *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataLinkParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newDataLinkOptions() *DataLinkOptions {
	return &DataLinkOptions{
		optionOem: cli.NewOptionBuilder().
			WithKeys(&schema.Keys{Doc: "oem"}).
			BuildStringOption(),
		optionHandle: cli.NewOptionBuilder().
			WithKeys(&schema.Keys{Doc: "handle"}).
			BuildStringOption(),
		optionVersion: cli.NewOptionBuilder().
			WithKeys(&schema.Keys{Doc: "version"}).
			BuildStringOption(),
	}
}

func newDataLinksWriterTestFactory(current []broker.DataLink) dk.DataLinksWriterFactory {
	return func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
		var reader = &config.DataLinksReader{}

		return &dk.DataLinksWriter{
			DataLinksWriter: &config.DataLinksWriter{
				DataLinksReader: *reader.WithCurrent(current),
			},
		}
	}
}

func newSyncLinkCompleteCapture(capture *broker.DataLinkParams) SyncLinksTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newSyncLinkPretendCapture(capture *broker.DataLinkParams) SyncLinksTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataLinkParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
