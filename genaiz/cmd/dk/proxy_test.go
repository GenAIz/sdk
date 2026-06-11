package dk

import (
	"bytes"
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
	"genaiz.com/genaiz/task/broker"
)

func TestProxyExecutor_Add(t *testing.T) {
	var expectedHost = "host"
	var expectedPort = 1337
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
	var testExecutor = &ProxyExecutor{
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
		ProxyOptions: &ProxyOptions{
			BaseOptions: makeTestBaseOptions(),
			optionTcp: cli.Options.Proxies.Tcp().
				WithKeys(&schema.Keys{Doc: "tcp"}).
				BuildBoolOption(),
			optionUdp: cli.Options.Proxies.Udp().
				WithKeys(&schema.Keys{Doc: "udp"}).
				BuildBoolOption(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.NoError(t, testExecutor.Add(testLink.Handle, expectedHost, expectedPort))
	actual := testOutput.String()
	assert.Contains(t, actual, expectedHost)
	assert.Contains(t, actual, cast.ToString(expectedPort))
}

func TestProxyExecutor_Add_DataWriterError(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
			BaseOptions: makeTestBaseOptions(),
		},
	}

	testViper.Set(testExecutor.optionConfigType.Key, "notValid")
	assert.Error(t, testExecutor.Add("testArg", "testHost", 1337))
}

func TestProxyExecutor_Add_DataLinkNotFound(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{}),
	}

	assert.Error(t, testExecutor.Add("testArg", "testKey", 37))
}

func TestProxyExecutor_Add_ProxyDuplicated(t *testing.T) {
	var expectedHost = "host"
	var expectedPort = 1337
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
		OutboundProxies: []broker.Proxy{
			{
				Host: expectedHost,
				Port: expectedPort,
			},
		},
	}
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.Error(t, testExecutor.Add(testLink.Handle, expectedHost, expectedPort))
}

func TestProxyExecutor_Display(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testExecutor = &ProxyExecutor{}

	defer patch.Unpatch()
	testExecutor.Display()
	assert.NotEmpty(t, patch.CalledWith)
	params := cast.ToStringSlice(patch.CalledWith)
	assert.NotEmpty(t, params[0])
}

func TestProxyExecutor_Pretend(t *testing.T) {
	var calledParams broker.DataLinkParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
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

func TestProxyExecutor_Pretend_InvalidConfigType(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
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

func TestProxyExecutor_Proceed(t *testing.T) {
	var calledParams broker.DataLinkParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
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

func TestProxyExecutor_Proceed_InvalidConfigType(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
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

func TestProxyExecutor_Remove(t *testing.T) {
	var expectedHost = "host"
	var expectedPort = 1337
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
		OutboundProxies: []broker.Proxy{
			{
				Host: expectedHost,
				Port: expectedPort,
			},
		},
	}
	var testExecutor = &ProxyExecutor{
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
		ProxyOptions: &ProxyOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.NoError(t, testExecutor.Remove(testLink.Handle, expectedHost, expectedPort))
	actual := testOutput.String()
	assert.Contains(t, actual, expectedHost)
	assert.Contains(t, actual, cast.ToString(expectedPort))
}

func TestProxyExecutor_Remove_DataLinkNotFound(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{}),
	}

	assert.Error(t, testExecutor.Remove("testArg", "testHost", 37))
}

func TestProxyExecutor_Remove_DataWriterError(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
			BaseOptions: makeTestBaseOptions(),
		},
	}

	testViper.Set(testExecutor.optionConfigType.Key, "notValid")
	assert.Error(t, testExecutor.Remove("testArg", "testHost", 37))
}

func TestProxyExecutor_Remove_ProxyNotFound(t *testing.T) {
	var expectedHost = "host"
	var expectedPort = 1337
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testLink = &broker.DataLink{
		Handle:          "testHandle",
		Oem:             "testOem",
		Version:         "testVersion",
		OutboundProxies: []broker.Proxy{},
	}
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ProxyOptions: &ProxyOptions{
			BaseOptions: makeTestBaseOptions(),
		},

		dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testLink}),
	}

	testViper.Set(testExecutor.optionOem.Key, testLink.Oem)
	testViper.Set(testExecutor.optionVersion.Key, testLink.Version)
	assert.NoError(t, testExecutor.Remove(testLink.Handle, expectedHost, expectedPort))
}

func TestNewProxy(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCli = &Cli{}
	var testProxy = NewProxy(testLedger, testCli)

	assert.Equal(t, 2, len(testProxy.Commands()))
}

func TestNewProxyAddOptions(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewProxyAddOptions()
	var testCmd = &cobra.Command{}
	var testExecutor = newProxyAddFactory(testLedger, &Cli{}, testOptions)(testCmd)

	if testExecutor != nil {
		testViper.Set(testOptions.optionUdp.Key, false)
		assert.True(t, testLedger.GetBool(testOptions.optionTcp))
		return
	}

	assert.Fail(t, "could not make an AddFactory")
}

func TestNewProxyRemoveOptions(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewProxyRemoveOptions()
	var testCmd = &cobra.Command{}
	var testExecutor = newProxyRemoveFactory(testLedger, &Cli{}, testOptions)(testCmd)

	if testExecutor != nil {
		assert.Nil(t, testOptions.optionTcp)
		assert.Nil(t, testOptions.optionUdp)
		return
	}

	assert.Fail(t, "could not make an AddFactory")
}
