package sf

import (
	"os"
	"path/filepath"
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

func TestProxyExecutor_Add(t *testing.T) {
	var expectedProxies = []broker.Proxy{
		{
			Host: "anotherHost",
			Port: 22,
		},
		{
			Host:  "expectedProxy",
			Port:  37,
			Flags: 3,
		},
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = NewProxyAddOptions()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		ProxyOptions: testOptions,

		innerProxies: cli.NewOptionBuilder().
			WithKeys(&schema.Keys{Doc: "proxy"}).
			BuildOption(),
		innerType: cli.NewOptionBuilder().
			WithKeys(&schema.Keys{Doc: "type"}).
			BuildStringOption(),
	}

	testViper.Set(testExecutor.innerType.Key, layout.FunctionTypeConnector)
	testViper.Set(testExecutor.innerProxies.Key, []interface{}{
		map[string]any{
			"host": expectedProxies[0].Host,
			"port": expectedProxies[0].Port,
		},
	})
	assert.NoError(t, testExecutor.Add(expectedProxies[1].Host, expectedProxies[1].Port))
	assert.Equal(t, &expectedProxies[1], testExecutor.addedProxy)
	assert.Equal(t, expectedProxies, testExecutor.updatedProxies)
}

func TestProxyExecutor_Add_AlreadyConfigured(t *testing.T) {
	var expectedHost = "expectedHost"
	var expectedPort = 37
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		innerProxies: cli.NewOptionBuilder().
			WithKeys(&schema.Keys{Doc: "proxy"}).
			BuildOption(),
		innerType: cli.NewOptionBuilder().
			WithKeys(&schema.Keys{Doc: "type"}).
			BuildStringOption(),
	}

	testViper.Set(testExecutor.innerType.Key, layout.FunctionTypeConnector)
	testViper.Set(testExecutor.innerProxies.Key, []interface{}{
		map[string]any{
			"host": expectedHost,
			"port": expectedPort,
		},
		map[string]any{
			"host": "anotherHost",
			"port": 22,
		},
	})
	assert.Error(t, testExecutor.Add(expectedHost, expectedPort))
	assert.Empty(t, testExecutor.addedProxy)
	assert.Empty(t, testExecutor.updatedProxies)
}

func TestProxyExecutor_Add_InvalidConnector(t *testing.T) {
	var expectedHost = "expectedHost"
	var expectedPort = 37
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewProxyAddOptions()
	var testExecutor = newProxyAddFactory(testLedger, testCli, testOptions)(&cobra.Command{})

	testViper.Set(schema.Genaiz.Function.Publish.Type.Doc, layout.FunctionTypeTrigger)
	assert.ErrorIs(t, testExecutor.Add(expectedHost, expectedPort), errInvalidConnectorType)
}

func TestProxyExecutor_Display_NoAddRemove(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testExecutor = &ProxyExecutor{}

	defer patch.Unpatch()
	testExecutor.Display()

	assert.Empty(t, patch.CalledWith)
}

func TestProxyExecutor_Pretend(t *testing.T) {
	var calledInit bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		ProxyOptions: NewProxyAddOptions(),

		initTaskFactory: newInitTaskPretendStub(&calledInit),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.Pretend()
		assert.True(t, calledInit)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestProxyExecutor_Proceed(t *testing.T) {
	var calledInit bool
	var expectedProxy = &broker.Proxy{
		Host: "expectedHost",
		Port: 37,
	}
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		ProxyOptions: NewProxyAddOptions(),

		addedProxy:     expectedProxy,
		updatedProxies: []broker.Proxy{*expectedProxy},
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
			assert.NotEmpty(t, params.OutboundProxies)
			assert.Equal(t, expectedProxy.Host, params.OutboundProxies[0].Host)
			assert.Equal(t, expectedProxy.Port, params.OutboundProxies[0].Port)
		}),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testLedger.Logger = logrus.New()
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.Proceed()
		assert.True(t, calledInit)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestProxyExecutor_Remove(t *testing.T) {
	var removedProxy = &broker.Proxy{
		Host: "expectedHost",
		Port: 37,
	}
	var remainingProxy = &broker.Proxy{
		Host: "remainingHost",
		Port: 22,
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		innerProxies: cli.NewOptionBuilder().
			WithKeys(&schema.Keys{Doc: "proxy"}).
			BuildOption(),
	}

	testViper.Set(testExecutor.innerProxies.Key, []interface{}{
		map[string]any{
			"host": removedProxy.Host,
			"port": removedProxy.Port,
		},
		map[string]any{
			"host": remainingProxy.Host,
			"port": remainingProxy.Port,
		},
	})
	testExecutor.Remove(removedProxy.Host, removedProxy.Port)
	assert.Equal(t, removedProxy, testExecutor.removedProxy)
	assert.Equal(t, []broker.Proxy{*remainingProxy}, testExecutor.updatedProxies)
}

func TestProxyExecutor_Remove_NotFound(t *testing.T) {
	var expectedHost = "expectedHost"
	var expectedPort = 37
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = newProxyRemoveFactory(testLedger, NewSfCli(nil, nil, nil))(&cobra.Command{})

	testExecutor.Remove(expectedHost, expectedPort)
	concreteExecutor, ok := testExecutor.(*ProxyExecutor)
	assert.True(t, ok)
	assert.Empty(t, concreteExecutor.removedProxy)
	assert.Empty(t, concreteExecutor.updatedProxies)
}
