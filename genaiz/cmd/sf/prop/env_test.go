package prop

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestEnvExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(testOutput).
		Build()
	var testExecutor = &EnvExecutor{
		EnvOptions: NewEnvOptions(),
		Ledger:     testLedger,

		key:   "expectedKey",
		value: "expectedValue",
	}
	var expectedContext = "expectedContext"
	var expectedEnvFile = "expectedFile"

	testViper.Set(testExecutor.optionContext.Key, expectedContext)
	testViper.Set(testExecutor.optionEnvFile.Key, expectedEnvFile)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testExecutor.optionContext.Param+`:[\s\t]*`+expectedContext), actual)
	assert.Regexp(t, regexp.MustCompile(testExecutor.optionEnvFile.Param+`:[\s\t]*`+expectedEnvFile), actual)
	assert.Regexp(t, regexp.MustCompile(`key:[\s\t]*`+testExecutor.key), actual)
	assert.Regexp(t, regexp.MustCompile(`value:[\s\t]*`+testExecutor.value), actual)
}

func TestEnvExecutor_List(t *testing.T) {
	var testViper = viper.New()
	var testOptions = NewEnvOptions()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "")
	var actualKeys []cobra.Completion
	var err error

	testLedger.InitLogging()
	actualKeys, err = testExecutor.List()
	assert.Nil(t, actualKeys)
	assert.Error(t, err)
	assert.Contains(t, "not found", err.Error())
}

func TestEnvExecutor_List_InvalidDataLink(t *testing.T) {
	var testViper = viper.New()
	var testOptions = NewEnvOptions()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "")
	var expectedDataLinks = "expectedInvalid"
	var actualKeys []cobra.Completion
	var err error

	testViper.Set(schema.Genaiz.Function.Publish.PropSpecs.Doc, []broker.PropSpec{})
	testViper.Set(schema.Genaiz.Function.Publish.DataSources.Doc, []string{expectedDataLinks})
	testLedger.InitLogging()
	actualKeys, err = testExecutor.List()
	assert.Nil(t, actualKeys)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedDataLinks)
}

func TestEnvExecutor_List_NoDataLinks(t *testing.T) {
	var testViper = viper.New()
	var testOptions = NewEnvOptions()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testPropSpecs = []broker.PropSpec{
		{
			Key: "key1",
		},
		{
			Key:         "key2",
			Description: "key2",
		},
		{
			Key:         "key3",
			Description: "diff description",
		},
	}
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "", "")
	var expectedKeys = []string{
		testPropSpecs[0].Key,
		testPropSpecs[1].Key,
		fmt.Sprintf("%s\t%s", testPropSpecs[2].Key, testPropSpecs[2].Description),
	}
	var actualKeys []cobra.Completion
	var err error

	testViper.Set(schema.Genaiz.Function.Publish.PropSpecs.Doc, testPropSpecs)
	testLedger.InitLogging()
	actualKeys, err = testExecutor.List()
	assert.NoError(t, err)
	assert.Equal(t, expectedKeys, actualKeys)
}

func TestEnvExecutor_List_NoPropSpecs(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testVarSpecs = []shared.VarSpec{
		broker.PropSpec{
			Key: "key1",
		},
		broker.PropSpec{
			Key:         "key2",
			Description: "key2",
		},
		broker.PropSpec{
			Key:         "key3",
			Description: "diff description",
		},
	}
	var testExecutor = &EnvExecutor{
		EnvOptions: NewEnvOptions(),
		Ledger:     testLedger,
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkCompleteReturn(testVarSpecs)).
			Build(),

		innerSources: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataSources).
			BuildListOption(),
		innerStores: &config.ListOption{},
	}
	var expectedKeys = []string{
		testVarSpecs[0].GetKey(),
		testVarSpecs[1].GetKey(),
		fmt.Sprintf("%s\t%s", testVarSpecs[2].GetKey(), testVarSpecs[2].GetDescription()),
	}
	var actualKeys []cobra.Completion
	var err error

	testViper.Set(schema.Genaiz.Function.Publish.DataSources.Doc, []string{
		fmt.Sprintf("%s/%s:%s", testDataLink.Oem, testDataLink.Handle, testDataLink.Version),
	})
	testLedger.InitLogging()
	actualKeys, err = testExecutor.List()
	assert.NoError(t, err)
	assert.Equal(t, expectedKeys, actualKeys)
}

func TestEnvExecutor_Pretend(t *testing.T) {
	var testDir = t.TempDir()
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, ".env")); err == nil {
		var testViper = viper.New()
		var testLedger = config.NewBuilder().
			WithViper(testViper).
			Build()
		var testOptions = NewEnvOptions()
		var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

		if _, err = fd.Write([]byte(fmt.Sprintf("%s=\"%s\"", testExecutor.key, "valueToReplace"))); err == nil {
			var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})

			defer patch.Unpatch()
			testLedger.WorkDir = testDir
			testViper.Set(schema.Genaiz.Function.Publish.PropSpecs.Doc, []interface{}{
				map[string]interface{}{
					"key":  testExecutor.key,
					"type": broker.PropSpecTypeString,
				},
			})
			testExecutor.Pretend()
			assert.NotEmpty(t, patch.CalledWith)
			params := cast.ToStringSlice(patch.CalledWith)
			assert.Equal(t, testExecutor.key, params[1])
			assert.Equal(t, testExecutor.value, params[2])
			assert.Equal(t, testOptions.optionEnvFile.DefaultGetter(testLedger), params[3])
		}
	}

	assert.NoError(t, err)
}

func TestEnvExecutor_Pretend_PathError(t *testing.T) {
	var testDir = t.TempDir()
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewEnvOptions()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

	defer patch.Unpatch()
	testLedger.WorkDir = testDir
	testViper.Set(schema.Genaiz.Function.Publish.PropSpecs.Doc, []interface{}{
		map[string]interface{}{
			"key":  testExecutor.key,
			"type": broker.PropSpecTypeString,
		},
	})
	testExecutor.Pretend()
	assert.NotEmpty(t, patch.CalledWith)
	params := cast.ToStringSlice(patch.CalledWith)
	assert.Equal(t, testExecutor.key, params[1])
	assert.Equal(t, testExecutor.value, params[2])
	assert.Equal(t, testOptions.optionEnvFile.DefaultGetter(testLedger), params[3])
}

func TestEnvExecutor_Pretend_PermissionError(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "not_exist")
	var fd *os.File
	var err error

	if fd, err = os.OpenFile(testFile, os.O_CREATE|os.O_TRUNC, 0222); err == nil {
		defer filez.CloseSilently(fd)
		var patch = mock.Patches{T: t}.OsExit(func(int) {})
		var testViper = viper.New()
		var testLedger = config.NewBuilder().
			WithViper(testViper).
			Build()
		var testOptions = NewEnvOptions()
		var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

		defer patch.Unpatch()
		testViper.Set(testOptions.optionEnvFile.Key, testFile)
		testViper.Set(schema.Genaiz.Function.Publish.PropSpecs.Doc, []interface{}{
			map[string]interface{}{
				"key":  testExecutor.key,
				"type": broker.PropSpecTypeString,
			},
		})
		testExecutor.Pretend()
		assert.True(t, patch.Called)
		assert.EqualValues(t, 1, patch.CalledWith)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestEnvExecutor_Pretend_ValidationError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewEnvOptions()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

	defer patch.Unpatch()
	testExecutor.Pretend()
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestEnvExecutor_Pretend_VarSpecNotFound(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var expectedKey = "expectedKey"
	var testExecutor = &EnvExecutor{
		EnvOptions: NewEnvOptions(),
		Ledger:     testLedger,
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkCompleteReturn([]shared.VarSpec{})).
			Build(),

		innerSources: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataSources).
			BuildListOption(),
		innerStores: &config.ListOption{},
		key:         expectedKey,
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.innerSources.Key, []string{
		fmt.Sprintf("%s/%s:%s", testDataLink.Oem, testDataLink.Handle, testDataLink.Version),
	})
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestEnvExecutor_Proceed(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), ".env")
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		var testViper = viper.New()
		var testLedger = config.NewBuilder().WithViper(testViper).Build()
		var testOptions = NewEnvOptions()
		var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")
		var testKeyPair = fmt.Sprintf("%s=\"%s\"\n", testExecutor.key, "notRightValue")
		var expectedOther = fmt.Sprintf("%s=\"%s\"\n", "otherKey", "otherValue")

		defer filez.CloseSilently(fd)

		if _, err = fd.Write([]byte(testKeyPair + expectedOther)); err == nil {
			var patch = mock.Patches{T: t}.OsExit(func(int) {})
			var content []byte

			defer patch.Unpatch()
			testViper.Set(testExecutor.optionEnvFile.Key, testFile)
			testViper.Set(schema.Genaiz.Function.Publish.PropSpecs.Doc, []interface{}{
				map[string]interface{}{
					"key":  testExecutor.key,
					"type": broker.PropSpecTypeString,
				},
			})
			testExecutor.Proceed()

			if content, err = os.ReadFile(testFile); err == nil {
				var expectedKeyPair = fmt.Sprintf("%s=\"%s\"\n", testExecutor.key, testExecutor.value)

				actual := string(content)
				assert.Contains(t, actual, expectedOther)
				assert.Contains(t, actual, expectedKeyPair)
				assert.Empty(t, patch.CalledWith)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestEnvExecutor_Proceed_CreateFileError(t *testing.T) {
	var testDir = t.TempDir()
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewEnvOptions()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionEnvFile.Key, filepath.Join(testDir, "not_exist", ".env"))
	testViper.Set(schema.Genaiz.Function.Publish.PropSpecs.Doc, []interface{}{
		map[string]interface{}{
			"key":  testExecutor.key,
			"type": broker.PropSpecTypeString,
		},
	})
	testExecutor.Proceed()
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestEnvExecutor_Proceed_ValidationError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewEnvOptions()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

	defer patch.Unpatch()
	testExecutor.Proceed()
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestEnvExecutor_Proceed_VarSpecFailure(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var expectedError = errors.New("expected")
	var testExecutor = &EnvExecutor{
		EnvOptions: NewEnvOptions(),
		Ledger:     testLedger,
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkCompleteFailure(expectedError)).
			Build(),

		innerSources: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataSources).
			BuildListOption(),
		innerStores: &config.ListOption{},
		key:         "key",
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.innerSources.Key, []string{
		fmt.Sprintf("%s/%s:%s", testDataLink.Oem, testDataLink.Handle, testDataLink.Version),
	})
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestEnvExecutor_Proceed_VarSpecSuccess(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testFile = filepath.Join(t.TempDir(), ".env")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testVarSpec = &broker.PropSpec{
		Key: "expectedKey",
	}
	var testExecutor = &EnvExecutor{
		EnvOptions: NewEnvOptions(),
		Ledger:     testLedger,
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkCompleteReturn([]shared.VarSpec{*testVarSpec})).
			Build(),

		innerSources: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataSources).
			BuildListOption(),
		innerStores: &config.ListOption{},
		key:         testVarSpec.Key,
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.innerSources.Key, []string{
		fmt.Sprintf("%s/%s:%s", testDataLink.Oem, testDataLink.Handle, testDataLink.Version),
	})
	testViper.Set(testExecutor.optionEnvFile.Key, testFile)
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.False(t, patch.Called)
}

func TestNewEnv(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, ".env")
	var envCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(testOutput).
		Build()
	var testEnv = NewEnv(testLedger, &cli.BaseCli{Dry: func(ledger *config.Ledger) bool {
		return true
	}})
	var expectedKey = "EXPECTED_KEY"
	var expectedValue = "expectedValue"

	testEnv.PostRun = func(cmd *cobra.Command, args []string) {
		envCompleted = true
	}
	testViper.Set(schema.Genaiz.Function.Env.File.Doc, testFile)
	testViper.Set(schema.Genaiz.Function.Env.Context.Doc, testDir)
	testEnv.SetArgs([]string{expectedKey, expectedValue})
	// Execute will change the work dir, so we need to reset it, by changing it with testing first
	t.Chdir(testDir)
	assert.NoError(t, testEnv.Execute())
	assert.True(t, envCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, testFile)
		assert.Contains(t, actual, testDir)
		assert.Contains(t, actual, expectedKey)
		assert.Contains(t, actual, expectedValue)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func TestNewEnv_Validation(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testEnv = NewEnv(testLedger, &cli.BaseCli{Dry: func(ledger *config.Ledger) bool {
		return true
	}})

	assert.Error(t, testEnv.ValidateArgs([]string{"..NotValid..", "value"}))
	assert.NoError(t, testEnv.ValidateArgs([]string{"MY_KEY", "value"}))
}

func TestNewEnv_ValidArgsFunction(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testConfig = filepath.Join(t.TempDir(), testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var expectedKey = "expectedKey"
	var err error

	testViper.Set(schema.Genaiz.Function.Publish.PropSpecs.Doc, []broker.PropSpec{
		{
			Key: expectedKey,
		},
	})

	if err = testViper.WriteConfigAs(testConfig); err == nil {
		var testEnv = NewEnv(testLedger, &cli.BaseCli{})

		testLedger.InitLogging()
		actual, directive := testEnv.ValidArgsFunction(testEnv, []string{"one"}, "complete")
		assert.Equal(t, []string{expectedKey}, actual)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestNewEnv_ValidArgsFunction_ListError(t *testing.T) {
	var testLedger = config.NewBuilder().Build()
	var testEnv = NewEnv(testLedger, &cli.BaseCli{})

	testLedger.InitLogging()
	actual, directive := testEnv.ValidArgsFunction(testEnv, []string{"one"}, "complete")
	assert.Empty(t, actual)
	assert.Equal(t, cobra.ShellCompDirectiveError|cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestNewEnv_ValidArgsFunction_NoArgs(t *testing.T) {
	var testLedger = config.NewBuilder().Build()
	var testEnv = NewEnv(testLedger, &cli.BaseCli{})

	actual, directive := testEnv.ValidArgsFunction(testEnv, []string{"one", "two"}, "complete")
	assert.Empty(t, actual)
	assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
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

func newExportLinkCompleteReturn(varSpecs []shared.VarSpec) dk.ExportLinkTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				state.Internal = shared.VarSpecTracking{
					VarSpecs: varSpecs,
				}
				return nil
			},
		}
	}
}

func newExportLinkCompleteFailure(err error) dk.ExportLinkTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				return err
			},
		}
	}
}
