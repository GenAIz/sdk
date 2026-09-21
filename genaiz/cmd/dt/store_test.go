package dt

import (
	"bytes"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
)

func TestStoreListExecutor_List(t *testing.T) {
	var expectedArg = "oem/handle:version-rc-2"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = NewStoreListOptions()
	var testPrinter = &stubPrinter{}
	var testListFacade = &stubUserLinkInstanceFacade{
		getInstances: []mgmt.UserLinkInstance{
			{
				Id: new(int64(37)),
			},
		},
	}
	var testExecutor = &StoreListExecutor{
		BaseListExecutor: BaseListExecutor{
			ledger:        testLedger,
			accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams: &stubPrinterParametric{printer: testPrinter},
		},
		StoreListOptions: testOptions,

		userDataStoreFacadeProvider: func() mgmt.UserDataStoreFacade {
			return testListFacade
		},
	}

	testLedger.InitLogging()
	assert.NoError(t, testExecutor.List(expectedArg))

	if actual, ok := testPrinter.printOut.([]mgmt.UserLinkInstance); ok {
		assert.Equal(t, testListFacade.getInstances, actual)
	} else {
		assert.Fail(t, "did not receive the stores")
	}
}

func TestStoreListExecutor_List_GetError(t *testing.T) {
	var expectedArg = "oem/handle:version"
	var expectedError = task.NewError("expected")
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = NewStoreListOptions()
	var testPrinter = &stubPrinter{
		err: expectedError,
	}
	var testListFacade = &stubUserLinkInstanceFacade{
		getError: expectedError,
	}
	var testExecutor = &StoreListExecutor{
		BaseListExecutor: BaseListExecutor{
			ledger:        testLedger,
			accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams: &stubPrinterParametric{printer: testPrinter},
		},
		StoreListOptions: testOptions,

		userDataStoreFacadeProvider: func() mgmt.UserDataStoreFacade {
			return testListFacade
		},
	}

	testLedger.InitLogging()
	assert.ErrorIs(t, testExecutor.List(expectedArg), expectedError)

	if actual, ok := testPrinter.printError.(task.Error); ok {
		assert.ErrorIs(t, actual, testListFacade.getError)
	} else {
		assert.Fail(t, "did not receive the error")
	}
}

func TestStoreListExecutor_List_NoArg(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = NewStoreListOptions()
	var testPrinter = &stubPrinter{}
	var testListFacade = &stubUserLinkInstanceFacade{
		getInstances: []mgmt.UserLinkInstance{
			{
				Id: new(int64(37)),
			},
		},
	}
	var testExecutor = &StoreListExecutor{
		BaseListExecutor: BaseListExecutor{
			ledger:        testLedger,
			accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams: &stubPrinterParametric{printer: testPrinter},
		},
		StoreListOptions: testOptions,

		userDataStoreFacadeProvider: func() mgmt.UserDataStoreFacade {
			return testListFacade
		},
	}

	testLedger.InitLogging()
	assert.NoError(t, testExecutor.List(""))

	if actual, ok := testPrinter.printOut.([]mgmt.UserLinkInstance); ok {
		assert.Equal(t, testListFacade.getInstances, actual)
	} else {
		assert.Fail(t, "did not receive the stores")
	}
}

func TestStoreListExecutor_List_SeqError(t *testing.T) {
	var expectedArg = "oem/handle:version-rc-F"
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewStoreListOptions()
	var testCmd = &cobra.Command{}
	var testExecutor = newStoreListExecutorFactory(testLedger, testOptions)(testCmd)

	assert.ErrorIs(t, testExecutor.List(expectedArg), errorLinkSeqInvalid)
}
