package sf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
)

func TestStoreExecutor_Add(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = NewStoreAddOptions(testCli)
	var testCmd = &cobra.Command{}
	var testExecutor = newStoreAddFactory(testLedger, testCli, testOptions)(testCmd)
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "0.0.1"
	var expectedLink = fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion)

	testViper.Set(schema.Genaiz.Function.Publish.Type.Doc, layout.FunctionTypeConnector)
	assert.NoError(t, testExecutor.Add(expectedLink))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionNoValidation.Param+`:[\s\t]*false`), actual)
}

func TestStoreExecutor_Add_Duplicate(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = NewStoreAddOptions(testCli)
	var testExecutor = NewStoreExecutor(t.Context(), testLedger, testCli, testOptions)
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "0.0.1"
	var expectedLink = fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion)

	testViper.Set(schema.Genaiz.Function.Publish.Type.Doc, layout.FunctionTypeConnector)
	testViper.Set(testExecutor.innerStores.Key, []string{expectedLink})
	assert.Error(t, testExecutor.Add(expectedLink))
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestStoreExecutor_Add_InvalidOem(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = NewStoreAddOptions(testCli)
	var testCmd = &cobra.Command{}
	var testExecutor = newStoreAddFactory(testLedger, testCli, testOptions)(testCmd)
	var expectedHandle = "expected-handle"
	var expectedVersion = "0.0.1"
	var expectedLink = fmt.Sprintf("%s:%s", expectedHandle, expectedVersion)

	defer patch.Unpatch()
	testViper.Set(schema.Genaiz.Function.Publish.Type.Doc, layout.FunctionTypeConnector)
	assert.NoError(t, testExecutor.Add(expectedLink))
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestStoreExecutor_Add_InvalidType(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = NewStoreAddOptions(testCli)
	var testCmd = &cobra.Command{}
	var testExecutor = newStoreAddFactory(testLedger, testCli, testOptions)(testCmd)
	var expectedHandle = "expected-handle"
	var expectedVersion = "0.0.1"
	var expectedLink = fmt.Sprintf("%s:%s", expectedHandle, expectedVersion)

	testViper.Set(schema.Genaiz.Function.Publish.Type.Doc, layout.FunctionTypeFunction)
	assert.ErrorIs(t, testExecutor.Add(expectedLink), errInvalidConnectorType)
}

func TestStoreExecutor_Pretend_Add(t *testing.T) {
	var calledInit, calledLinks bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewStoreAddOptions(testCli)
	var testExecutor = &StoreExecutor{
		DataLinkExecutor: DataLinkExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			DataLinkOptions: testOptions,
		},

		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		listLinksTaskFactory: newListLinksTaskPretendStub(&calledLinks),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.addParams = &broker.DataLinkParams{}
		testExecutor.Pretend()
		assert.True(t, calledInit)
		assert.True(t, calledLinks)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestStoreExecutor_Pretend_Remove(t *testing.T) {
	var calledInit, calledLinks bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewStoreAddOptions(testCli)
	var testExecutor = &StoreExecutor{
		DataLinkExecutor: DataLinkExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			DataLinkOptions: testOptions,
		},

		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		listLinksTaskFactory: newListLinksTaskPretendStub(&calledLinks),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.rmParams = &broker.DataLinkParams{}
		testExecutor.Pretend()
		assert.True(t, calledInit)
		assert.False(t, calledLinks)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestStoreExecutor_Proceed_Add(t *testing.T) {
	var calledInit, calledLinks bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewStoreAddOptions(testCli)
	var expectedParams = &broker.DataLinkParams{
		DataLink: &broker.DataLink{
			Oem:     "expected.oem",
			Handle:  "expected-handle",
			Version: "1.0.9",
		},
	}
	var testExecutor = &StoreExecutor{
		DataLinkExecutor: DataLinkExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			DataLinkOptions: testOptions,
		},

		initTaskFactory: newInitTaskCompleteStub(func(actual *layout.InitParams) {
			calledInit = true
			assert.Equal(t, []string{expectedParams.ToString()}, actual.DataStores)
		}),
		listLinksTaskFactory: newListLinksTaskCompleteStub(func(actual *broker.DataLinkParams) {
			calledLinks = true
			assert.Equal(t, expectedParams.ToString(), actual.ToString())
		}),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testLedger.Logger = logrus.New()
		testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.addParams = expectedParams
		testExecutor.updatedStores = []string{expectedParams.ToString()}
		testExecutor.Proceed()
		assert.True(t, calledInit)
		assert.True(t, calledLinks)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestStoreExecutor_Proceed_Remove(t *testing.T) {
	var calledInit bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewStoreAddOptions(testCli)
	var expectedParams = &broker.DataLinkParams{
		DataLink: &broker.DataLink{
			Oem:     "expected.oem",
			Handle:  "expected-handle",
			Version: "1.0.9",
		},
	}
	var testExecutor = &StoreExecutor{
		DataLinkExecutor: DataLinkExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			DataLinkOptions: testOptions,
		},

		initTaskFactory: newInitTaskCompleteStub(func(actual *layout.InitParams) {
			calledInit = true
			assert.Equal(t, []string{expectedParams.ToString()}, actual.DataStores)
		}),
		listLinksTaskFactory: newListLinksTaskCompleteStub(func(actual *broker.DataLinkParams) {
			assert.Fail(t, "no list links call expected")
		}),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testLedger.Logger = logrus.New()
		testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.rmParams = expectedParams
		testExecutor.updatedStores = []string{expectedParams.ToString()}
		testExecutor.Proceed()
		assert.True(t, calledInit)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestStoreExecutor_Remove(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = NewStoreRemoveOptions(testCli)
	var testExecutor = NewStoreExecutor(t.Context(), testLedger, testCli, testOptions)
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "0.0.1"
	var expectedLink = fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion)

	testViper.Set(testExecutor.innerStores.Key, []string{"not.expected/handle:1.0.0", expectedLink})
	testViper.Set(testOptions.optionOem.Key, expectedOem)
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testExecutor.Remove("")
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`no-validation:[\s\t]*false`), actual)
}

func TestStoreExecutor_Remove_InvalidOem(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = NewStoreRemoveOptions(testCli)
	var testExecutor = NewStoreExecutor(t.Context(), testLedger, testCli, testOptions)
	var expectedHandle = "expected-handle"
	var expectedVersion = "0.0.1"

	defer patch.Unpatch()
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testExecutor.Remove("")
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestStoreExecutor_Remove_NotExisting(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = NewStoreRemoveOptions(testCli)
	var testCmd = &cobra.Command{}
	var testExecutor = newStoreRemoveFactory(testLedger, testCli, testOptions)(testCmd)
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "0.0.1"

	testViper.Set(testOptions.optionOem.Key, expectedOem)
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testExecutor.Remove("")
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`no-validation:[\s\t]*false`), actual)
}
