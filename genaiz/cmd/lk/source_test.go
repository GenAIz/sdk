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

func TestSourceExecutor_Add(t *testing.T) {
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
	var testOptions = NewSourceAddOptions()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SourceOptions: testOptions,
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

func TestSourceExecutor_Add_EmptyOem(t *testing.T) {
	var testExecutor = &SourceExecutor{}
	var testHandle = "handleArg"
	// parsing assumes single atoms to be handles
	var testDatalinkArg = "dataLinkArg"

	assert.ErrorIs(t, testExecutor.Add(testHandle, testDatalinkArg), errorDataLinkOemRequired)
}

func TestSourceExecutor_Add_EmptyVersion(t *testing.T) {
	var testExecutor = &SourceExecutor{}
	var testHandle = "handleArg"
	// no support for default versioning
	var testDatalinkArg = "dataLinkOem/dataLinkHandle"

	assert.ErrorIs(t, testExecutor.Add(testHandle, testDatalinkArg), errorDataLinkVersionRequired)
}

func TestSourceExecutor_Add_InvalidSeq(t *testing.T) {
	var testExecutor = &SourceExecutor{}
	var testHandle = "handleArg"
	// no support for default versioning
	var testDatalinkArg = "dataLinkOem/dataLinkHandle:ver-rc-notSequence"

	assert.ErrorIs(t, testExecutor.Add(testHandle, testDatalinkArg), errorDataLinkSequenceInvalid)
}

func TestSourceExecutor_Add_WithSequence(t *testing.T) {
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
	var testOptions = NewSourceAddOptions()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SourceOptions: testOptions,
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

func TestSourceExecutor_Display_UpdateSecret(t *testing.T) {
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
	var testOptions = NewSourceAddOptions()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SourceOptions: testOptions,

		handleArg: expectedHandle,
		keyArg:    expectedKey,
		secretArg: expectedSecret,
	}

	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`prop-key:[\s\t]*`+expectedKey), actual)
	assert.Regexp(t, regexp.MustCompile(`prop-value:[\s\t]*\*+\n`), actual)
}

func TestSourceExecutor_Pretend(t *testing.T) {
	var captureSourceAdd locker.SourceAddParams
	var captureExport broker.DataLinkParams
	var expectedSourceHandle = "sourceHandle"
	var expectedOem = "oem"
	var expectedHandle = "value"
	var expectedVersion = "version"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewSourceAddOptions()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SourceOptions: testOptions,
		handleArg:     expectedSourceHandle,
		addOem:        expectedOem,
		addHandle:     expectedHandle,
		addVersion:    expectedVersion,

		accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
		dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
			return nil
		},
		exportLinkTaskFactory: newExportLinkTaskPretendStub(&captureExport),
		sourceAddTaskFactory:  newSourceAddTaskPretendStub(&captureSourceAdd),
	}

	t.Setenv(passphraseEnvKey, "myPass")
	testExecutor.Pretend()
	assert.NotNil(t, captureExport)
	assert.Equal(t, expectedOem, captureExport.Oem)
	assert.Equal(t, expectedHandle, captureExport.Handle)
	assert.Equal(t, expectedVersion, captureExport.Version)
	assert.NotNil(t, captureSourceAdd)
	assert.Equal(t, expectedSourceHandle, captureSourceAdd.SourceHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureSourceAdd.LockerPath)
}

func TestSourceExecutor_Pretend_Update(t *testing.T) {
	var captureSourceFind locker.SourceFindParams
	var captureSourceUpdate locker.SourceUpdateParams
	var captureCollect broker.DataLinkParams
	var expectedHandle = "sourceHandle"
	var expectedKey = "key"
	var expectedValue = "value"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewSourceUpdateOptions()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SourceOptions: testOptions,
		handleArg:     expectedHandle,
		keyArg:        expectedKey,
		valueArg:      expectedValue,

		accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
		dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
			return nil
		},
		sourceFindTaskFactory:   newSourceFindTaskPretendStub(&captureSourceFind),
		collectLinkTaskFactory:  newCollectLinkTaskPretendStub(&captureCollect),
		sourceUpdateTaskFactory: newSourceUpdateTaskPretendStub(&captureSourceUpdate),
	}

	testExecutor.Pretend()
	assert.NotNil(t, captureSourceFind)
	assert.Equal(t, expectedHandle, captureSourceFind.SourceHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureSourceFind.LockerPath)
	assert.NotEmpty(t, captureCollect.Oem)
	assert.NotEmpty(t, captureCollect.Handle)
	assert.NotEmpty(t, captureCollect.Version)
	assert.Equal(t, expectedKey, captureSourceUpdate.Key)
	assert.Equal(t, expectedValue, captureSourceUpdate.Value)
}

func TestSourceExecutor_Proceed(t *testing.T) {
	var captureSourceAdd locker.SourceAddParams
	var captureExport broker.DataLinkParams
	var expectedSourceHandle = "sourceHandle"
	var expectedOem = "oem"
	var expectedHandle = "value"
	var expectedVersion = "version"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewSourceAddOptions()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SourceOptions: testOptions,
		handleArg:     expectedSourceHandle,
		addOem:        expectedOem,
		addHandle:     expectedHandle,
		addVersion:    expectedVersion,

		accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
		dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
			return nil
		},
		exportLinkTaskFactory: newExportLinkTaskProceedStub(&captureExport),
		sourceAddTaskFactory:  newSourceAddTaskProceedStub(&captureSourceAdd),
	}

	t.Setenv(passphraseEnvKey, "myPass")
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotNil(t, captureExport)
	assert.Equal(t, expectedOem, captureExport.Oem)
	assert.Equal(t, expectedHandle, captureExport.Handle)
	assert.Equal(t, expectedVersion, captureExport.Version)
	assert.NotNil(t, captureSourceAdd)
	assert.Equal(t, expectedSourceHandle, captureSourceAdd.SourceHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureSourceAdd.LockerPath)
}

func TestSourceExecutor_Proceed_Secret(t *testing.T) {
	var captureSourceFind locker.SourceFindParams
	var captureSourceUpdate locker.SourceUpdateParams
	var captureCollect broker.DataLinkParams
	var expectedHandle = "sourceHandle"
	var expectedKey = "key"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readFactoryPassword("myPass")).
		Build()
	var testOptions = NewSourceUpdateOptions()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SourceOptions: testOptions,
		handleArg:     expectedHandle,
		keyArg:        expectedKey,
		secretArg:     memguard.NewEnclave([]byte("test")),

		accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
		dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
			return nil
		},
		sourceFindTaskFactory:   newSourceFindTaskProceedStub(&captureSourceFind),
		collectLinkTaskFactory:  newCollectLinkTaskProceedStub(&captureCollect),
		sourceUpdateTaskFactory: newSourceUpdateTaskProceedStub(&captureSourceUpdate),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotNil(t, captureSourceFind)
	assert.Equal(t, expectedHandle, captureSourceFind.SourceHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureSourceFind.LockerPath)
	assert.NotEmpty(t, captureCollect.Oem)
	assert.NotEmpty(t, captureCollect.Handle)
	assert.NotEmpty(t, captureCollect.Version)
	assert.Equal(t, expectedKey, captureSourceUpdate.Key)
	assert.Empty(t, captureSourceUpdate.Value)
	assert.Same(t, testExecutor.secretArg, captureSourceUpdate.Secret)
}

func TestSourceExecutor_Proceed_Update(t *testing.T) {
	var captureSourceFind locker.SourceFindParams
	var captureSourceUpdate locker.SourceUpdateParams
	var captureCollect broker.DataLinkParams
	var expectedHandle = "sourceHandle"
	var expectedKey = "key"
	var expectedValue = "value"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(t.TempDir()).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewSourceUpdateOptions()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SourceOptions: testOptions,
		handleArg:     expectedHandle,
		keyArg:        expectedKey,
		valueArg:      expectedValue,

		accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
		dataLinksWriterFactory: func(ledger *config.Ledger, s string) *dk.DataLinksWriter {
			return nil
		},
		sourceFindTaskFactory:   newSourceFindTaskProceedStub(&captureSourceFind),
		collectLinkTaskFactory:  newCollectLinkTaskProceedStub(&captureCollect),
		sourceUpdateTaskFactory: newSourceUpdateTaskProceedStub(&captureSourceUpdate),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotNil(t, captureSourceFind)
	assert.Equal(t, expectedHandle, captureSourceFind.SourceHandle)
	assert.Equal(t, filepath.Join(testLedger.UserPath, "locker.bin"), captureSourceFind.LockerPath)
	assert.NotEmpty(t, captureCollect.Oem)
	assert.NotEmpty(t, captureCollect.Handle)
	assert.NotEmpty(t, captureCollect.Version)
	assert.Equal(t, expectedKey, captureSourceUpdate.Key)
	assert.Equal(t, expectedValue, captureSourceUpdate.Value)
}

func TestSourceExecutor_Update(t *testing.T) {
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
	var testOptions = NewSourceUpdateOptions()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SourceOptions: testOptions,
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

func TestSourceExecutor_Update_PipeError(t *testing.T) {
	var expectedError = errors.New("error")
	var testLedger = config.NewBuilder().
		WithInput(&gio.StubReader{ReadError: expectedError}).
		Build()
	var testExecutor = &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
	}
	var expectedKey = "key"
	var expectedValue = "value"
	var testHandle = "handleArg"

	assert.ErrorIs(t, testExecutor.Update(testHandle, expectedKey, expectedValue), expectedError)
}

func TestNewSource(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCmd = NewSource(testLedger, &Cli{})

	assert.Equal(t, 2, len(testCmd.Commands()))
}

func TestNewSourceExecutor_Add(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCmd = &cobra.Command{}
	var testFactory = newSourceAddExecutorFactory(testLedger, &Cli{}, NewSourceAddOptions())

	assert.NotNil(t, testFactory(testCmd))
}

func TestNewSourceExecutor_Update(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCmd = &cobra.Command{}
	var testFactory = newSourceUpdateExecutorFactory(testLedger, &Cli{}, NewSourceUpdateOptions())

	assert.NotNil(t, testFactory(testCmd))
}

func newCollectLinkTaskPretendStub(captured *broker.DataLinkParams) dk.CollectLinkTaskFactory {
	return func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataLinkParams, state *task.State) error {
				params.Oem = "oem"
				params.Handle = "handle"
				params.Version = "version"
				*captured = *params
				return nil
			},
		}
	}
}

func newExportLinkTaskPretendStub(captured *broker.DataLinkParams) dk.ExportLinkTaskFactory {
	return func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataLinkParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newSourceAddTaskPretendStub(captured *locker.SourceAddParams) SourceAddTaskFactory {
	return func() *task.Task[locker.SourceAddParams] {
		return &task.Task[locker.SourceAddParams]{
			OnPrepare: func(params *locker.SourceAddParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *locker.SourceAddParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newSourceFindTaskPretendStub(captured *locker.SourceFindParams) SourceFindTaskFactory {
	return func() *task.Task[locker.SourceFindParams] {
		return &task.Task[locker.SourceFindParams]{
			OnPrepare: func(params *locker.SourceFindParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *locker.SourceFindParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newSourceUpdateTaskPretendStub(captured *locker.SourceUpdateParams) SourceUpdateTaskFactory {
	return func() *task.Task[locker.SourceUpdateParams] {
		return &task.Task[locker.SourceUpdateParams]{
			OnPrepare: func(params *locker.SourceUpdateParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *locker.SourceUpdateParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newCollectLinkTaskProceedStub(captured *broker.DataLinkParams) dk.CollectLinkTaskFactory {
	return func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				params.DataLink = &broker.DataLink{
					Oem:     "oem",
					Handle:  "handle",
					Version: "version",
				}
				*captured = *params
				return nil
			},
		}
	}
}

func newExportLinkTaskProceedStub(captured *broker.DataLinkParams) dk.ExportLinkTaskFactory {
	return func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newSourceAddTaskProceedStub(captured *locker.SourceAddParams) SourceAddTaskFactory {
	return func() *task.Task[locker.SourceAddParams] {
		return &task.Task[locker.SourceAddParams]{
			OnPrepare: func(params *locker.SourceAddParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *locker.SourceAddParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newSourceFindTaskProceedStub(captured *locker.SourceFindParams) SourceFindTaskFactory {
	return func() *task.Task[locker.SourceFindParams] {
		return &task.Task[locker.SourceFindParams]{
			OnPrepare: func(params *locker.SourceFindParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *locker.SourceFindParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newSourceUpdateTaskProceedStub(captured *locker.SourceUpdateParams) SourceUpdateTaskFactory {
	return func() *task.Task[locker.SourceUpdateParams] {
		return &task.Task[locker.SourceUpdateParams]{
			OnPrepare: func(params *locker.SourceUpdateParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *locker.SourceUpdateParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}
