package lk

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	gio "genaiz.com/genaiz-lib/mock/io"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/locker"
)

func TestStorePublishExecutor_Publish(t *testing.T) {
	var expectedHandle = "expectedHandle"
	var testOutput bytes.Buffer
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithOutput(io.Writer(&testOutput)).
		Build()
	var testExecutor = &StorePublishExecutor{
		BasePublishExecutor: BasePublishExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			PublishOptions: NewStorePublishOptions(),
		},
	}
	var expectedLockerPath = filepath.Join(testLedger.UserPath, "locker.bin")

	assert.NoError(t, testExecutor.Publish(expectedHandle))
	assert.Equal(t, expectedHandle, testExecutor.handleArg)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`locker:[\s\t]*`+expectedLockerPath), actual)
}

func TestStorePublishExecutor_Display(t *testing.T) {
	var expectedAccount = "expectedAccount"
	var expectedLockerPath = "expectedLocker"
	var expectedHandle = "expectedHandle"
	var expectedName = "expectedName"
	var expectedDesc = "expectedDesc"
	var testOutput bytes.Buffer
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		WithOutput(io.Writer(&testOutput)).
		Build()
	var testExecutor = &StorePublishExecutor{
		BasePublishExecutor: BasePublishExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			PublishOptions: NewStorePublishOptions(),
		},
	}

	testViper.Set(testExecutor.optionAccount.Key, expectedAccount)
	testViper.Set(testExecutor.optionLocker.Key, expectedLockerPath)
	testViper.Set(testExecutor.optionName.Key, expectedName)
	testViper.Set(testExecutor.optionDescription.Key, expectedDesc)
	testViper.Set(testExecutor.optionVisibility.Key, broker.VisibilityOrg)
	assert.NoError(t, testExecutor.Publish(expectedHandle))
	assert.Equal(t, expectedHandle, testExecutor.handleArg)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`locker:[\s\t]*`+expectedLockerPath), actual)
	assert.Regexp(t, regexp.MustCompile(`account:[\s\t]*`+expectedAccount), actual)
	assert.Regexp(t, regexp.MustCompile(`name:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(`description:[\s\t]*`+expectedDesc), actual)
	assert.Regexp(t, regexp.MustCompile(`visibility:[\s\t]*`+broker.VisibilityOrg), actual)
}

func TestStorePublishExecutor_Pretend(t *testing.T) {
	var capturedFindParams, capturedSyncParams locker.StoreFindParams
	var capturedPublishParams locker.StorePublishParams
	var capturedDataLinkParams broker.DataLinkParams
	var testOptions = NewStorePublishOptions()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testExecutor = &StorePublishExecutor{
		BasePublishExecutor: BasePublishExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			PublishOptions: testOptions,

			handleArg:               "expectedHandle",
			accountParams:           config.NewAccountParams(testLedger, testOptions.optionAccount),
			dataLinkFindTaskFactory: newDataLinkFindTaskPretendStub(&capturedDataLinkParams, nil),
		},

		storeFindTaskFactory:    newStoreFindTaskPretendStub(&capturedFindParams),
		storeSyncTaskFactory:    newStoreSyncTaskPretendStub(&capturedSyncParams),
		storePublishTaskFactory: newStorePublishTaskPretendStub(&capturedPublishParams),
	}

	testViper.Set(testOptions.optionVisibility.Key, broker.VisibilityOrg)
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.NotEmpty(t, capturedFindParams)
	assert.NotEmpty(t, capturedSyncParams)
	assert.NotEmpty(t, capturedPublishParams)
	assert.Equal(t, capturedPublishParams.Visibility, broker.VisibilityOrg)
	assert.NotEmpty(t, capturedDataLinkParams)
}

func TestStorePublishExecutor_Proceed(t *testing.T) {
	var capturedFindParams, capturedSyncParams locker.StoreFindParams
	var capturedPublishParams locker.StorePublishParams
	var capturedDataLinkParams broker.DataLinkParams
	var testOptions = NewStorePublishOptions()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testExecutor = &StorePublishExecutor{
		BasePublishExecutor: BasePublishExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			PublishOptions: testOptions,

			handleArg:               "expectedHandle",
			accountParams:           config.NewAccountParams(testLedger, testOptions.optionAccount),
			dataLinkFindTaskFactory: newDataLinkFindTaskProceedStub(&capturedDataLinkParams, nil),
		},

		storeFindTaskFactory:    newStoreFindTaskProceedStub(&capturedFindParams),
		storeSyncTaskFactory:    newStoreSyncTaskProceedStub(&capturedSyncParams),
		storePublishTaskFactory: newStorePublishTaskProceedStub(&capturedPublishParams),
	}

	testViper.Set(testOptions.optionVisibility.Key, broker.VisibilityOrg)
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, capturedFindParams)
	assert.NotEmpty(t, capturedSyncParams)
	assert.NotEmpty(t, capturedPublishParams)
	assert.Equal(t, capturedPublishParams.Visibility, broker.VisibilityOrg)
	assert.NotEmpty(t, capturedDataLinkParams)
}

func TestNewStorePublishExecutor(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCmd = &cobra.Command{}
	var testFactory = newStorePublishExecutorFactory(testLedger, &Cli{}, NewStorePublishOptions())

	assert.NotNil(t, testFactory(testCmd))
}

func TestStoreLockerExecutor_Add(t *testing.T) {
	var testOutput bytes.Buffer
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(&testOutput)).
		Build()
	var testOptions = NewStoreAddOptions()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			LockerOptions: testOptions,
		},
	}
	var expectedOem = "dataLinkOem"
	var expectedHandle = "dataLinkHandle"
	var expectedVersion = "dataLinkVersion"
	var testHandle = "handleArg"
	var testDatalinkArg = fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion)

	assert.NoError(t, testExecutor.Add(testHandle, testDatalinkArg))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+testHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`datalink-oem:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(`datalink-handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`datalink-version:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`datalink-seq:[\s\t]*\n`), actual)
}

func TestStoreLockerExecutor_Add_EmptyOem(t *testing.T) {
	var testExecutor = &StoreLockerExecutor{}
	var testHandle = "handleArg"
	// parsing assumes single atoms to be handles
	var testDatalinkArg = "dataLinkArg"

	assert.ErrorIs(t, testExecutor.Add(testHandle, testDatalinkArg), errorDataLinkOemRequired)
}

func TestStoreLockerExecutor_Add_EmptyVersion(t *testing.T) {
	var testExecutor = &StoreLockerExecutor{}
	var testHandle = "handleArg"
	// no support for default versioning
	var testDatalinkArg = "dataLinkOem/dataLinkHandle"

	assert.ErrorIs(t, testExecutor.Add(testHandle, testDatalinkArg), errorDataLinkVersionRequired)
}

func TestStoreLockerExecutor_Add_InvalidSeq(t *testing.T) {
	var testExecutor = &StoreLockerExecutor{}
	var testHandle = "handleArg"
	// no support for default versioning
	var testDatalinkArg = "dataLinkOem/dataLinkHandle:ver-rc-notSequence"

	assert.ErrorIs(t, testExecutor.Add(testHandle, testDatalinkArg), errorDataLinkSequenceInvalid)
}

func TestStoreLockerExecutor_Add_WithSequence(t *testing.T) {
	var testOutput bytes.Buffer
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(&testOutput)).
		Build()
	var testOptions = NewStoreAddOptions()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			LockerOptions: testOptions,
		},
	}
	var expectedOem = "dataLinkOem"
	var expectedHandle = "dataLinkHandle"
	var expectedVersion = "dataLinkVersion"
	var expectedSeq = 37
	var testHandle = "handleArg"
	var testDatalinkArg = fmt.Sprintf("%s/%s:%s-rc-%d",
		expectedOem, expectedHandle, expectedVersion, expectedSeq)

	assert.NoError(t, testExecutor.Add(testHandle, testDatalinkArg))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+testHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`datalink-oem:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(`datalink-handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`datalink-version:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`datalink-seq:[\s\t]*`+cast.ToString(expectedSeq)), actual)
}

func TestStoreLockerExecutor_Display_UpdateSecret(t *testing.T) {
	var expectedSecret = memguard.NewEnclave([]byte("secret"))
	var expectedHandle = "handleArg"
	var expectedKey = "key"
	var testOutput bytes.Buffer
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(&testOutput)).
		Build()
	var testOptions = NewStoreAddOptions()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			LockerOptions: testOptions,

			handleArg: expectedHandle,
			keyArg:    expectedKey,
			secretArg: expectedSecret,
		},
	}

	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`prop-key:[\s\t]*`+expectedKey), actual)
	assert.Regexp(t, regexp.MustCompile(`prop-value:[\s\t]*\*+\n`), actual)
}

func TestStoreLockerExecutor_Pretend(t *testing.T) {
	var captureStoreAdd locker.StoreAddParams
	var captureExport broker.DataLinkParams
	var expectedLockerHandle = "storeHandle"
	var expectedOem = "oem"
	var expectedHandle = "value"
	var expectedVersion = "version"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewStoreAddOptions()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			LockerOptions: testOptions,
			handleArg:     expectedLockerHandle,
			addOem:        expectedOem,
			addHandle:     expectedHandle,
			addVersion:    expectedVersion,

			accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
			dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
				return nil
			},
			exportLinkTaskFactory: newExportLinkTaskPretendStub(&captureExport),
		},

		storeAddTaskFactory: newStoreAddTaskPretendStub(&captureStoreAdd),
	}

	t.Setenv(passphraseEnvKey, "myPass")
	testExecutor.Pretend()
	assert.NotNil(t, captureExport)
	assert.Equal(t, expectedOem, captureExport.Oem)
	assert.Equal(t, expectedHandle, captureExport.Handle)
	assert.Equal(t, expectedVersion, captureExport.Version)
	assert.NotNil(t, captureStoreAdd)
	assert.Equal(t, expectedLockerHandle, captureStoreAdd.LockerHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureStoreAdd.LockerPath)
}

func TestStoreLockerExecutor_Pretend_Update(t *testing.T) {
	var captureStoreFind locker.StoreFindParams
	var captureStoreUpdate locker.StoreUpdateParams
	var captureCollect broker.DataLinkParams
	var expectedHandle = "storeHandle"
	var expectedKey = "key"
	var expectedValue = "value"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewStoreUpdateOptions()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			LockerOptions: testOptions,
			handleArg:     expectedHandle,
			keyArg:        expectedKey,
			valueArg:      expectedValue,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			collectLinkTaskFactory: newCollectLinkTaskPretendStub(&captureCollect),
			dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
				return nil
			},
		},

		storeFindTaskFactory:   newStoreFindTaskPretendStub(&captureStoreFind),
		storeUpdateTaskFactory: newStoreUpdateTaskPretendStub(&captureStoreUpdate),
	}

	testExecutor.Pretend()
	assert.NotNil(t, captureStoreFind)
	assert.Equal(t, expectedHandle, captureStoreFind.LockerHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureStoreFind.LockerPath)
	assert.NotEmpty(t, captureCollect.Oem)
	assert.NotEmpty(t, captureCollect.Handle)
	assert.NotEmpty(t, captureCollect.Version)
	assert.Equal(t, expectedKey, captureStoreUpdate.Key)
	assert.Equal(t, expectedValue, captureStoreUpdate.Value)
}

func TestStoreLockerExecutor_Proceed(t *testing.T) {
	var captureStoreAdd locker.StoreAddParams
	var captureExport broker.DataLinkParams
	var expectedStoreHandle = "storeHandle"
	var expectedOem = "oem"
	var expectedHandle = "value"
	var expectedVersion = "version"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewStoreAddOptions()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			LockerOptions: testOptions,
			handleArg:     expectedStoreHandle,
			addOem:        expectedOem,
			addHandle:     expectedHandle,
			addVersion:    expectedVersion,

			accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
			dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
				return nil
			},
			exportLinkTaskFactory: newExportLinkTaskProceedStub(&captureExport),
		},

		storeAddTaskFactory: newStoreAddTaskProceedStub(&captureStoreAdd),
	}

	t.Setenv(passphraseEnvKey, "myPass")
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotNil(t, captureExport)
	assert.Equal(t, expectedOem, captureExport.Oem)
	assert.Equal(t, expectedHandle, captureExport.Handle)
	assert.Equal(t, expectedVersion, captureExport.Version)
	assert.NotNil(t, captureStoreAdd)
	assert.Equal(t, expectedStoreHandle, captureStoreAdd.LockerHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureStoreAdd.LockerPath)
}

func TestStoreLockerExecutor_Proceed_Secret(t *testing.T) {
	var captureStoreFind locker.StoreFindParams
	var captureStoreUpdate locker.StoreUpdateParams
	var captureCollect broker.DataLinkParams
	var expectedHandle = "storeHandle"
	var expectedKey = "key"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readFactoryPassword("myPass")).
		Build()
	var testOptions = NewStoreUpdateOptions()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			LockerOptions: testOptions,
			handleArg:     expectedHandle,
			keyArg:        expectedKey,
			secretArg:     memguard.NewEnclave([]byte("test")),

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			collectLinkTaskFactory: newCollectLinkTaskProceedStub(&captureCollect),
			dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
				return nil
			},
		},

		storeFindTaskFactory:   newStoreFindTaskProceedStub(&captureStoreFind),
		storeUpdateTaskFactory: newStoreUpdateTaskProceedStub(&captureStoreUpdate),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotNil(t, captureStoreFind)
	assert.Equal(t, expectedHandle, captureStoreFind.LockerHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureStoreFind.LockerPath)
	assert.NotEmpty(t, captureCollect.Oem)
	assert.NotEmpty(t, captureCollect.Handle)
	assert.NotEmpty(t, captureCollect.Version)
	assert.Equal(t, expectedKey, captureStoreUpdate.Key)
	assert.Empty(t, captureStoreUpdate.Value)
	assert.Same(t, testExecutor.secretArg, captureStoreUpdate.Secret)
}

func TestStoreLockerExecutor_Proceed_Update(t *testing.T) {
	var captureStoreFind locker.StoreFindParams
	var captureStoreUpdate locker.StoreUpdateParams
	var captureCollect broker.DataLinkParams
	var expectedHandle = "storeHandle"
	var expectedKey = "key"
	var expectedValue = "value"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewStoreUpdateOptions()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			LockerOptions: testOptions,
			handleArg:     expectedHandle,
			keyArg:        expectedKey,
			valueArg:      expectedValue,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			collectLinkTaskFactory: newCollectLinkTaskProceedStub(&captureCollect),
			dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
				return nil
			},
		},

		storeFindTaskFactory:   newStoreFindTaskProceedStub(&captureStoreFind),
		storeUpdateTaskFactory: newStoreUpdateTaskProceedStub(&captureStoreUpdate),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotNil(t, captureStoreFind)
	assert.Equal(t, expectedHandle, captureStoreFind.LockerHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureStoreFind.LockerPath)
	assert.NotEmpty(t, captureCollect.Oem)
	assert.NotEmpty(t, captureCollect.Handle)
	assert.NotEmpty(t, captureCollect.Version)
	assert.Equal(t, expectedKey, captureStoreUpdate.Key)
	assert.Equal(t, expectedValue, captureStoreUpdate.Value)
}

func TestStoreLockerExecutor_Update(t *testing.T) {
	var testOutput bytes.Buffer
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(&testOutput)).
		Build()
	var testOptions = NewStoreUpdateOptions()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			LockerOptions: testOptions,
		},
	}
	var expectedKey = "key"
	var expectedValue = "value"
	var testHandle = "handleArg"

	assert.NoError(t, testExecutor.Update(testHandle, expectedKey, expectedValue))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+testHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`prop-key:[\s\t]*`+expectedKey), actual)
	assert.Regexp(t, regexp.MustCompile(`prop-value:[\s\t]*`+expectedValue), actual)
}

func TestStoreLockerExecutor_Update_PipeError(t *testing.T) {
	var expectedError = errors.New("error")
	var testLedger = config.NewBuilder().
		WithInput(&gio.StubReader{ReadError: expectedError}).
		Build()
	var testExecutor = &StoreLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
		},
	}
	var expectedKey = "key"
	var expectedValue = "value"
	var testHandle = "handleArg"

	assert.ErrorIs(t, testExecutor.Update(testHandle, expectedKey, expectedValue), expectedError)
}

func TestNewStore(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCmd = NewStore(testLedger, &Cli{})

	assert.Equal(t, 3, len(testCmd.Commands()))
}

func TestNewStoreLockerExecutor_Add(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCmd = &cobra.Command{}
	var testFactory = newStoreAddExecutorFactory(testLedger, &Cli{}, NewStoreAddOptions())

	assert.NotNil(t, testFactory(testCmd))
}

func TestNewStoreLockerExecutor_Update(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCmd = &cobra.Command{}
	var testFactory = newStoreUpdateExecutorFactory(testLedger, &Cli{}, NewStoreUpdateOptions())

	assert.NotNil(t, testFactory(testCmd))
}

func newStoreAddTaskPretendStub(captured *locker.StoreAddParams) StoreAddTaskFactory {
	return func() *task.Task[locker.StoreAddParams] {
		return &task.Task[locker.StoreAddParams]{
			OnPrepare: func(params *locker.StoreAddParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *locker.StoreAddParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newStoreAddTaskProceedStub(captured *locker.StoreAddParams) StoreAddTaskFactory {
	return func() *task.Task[locker.StoreAddParams] {
		return &task.Task[locker.StoreAddParams]{
			OnPrepare: func(params *locker.StoreAddParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *locker.StoreAddParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newStoreFindTaskPretendStub(captured *locker.StoreFindParams) StoreFindTaskFactory {
	return func() *task.Task[locker.StoreFindParams] {
		return &task.Task[locker.StoreFindParams]{
			OnPrepare: func(params *locker.StoreFindParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *locker.StoreFindParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newStoreFindTaskProceedStub(captured *locker.StoreFindParams) StoreFindTaskFactory {
	return func() *task.Task[locker.StoreFindParams] {
		return &task.Task[locker.StoreFindParams]{
			OnPrepare: func(params *locker.StoreFindParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *locker.StoreFindParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newStorePublishTaskPretendStub(captured *locker.StorePublishParams) StorePublishTaskFactory {
	return func() *task.Task[locker.StorePublishParams] {
		return &task.Task[locker.StorePublishParams]{
			OnPrepare: func(params *locker.StorePublishParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *locker.StorePublishParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newStorePublishTaskProceedStub(captured *locker.StorePublishParams) StorePublishTaskFactory {
	return func() *task.Task[locker.StorePublishParams] {
		return &task.Task[locker.StorePublishParams]{
			OnPrepare: func(params *locker.StorePublishParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *locker.StorePublishParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newStoreSyncTaskPretendStub(captured *locker.StoreFindParams) StoreSyncTaskFactory {
	return func() *task.Task[locker.StoreFindParams] {
		return &task.Task[locker.StoreFindParams]{
			OnPrepare: func(params *locker.StoreFindParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *locker.StoreFindParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newStoreSyncTaskProceedStub(captured *locker.StoreFindParams) StoreSyncTaskFactory {
	return func() *task.Task[locker.StoreFindParams] {
		return &task.Task[locker.StoreFindParams]{
			OnPrepare: func(params *locker.StoreFindParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *locker.StoreFindParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newStoreUpdateTaskPretendStub(captured *locker.StoreUpdateParams) StoreUpdateTaskFactory {
	return func() *task.Task[locker.StoreUpdateParams] {
		return &task.Task[locker.StoreUpdateParams]{
			OnPrepare: func(params *locker.StoreUpdateParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *locker.StoreUpdateParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newStoreUpdateTaskProceedStub(captured *locker.StoreUpdateParams) StoreUpdateTaskFactory {
	return func() *task.Task[locker.StoreUpdateParams] {
		return &task.Task[locker.StoreUpdateParams]{
			OnPrepare: func(params *locker.StoreUpdateParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *locker.StoreUpdateParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}
