package lk

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/locker"
)

func TestInitExecutor_Display(t *testing.T) {
	var expectedPath = t.TempDir()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewInitOptions()
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions: testOptions,
		path:        expectedPath,
	}

	testViper.Set(testOptions.optionOverwrite.Key, true)
	testViper.Set(testOptions.optionUpdate.Key, true)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionPath.Param+`:[\s\t]*`+expectedPath), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOverwrite.Param+`:[\s\t]*true`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionUpdate.Param+`:[\s\t]*true`), actual)
}

func TestInitExecutor_Pretend(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedFile = filepath.Join(expectedPath, ".config", "genaiz", "locker.bin")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var captureParams locker.InitParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(expectedPath).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewInitOptions()
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions:           testOptions,
		initLockerTaskFactory: newInitLockerTaskPretendCapture(&captureParams),
	}

	defer patch.Unpatch()

	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.Equal(t, expectedFile, captureParams.LockerPath)
	assert.False(t, captureParams.Update)
	assert.False(t, patch.Called)
}

func TestInitExecutor_Pretend_PermissionError(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedFile = filepath.Join(expectedPath, "noAllowed")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var captureParams locker.InitParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewInitOptions()
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions:           testOptions,
		path:                  expectedFile,
		initLockerTaskFactory: newInitLockerTaskPretendCapture(&captureParams),
	}
	var err error

	defer patch.Unpatch()

	if _, err = os.Create(expectedFile); err == nil {
		if err = os.Chmod(expectedFile, 0200); err == nil {
			testViper.Set(testOptions.optionOverwrite.Key, true)
			testLedger.InitLogging()
			testExecutor.Pretend()
			assert.True(t, patch.Called)
			assert.EqualValues(t, 1, patch.CalledWith)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func TestInitExecutor_Proceed(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedFile = filepath.Join(expectedPath, "myLocker")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var captureParams locker.InitParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewInitOptions()
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions:           testOptions,
		path:                  expectedFile,
		initLockerTaskFactory: newInitLockerTaskCompleteCapture(&captureParams),
	}

	defer patch.Unpatch()
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedFile, captureParams.LockerPath)
	assert.False(t, captureParams.Update)
	assert.False(t, patch.Called)
}

func TestInitExecutor_Proceed_ConfirmNothing(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedFile = filepath.Join(expectedPath, "myLocker")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var captureParams locker.InitParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = &InitOptions{
		optionOverwrite: cli.Options.Lockers.Overwrite().BuildBoolOption(),
		optionPath: cli.Options.Lockers.Path().
			WithKeys(&schema.Keys{Doc: "test"}).
			BuildStringOption(),
		optionUpdate: cli.Options.Lockers.Update().BuildBoolOption(),
		dialogYes: func(writer io.Writer, reader io.Reader, s string) bool {
			return false
		},
	}
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions:           testOptions,
		path:                  expectedFile,
		initLockerTaskFactory: newInitLockerTaskCompleteCapture(&captureParams),
	}
	var fd *os.File
	var err error

	defer patch.Unpatch()

	if fd, err = os.Create(expectedFile); err == nil {
		defer filez.CloseSilently(fd)
		testViper.Set(testOptions.optionOverwrite.Key, false)
		testViper.Set(testOptions.optionUpdate.Key, false)
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.True(t, patch.Called)
		assert.EqualValues(t, 1, patch.CalledWith)
		return
	}

	assert.Fail(t, err.Error())
}

func TestInitExecutor_Proceed_ConfirmOverwrite(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedFile = filepath.Join(expectedPath, "myLocker")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var captureParams locker.InitParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = &InitOptions{
		optionOverwrite: cli.Options.Lockers.Overwrite().BuildBoolOption(),
		optionPath:      cli.Options.Lockers.Path().WithKeys(&schema.Keys{Doc: "test"}).BuildStringOption(),
		optionUpdate:    cli.Options.Lockers.Update().BuildBoolOption(),
		dialogYes: func(writer io.Writer, reader io.Reader, s string) bool {
			return strings.Contains(s, "Overwrite ")
		},
	}
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions:           testOptions,
		path:                  expectedFile,
		initLockerTaskFactory: newInitLockerTaskCompleteCapture(&captureParams),
	}
	var fd *os.File
	var err error

	defer patch.Unpatch()

	if fd, err = os.Create(expectedFile); err == nil {
		defer filez.CloseSilently(fd)
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.Equal(t, expectedFile, captureParams.LockerPath)
		assert.False(t, captureParams.Update)
		assert.False(t, patch.Called)
		return
	}

	assert.Fail(t, err.Error())
}

func TestInitExecutor_Proceed_ConfirmUpdate(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedFile = filepath.Join(expectedPath, "myLocker")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var captureParams locker.InitParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = &InitOptions{
		optionOverwrite: cli.Options.Lockers.Overwrite().BuildBoolOption(),
		optionPath: cli.Options.Lockers.Path().
			WithKeys(&schema.Keys{Doc: "test"}).
			BuildStringOption(),
		optionUpdate: cli.Options.Lockers.Update().BuildBoolOption(),
		dialogYes: func(writer io.Writer, reader io.Reader, s string) bool {
			return strings.Contains(s, "Update ")
		},
	}
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions:           testOptions,
		path:                  expectedFile,
		initLockerTaskFactory: newInitLockerTaskCompleteCapture(&captureParams),
	}
	var fd *os.File
	var err error

	defer patch.Unpatch()

	if fd, err = os.Create(expectedFile); err == nil {
		defer filez.CloseSilently(fd)
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.Equal(t, expectedFile, captureParams.LockerPath)
		assert.True(t, captureParams.Update)
		assert.False(t, patch.Called)
		return
	}

	assert.Fail(t, err.Error())
}

func TestInitExecutor_Proceed_Overwrite(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedFile = filepath.Join(expectedPath, "myLocker")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var captureParams locker.InitParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewInitOptions()
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions:           testOptions,
		path:                  expectedFile,
		initLockerTaskFactory: newInitLockerTaskCompleteCapture(&captureParams),
	}
	var fd *os.File
	var err error

	defer patch.Unpatch()

	if fd, err = os.Create(expectedFile); err == nil {
		defer filez.CloseSilently(fd)
		testViper.Set(testOptions.optionOverwrite.Key, true)
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.Equal(t, expectedFile, captureParams.LockerPath)
		assert.False(t, captureParams.Update)
		assert.False(t, patch.Called)
		return
	}

	assert.Fail(t, err.Error())
}

func TestInitExecutor_Proceed_Update(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedFile = filepath.Join(expectedPath, "myLocker")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var captureParams locker.InitParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithSecretHandler(readFactoryPassword("myPass")).
		Build()
	var testOptions = NewInitOptions()
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions:           testOptions,
		path:                  expectedFile,
		initLockerTaskFactory: newInitLockerTaskCompleteCapture(&captureParams),
	}
	var fd *os.File
	var err error

	defer patch.Unpatch()

	if fd, err = os.Create(expectedFile); err == nil {
		defer filez.CloseSilently(fd)
		testViper.Set(testOptions.optionUpdate.Key, true)
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.Equal(t, expectedFile, captureParams.LockerPath)
		assert.True(t, captureParams.Update)
		assert.False(t, patch.Called)
		return
	}

	assert.Fail(t, err.Error())
}

func TestInitExecutor_Proceed_UpdateEnvPassword(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedFile = filepath.Join(expectedPath, "myLocker")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var captureParams locker.InitParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithSecretHandler(readEmptyPassword).
		Build()
	var testOptions = NewInitOptions()
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions:           testOptions,
		path:                  expectedFile,
		initLockerTaskFactory: newInitLockerTaskCompleteCapture(&captureParams),
	}
	var fd *os.File
	var err error

	defer patch.Unpatch()

	if fd, err = os.Create(expectedFile); err == nil {
		defer filez.CloseSilently(fd)

		t.Setenv(passphraseEnvKey, "my Pass")
		testViper.Set(testOptions.optionUpdate.Key, true)
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.Equal(t, expectedFile, captureParams.LockerPath)
		assert.True(t, captureParams.Update)
		assert.False(t, patch.Called)
		return
	}

	assert.Fail(t, err.Error())
}

func TestNewInit(t *testing.T) {
	var expectedPath = t.TempDir()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewInitOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = NewInit(testLedger, testCli)

	testExecutor.Run(&cobra.Command{}, []string{expectedPath})
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionPath.Param+`:[\s\t]*`+expectedPath), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOverwrite.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionUpdate.Param+`:[\s\t]*false`), actual)
}

func newInitLockerTaskPretendCapture(capture *locker.InitParams) initLockerTaskFactory {
	return func() *task.Task[locker.InitParams] {
		return &task.Task[locker.InitParams]{
			OnPrepare: func(params *locker.InitParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *locker.InitParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newInitLockerTaskCompleteCapture(capture *locker.InitParams) initLockerTaskFactory {
	return func() *task.Task[locker.InitParams] {
		return &task.Task[locker.InitParams]{
			OnPrepare: func(params *locker.InitParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *locker.InitParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func readEmptyPassword() ([]byte, error) {
	return []byte{}, nil
}

func readFactoryPassword(password string) func() ([]byte, error) {
	return func() ([]byte, error) {
		return []byte(password), nil
	}
}
