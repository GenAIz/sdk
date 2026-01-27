package dk

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type StubPropSpecOptions struct {
	buildError    error
	buildPropSpec *broker.PropSpec
	definers      []config.Definer
	isSecret      bool
}

func (s StubPropSpecOptions) Build(*config.Ledger, *broker.PropSpec) (*broker.PropSpec, error) {
	return s.buildPropSpec, s.buildError
}

func (s StubPropSpecOptions) Definers() []config.Definer {
	return s.definers
}

func (s StubPropSpecOptions) IsSecret(*config.Ledger) bool {
	return s.isSecret
}

func TestPropSpecExecutor_Add(t *testing.T) {
	var expectedKey = "expectedKey"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli: &Cli{
				BaseCli: cli.BaseCli{
					Dry: func(ledger *config.Ledger) bool {
						return true
					},
				},
			},
		},
		PropSpecOptions: &PropSpecOptions{
			PropSpecOptions: &StubPropSpecOptions{buildPropSpec: &broker.PropSpec{
				Key: expectedKey,
			}},
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.NoError(t, testExecutor.Add(testLink.Handle, expectedKey))
	actual := testOutput.String()
	assert.Contains(t, actual, expectedKey)
}

func TestPropSpecExecutor_Add_Secret(t *testing.T) {
	var expectedKey = "expectedKey"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli: &Cli{
				BaseCli: cli.BaseCli{
					Dry: func(ledger *config.Ledger) bool {
						return true
					},
				},
			},
		},
		PropSpecOptions: &PropSpecOptions{
			PropSpecOptions: &StubPropSpecOptions{
				buildPropSpec: &broker.PropSpec{
					Key: expectedKey,
				},
				isSecret: true,
			},
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.NoError(t, testExecutor.Add(testLink.Handle, expectedKey))
	actual := testOutput.String()
	assert.Contains(t, actual, expectedKey)
}

func TestPropSpecExecutor_Add_DataLinkBuildError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			PropSpecOptions: &StubPropSpecOptions{buildError: expectedError},
			BaseOptions:     makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.ErrorIs(t, testExecutor.Add(testLink.Handle, "key"), expectedError)
}

func TestPropSpecExecutor_Add_DataLinkDuplicated(t *testing.T) {
	var testKey = "testKey"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
		PropSpecs: []broker.PropSpec{
			{
				Key: testKey,
			},
		},
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.Error(t, testExecutor.Add(testLink.Handle, testKey))
}

func TestPropSpecExecutor_Add_DataLinkNotFound(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{}),
	}

	assert.Error(t, testExecutor.Add("testArg", "testKey"))
}

func TestPropSpecExecutor_Add_DataWriterError(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: makeTestBaseOptions(),
		},
	}

	testViper.Set(testExecutor.optionConfigType.Key, "notValid")
	assert.Error(t, testExecutor.Add("testArg", "testKey"))
}

func TestPropSpecExecutor_Display(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testExecutor = &PropSpecExecutor{}

	defer patch.Unpatch()
	testExecutor.Display()
	assert.NotEmpty(t, patch.CalledWith)
	params := cast.ToStringSlice(patch.CalledWith)
	assert.NotEmpty(t, params[0])
}

func TestPropSpecExecutor_Edit(t *testing.T) {
	var expectedKey = "expectedKey"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
		PropSpecs: []broker.PropSpec{
			{
				Key: expectedKey,
			},
		},
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli: &Cli{
				BaseCli: cli.BaseCli{
					Dry: func(ledger *config.Ledger) bool {
						return true
					},
				},
			},
		},
		PropSpecOptions: &PropSpecOptions{
			PropSpecOptions: &StubPropSpecOptions{buildPropSpec: &broker.PropSpec{
				Key: expectedKey,
			}},
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.NoError(t, testExecutor.Edit(testLink.Handle, expectedKey))
	actual := testOutput.String()
	assert.Contains(t, actual, expectedKey)
}

func TestPropSpecExecutor_Edit_DataLinkBuildError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testKey = "testKey"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
		PropSpecs: []broker.PropSpec{
			{
				Key: testKey,
			},
		},
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			PropSpecOptions: &StubPropSpecOptions{buildError: expectedError},
			BaseOptions:     makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.ErrorIs(t, testExecutor.Edit(testLink.Handle, testKey), expectedError)
}

func TestPropSpecExecutor_Edit_DataLinkNotFound(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{}),
	}

	assert.Error(t, testExecutor.Edit("testArg", "testKey"))
}

func TestPropSpecExecutor_Edit_DataWriterError(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: makeTestBaseOptions(),
		},
	}

	testViper.Set(testExecutor.optionConfigType.Key, "notValid")
	assert.Error(t, testExecutor.Edit("testArg", "testKey"))
}

func TestPropSpecExecutor_Edit_PropSpecNotFound(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli: &Cli{
				BaseCli: cli.BaseCli{
					Dry: func(ledger *config.Ledger) bool {
						return true
					},
				},
			},
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.Error(t, testExecutor.Edit(testLink.Handle, "testKey"))
}

func TestPropSpecExecutor_Edit_SecretSpec(t *testing.T) {
	var expectedKey = "expectedKey"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
		SecretSpecs: []broker.PropSpec{
			{
				Key: expectedKey,
			},
		},
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli: &Cli{
				BaseCli: cli.BaseCli{
					Dry: func(ledger *config.Ledger) bool {
						return true
					},
				},
			},
		},
		PropSpecOptions: &PropSpecOptions{
			PropSpecOptions: &StubPropSpecOptions{buildPropSpec: &broker.PropSpec{
				Key: expectedKey,
			}},
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.NoError(t, testExecutor.Edit(testLink.Handle, expectedKey))
	actual := testOutput.String()
	assert.Contains(t, actual, expectedKey)
}

func TestPropSpecExecutor_Edit_SecretSpecError(t *testing.T) {
	var expectedKey = "expectedKey"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
		SecretSpecs: []broker.PropSpec{
			{
				Key: expectedKey,
			},
		},
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli: &Cli{
				BaseCli: cli.BaseCli{
					Dry: func(ledger *config.Ledger) bool {
						return true
					},
				},
			},
		},
		PropSpecOptions: &PropSpecOptions{
			PropSpecOptions: &StubPropSpecOptions{buildPropSpec: &broker.PropSpec{
				Key:   expectedKey,
				Value: "someValue",
			}},
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.Error(t, testExecutor.Edit(testLink.Handle, expectedKey))
}

func TestPropSpecExecutor_Pretend(t *testing.T) {
	var calledParams broker.DataLinkParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: BaseOptions{
				optionConfigType:  &config.StringOption{Option: config.Option{Key: "configType"}},
				optionUserDefined: &config.BoolOption{Option: config.Option{Key: "userDefined"}},
			},
		},

		editedLink: &broker.DataLink{
			Handle:      "testHandle",
			Oem:         "testOem",
			Version:     "testVersion",
			Name:        "testName",
			Description: "testDescription",
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory(nil),
		editLinkTaskFactory:    newEditLinkPretendCapture(&calledParams),
	}

	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.Equal(t, testExecutor.editedLink.Description, calledParams.Description)
	assert.Equal(t, testExecutor.editedLink.Handle, calledParams.Handle)
	assert.Equal(t, testExecutor.editedLink.Name, calledParams.Name)
	assert.Equal(t, testExecutor.editedLink.Oem, calledParams.Oem)
	assert.Equal(t, testExecutor.editedLink.Version, calledParams.Version)
}

func TestPropSpecExecutor_Pretend_InvalidConfigType(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: BaseOptions{
				optionConfigType:  &config.StringOption{Option: config.Option{Key: "configType"}},
				optionUserDefined: &config.BoolOption{Option: config.Option{Key: "userDefined"}},
			},
		},
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, "notValid")
	testExecutor.Pretend()
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestPropSpecExecutor_Proceed(t *testing.T) {
	var calledParams broker.DataLinkParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: BaseOptions{
				optionConfigType:  &config.StringOption{Option: config.Option{Key: "configType"}},
				optionUserDefined: &config.BoolOption{Option: config.Option{Key: "userDefined"}},
			},
		},

		editedLink: &broker.DataLink{
			Handle:      "testHandle",
			Oem:         "testOem",
			Version:     "testVersion",
			Name:        "testName",
			Description: "testDescription",
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory(nil),
		editLinkTaskFactory:    newEditLinkCompleteCapture(&calledParams),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, testExecutor.editedLink.Description, calledParams.Description)
	assert.Equal(t, testExecutor.editedLink.Handle, calledParams.Handle)
	assert.Equal(t, testExecutor.editedLink.Name, calledParams.Name)
	assert.Equal(t, testExecutor.editedLink.Oem, calledParams.Oem)
	assert.Equal(t, testExecutor.editedLink.Version, calledParams.Version)
}

func TestPropSpecExecutor_Proceed_InvalidConfigType(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: BaseOptions{
				optionConfigType:  &config.StringOption{Option: config.Option{Key: "configType"}},
				optionUserDefined: &config.BoolOption{Option: config.Option{Key: "userDefined"}},
			},
		},
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, "notValid")
	testExecutor.Proceed()
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestPropSpecExecutor_Remove(t *testing.T) {
	var expectedKey = "expectedKey"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
		PropSpecs: []broker.PropSpec{
			{
				Key: expectedKey,
			},
		},
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli: &Cli{
				BaseCli: cli.BaseCli{
					Dry: func(ledger *config.Ledger) bool {
						return true
					},
				},
			},
		},
		PropSpecOptions: &PropSpecOptions{
			PropSpecOptions: &StubPropSpecOptions{buildPropSpec: &broker.PropSpec{
				Key: expectedKey,
			}},
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.NoError(t, testExecutor.Remove(testLink.Handle, expectedKey))
	actual := testOutput.String()
	assert.Contains(t, actual, expectedKey)
}

func TestPropSpecExecutor_Remove_Secret(t *testing.T) {
	var expectedKey = "expectedKey"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
		SecretSpecs: []broker.PropSpec{
			{
				Key: expectedKey,
			},
		},
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli: &Cli{
				BaseCli: cli.BaseCli{
					Dry: func(ledger *config.Ledger) bool {
						return true
					},
				},
			},
		},
		PropSpecOptions: &PropSpecOptions{
			PropSpecOptions: &StubPropSpecOptions{buildPropSpec: &broker.PropSpec{
				Key: expectedKey,
			}},
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.NoError(t, testExecutor.Remove(testLink.Handle, expectedKey))
	actual := testOutput.String()
	assert.Contains(t, actual, expectedKey)
}

func TestPropSpecExecutor_Remove_DataLinkNotFound(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{}),
	}

	assert.Error(t, testExecutor.Remove("testArg", "testKey"))
}

func TestPropSpecExecutor_Remove_DataWriterError(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PropSpecOptions: &PropSpecOptions{
			BaseOptions: makeTestBaseOptions(),
		},
	}

	testViper.Set(testExecutor.optionConfigType.Key, "notValid")
	assert.Error(t, testExecutor.Remove("testArg", "testKey"))
}

func TestNewAddSpecOptions(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddSpecOptions(testLedger)
	var testCmd = &cobra.Command{}

	defer patch.Unpatch()
	assert.NotNil(t, newPropAddExecutorFactory(testLedger, &Cli{}, testOptions)(testCmd))
	testViper.Set(schema.Genaiz.DataLink.PropSpecAdd.DefaultValue.Doc, "defaultValue")
	testViper.Set(schema.Genaiz.DataLink.PropSpecAdd.Secret.Doc, "True")
	_, _ = testOptions.Build(testLedger, &broker.PropSpec{})
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewEditSpecOptions(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewEditSpecOptions()
	var testCmd = &cobra.Command{}

	assert.NotNil(t, newPropEditExecutorFactory(testLedger, &Cli{}, testOptions)(testCmd))
}

func TestNewRemoveSpecOptions(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewRemoveSpecOptions()
	var testCmd = &cobra.Command{}

	assert.NotNil(t, newPropRemoveExecutorFactory(testLedger, &Cli{}, testOptions)(testCmd))
}

func TestNewProp(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{}
	var testCmd = NewProp(testLedger, testCli)

	assert.Equal(t, 3, len(testCmd.Commands()))
}

func makeTestBaseOptions() BaseOptions {
	return BaseOptions{
		optionConfigType:  &config.StringOption{Option: config.Option{Key: "configType"}},
		optionHandle:      &config.StringOption{Option: config.Option{Key: "handle"}},
		optionOem:         &config.StringOption{Option: config.Option{Key: "oem"}},
		optionVersion:     &config.StringOption{Option: config.Option{Key: "version"}},
		optionUserDefined: &config.BoolOption{Option: config.Option{Key: "userDefined"}},
	}
}

func newDataLinksWriterTestFactory(current []broker.DataLink) DataLinksWriterFactory {
	return func(ledger *config.Ledger, s string) *DataLinksWriter {
		var reader = &config.DataLinksReader{}

		return &DataLinksWriter{
			DataLinksWriter: &config.DataLinksWriter{
				DataLinksReader: *reader.WithCurrent(current),
			},
		}
	}
}

func newEditLinkCompleteCapture(capture *broker.DataLinkParams) EditLinkTaskFactory {
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

func newEditLinkPretendCapture(capture *broker.DataLinkParams) EditLinkTaskFactory {
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
