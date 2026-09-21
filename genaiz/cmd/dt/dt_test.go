package dt

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubPrinter struct {
	err        error
	printError interface{}
	printOut   interface{}
}

func (s *stubPrinter) Error(i interface{}) error {
	s.printError = i
	return s.err
}

func (s *stubPrinter) Print(i interface{}) error {
	s.printOut = i
	return s.err
}

type stubPrinterParametric struct {
	defaultPrinter bool
	printer        cli.Printer
}

func (s stubPrinterParametric) IsDefault() bool {
	return s.defaultPrinter
}

func (s stubPrinterParametric) Printer() cli.Printer {
	return s.printer
}

type stubUserLinkInstanceFacade struct {
	getInstances []mgmt.UserLinkInstance
	getError     task.Error
	filter       string
	logger       *logrus.Logger
	params       *broker.DataInstanceListParams
}

func (s *stubUserLinkInstanceFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserLinkInstance] {
	s.filter = filter
	return s
}

func (s *stubUserLinkInstanceFacade) Get() ([]mgmt.UserLinkInstance, task.Error) {
	return s.getInstances, s.getError
}

func (s *stubUserLinkInstanceFacade) Provider() mgmt.Provider[[]mgmt.UserLinkInstance] {
	return s
}

func (s *stubUserLinkInstanceFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserLinkInstance, broker.DataInstanceListParams] {
	s.logger = logger
	return s
}

func (s *stubUserLinkInstanceFacade) WithParams(params *broker.DataInstanceListParams) mgmt.Facade[[]mgmt.UserLinkInstance, broker.DataInstanceListParams] {
	s.params = params
	return s
}

func TestNewDt(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewDt(testLedger)

	assert.Equal(t, 2, len(testCmd.Commands()))
}
