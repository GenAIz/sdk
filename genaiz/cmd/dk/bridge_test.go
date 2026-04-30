package dk

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestSyncBridge_MakeSyncPretenders(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVersion = "version"
	var testLinkString = fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion)
	var testLink = broker.DataLink{Oem: expectedOem, Handle: expectedHandle, Version: expectedVersion}
	var testLedger = config.NewBuilder().Build()
	var testNoSyncOption = &config.BoolOption{Option: config.Option{Key: "key"}}
	var testBridge = &SyncBridge{
		dataLinksWriterFactory: newTestDataLinksWriterFactory([]broker.DataLink{testLink}),
		exportTaskFactory:      newTestExportTaskFactory(),
	}

	actual, err := testBridge.MakeSyncPretenders([]string{testLinkString}, testLedger, testNoSyncOption)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual))
}

func TestSyncBridge_MakeSyncPretenders_NoDataLinks(t *testing.T) {
	var testLedger = config.NewBuilder().Build()
	var testNoSyncOption = &config.BoolOption{}
	var testBridge = &SyncBridge{}
	var actual, err = testBridge.MakeSyncPretenders([]string{}, testLedger, testNoSyncOption)

	assert.Empty(t, actual)
	assert.Nil(t, err)
}

func TestSyncBridge_MakeSyncPretenders_NoSyncOption(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVersion = "version"
	var testLinkString = fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion)
	var testLink = broker.DataLink{Oem: expectedOem, Handle: expectedHandle, Version: expectedVersion}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testNoSyncOption = &config.BoolOption{Option: config.Option{Key: "key"}}
	var testBridge = &SyncBridge{
		collectTaskFactory:     newTestCollectTaskFactory(),
		dataLinksWriterFactory: newTestDataLinksWriterFactory([]broker.DataLink{testLink}),
	}

	testViper.Set(testNoSyncOption.Key, "true")
	actual, err := testBridge.MakeSyncPretenders([]string{testLinkString}, testLedger, testNoSyncOption)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual))
}

func TestSyncBridge_MakeSyncPretenders_ParseDataLinkError(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var testLinkString = fmt.Sprintf("%s/%s:", expectedOem, expectedHandle)
	var testLink = broker.DataLink{Oem: expectedOem, Handle: expectedHandle, Version: "version"}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testNoSyncOption = &config.BoolOption{Option: config.Option{Key: "key"}}
	var testBridge = &SyncBridge{
		collectTaskFactory:     newTestCollectTaskFactory(),
		dataLinksWriterFactory: newTestDataLinksWriterFactory([]broker.DataLink{testLink}),
	}

	testViper.Set(testNoSyncOption.Key, "true")
	actual, err := testBridge.MakeSyncPretenders([]string{testLinkString}, testLedger, testNoSyncOption)
	assert.Error(t, err)
	assert.Nil(t, actual)
}

func TestSyncBridge_MakeSyncWorkers(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVersion = "version"
	var testLinkString = fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion)
	var testLink = broker.DataLink{Oem: expectedOem, Handle: expectedHandle, Version: expectedVersion}
	var testLedger = config.NewBuilder().Build()
	var testNoSyncOption = &config.BoolOption{Option: config.Option{Key: "key"}}
	var testBridge = &SyncBridge{
		dataLinksWriterFactory: newTestDataLinksWriterFactory([]broker.DataLink{testLink}),
		exportTaskFactory:      newTestExportTaskFactory(),
	}

	actual, err := testBridge.MakeSyncWorkers([]string{testLinkString}, testLedger, testNoSyncOption)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual))
}

func TestSyncBridge_MakeSyncWorkers_NoDataLinks(t *testing.T) {
	var testLedger = config.NewBuilder().Build()
	var testNoSyncOption = &config.BoolOption{}
	var testBridge = &SyncBridge{}
	var actual, err = testBridge.MakeSyncWorkers([]string{}, testLedger, testNoSyncOption)

	assert.Empty(t, actual)
	assert.Nil(t, err)
}

func TestSyncBridge_MakeSyncWorkers_NoSyncOption(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVersion = "version"
	var testLinkString = fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion)
	var testLink = broker.DataLink{Oem: expectedOem, Handle: expectedHandle, Version: expectedVersion}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var testNoSyncOption = &config.BoolOption{Option: config.Option{Key: "key"}}
	var testBridge = &SyncBridge{
		collectTaskFactory:     newTestCollectTaskFactory(),
		dataLinksWriterFactory: newTestDataLinksWriterFactory([]broker.DataLink{testLink}),
	}
	var err error

	if err = os.MkdirAll(testLedger.UserPath, 0750); err == nil {
		var fd *os.File

		if fd, err = os.Create(filepath.Join(testLedger.UserPath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
			var actual []task.Worker

			filez.CloseSilently(fd)
			testViper.Set(testNoSyncOption.Key, "true")
			actual, err = testBridge.MakeSyncWorkers([]string{testLinkString}, testLedger, testNoSyncOption)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(actual))
			return
		}
	}

	assert.NoError(t, err)
}

func TestSyncBridge_MakeSyncWorkers_NoSyncOption_NoRepo(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVersion = "version"
	var testLinkString = fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion)
	var testLink = broker.DataLink{Oem: expectedOem, Handle: expectedHandle, Version: expectedVersion}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var testNoSyncOption = &config.BoolOption{Option: config.Option{Key: "key"}}
	var testBridge = &SyncBridge{
		collectTaskFactory:     newTestCollectTaskFactory(),
		dataLinksWriterFactory: newTestDataLinksWriterFactory([]broker.DataLink{testLink}),
	}

	testViper.Set(testNoSyncOption.Key, "true")
	actual, err := testBridge.MakeSyncWorkers([]string{testLinkString}, testLedger, testNoSyncOption)
	assert.Error(t, err)
	assert.Nil(t, actual)
}

func TestNewSyncBridgeBuilder(t *testing.T) {
	var testBridge = NewSyncBridgeBuilder().
		WithDataLinksWriterFactory(newTestDataLinksWriterFactory([]broker.DataLink{})).
		WithCollectLinkTaskFactory(newTestCollectTaskFactory()).
		WithExportLinkTaskFactory(newTestExportTaskFactory()).
		Build()

	assert.NotNil(t, testBridge.dataLinksWriterFactory)
	assert.NotNil(t, testBridge.collectTaskFactory)
	assert.NotNil(t, testBridge.exportTaskFactory)
}

func newTestCollectTaskFactory() CollectLinkTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
		}
	}
}

func newTestDataLinksWriterFactory(current []broker.DataLink) DataLinksWriterFactory {
	var reader = config.DataLinksReader{}

	return func(ledger *config.Ledger, s string) *DataLinksWriter {
		return &DataLinksWriter{
			&config.DataLinksWriter{
				DataLinksReader: *reader.WithCurrent(current),
			},
		}
	}
}

func newTestExportTaskFactory() ExportLinkTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
		}
	}
}
