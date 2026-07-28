package ac

import (
	"errors"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestInspectExecutor_Inspect(t *testing.T) {
	var expectedSession = &broker.Session{Id: int64(37)}
	var testHostAddr = "hostAddr"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testPrinter = &stubCliPrinter{}
	var testExecutor = &InspectExecutor{
		ledger:                 testLedger,
		inspectAuthTaskFactory: newInspectAuthTaskCompleteStub(expectedSession, nil),
		printerParams: stubCliPrinterParametric{
			stubPrinter: testPrinter,
		},
	}

	testLedger.InitLogging()
	assert.NoError(t, testExecutor.Inspect(testHostAddr))

	if actual, ok := testPrinter.printPayload.([]mgmt.UserSession); ok {
		assert.Equal(t, expectedSession.Id, actual[0].Id)
	} else {
		assert.Fail(t, "expected a list of user sessions")
	}
}

func TestInspectExecutor_Inspect_JsonPrinter(t *testing.T) {
	var expectedSession = &broker.Session{Id: int64(37)}
	var testHostAddr = "hostAddr"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testPrinter = &stubCliPrinter{}
	var testExecutor = &InspectExecutor{
		ledger:                 testLedger,
		inspectAuthTaskFactory: newInspectAuthTaskCompleteStub(expectedSession, nil),
		printerParams: stubCliPrinterParametric{
			stubPrinter: testPrinter,
			stubDefault: new(false),
		},
	}

	testLedger.InitLogging()
	assert.NoError(t, testExecutor.Inspect(testHostAddr))

	if actual, ok := testPrinter.printPayload.(mgmt.UserSession); ok {
		assert.Equal(t, expectedSession.Id, actual.Id)
	} else {
		assert.Fail(t, "expected a user session")
	}
}

func TestInspectExecutor_Inspect_Error(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHostAddr = "hostAddr"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testPrinter = &stubCliPrinter{
		errorError: expectedError,
	}
	var testExecutor = &InspectExecutor{
		ledger:                 testLedger,
		inspectAuthTaskFactory: newInspectAuthTaskCompleteStub(nil, expectedError),
		printerParams: stubCliPrinterParametric{
			stubPrinter: testPrinter,
		},
	}

	testLedger.InitLogging()
	assert.ErrorIs(t, testExecutor.Inspect(testHostAddr), expectedError)
}

func TestNewInspect(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithUserPath(t.TempDir()).
		WithViper(testViper).
		Build()
	var testList = NewInspect(testLedger)

	defer patch.Unpatch()
	testLedger.InitLogging()
	testList.Run(testList, []string{"filter"})
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newInspectAuthTaskCompleteStub(session *broker.Session, expected error) inspectAuthTaskFactory {
	return func() *task.Task[broker.Broker] {
		return &task.Task[broker.Broker]{
			OnPrepare: func(params *broker.Broker, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.Broker, state *task.State) error {
				if session != nil {
					state.Internal = *session
				}

				return expected
			},
		}
	}
}
