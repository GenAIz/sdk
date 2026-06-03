package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	io2 "genaiz.com/genaiz-lib/mock/io"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
)

type shrinkRedGreenColumns struct {
	A      string `cli:"Column_A"`
	B      string `cli:"Column_B"`
	C      string `cli:"Column_C" redGreen:"Column_D"`
	D      string `cli:"Column_D,noShow"`
	shrink bool
}

func (stc shrinkRedGreenColumns) MarshalSlice() ([]string, error) {
	if stc.shrink {
		return []string{stc.A, stc.B, stc.C}, nil
	}

	return []string{stc.A, stc.B, stc.C, stc.D}, nil
}

type shrinkSelectedColumns struct {
	A      string `cli:"Column_A"`
	B      string `cli:"Column_B"`
	C      string `cli:"Selector" selected:"X"`
	shrink bool
}

func (stc shrinkSelectedColumns) MarshalSlice() ([]string, error) {
	if stc.shrink {
		return []string{stc.A, stc.B}, nil
	}

	return []string{stc.A, stc.B, stc.C}, nil
}

type simpleTwoColumns struct {
	A   string `cli:"Column_A"`
	B   string `cli:"Column_B"`
	Err error
}

func (stc simpleTwoColumns) MarshalSlice() ([]string, error) {
	if stc.Err == nil {
		return []string{stc.A, stc.B}, nil
	}

	return nil, stc.Err
}

type stubPrinter struct {
	err        error
	printerErr interface{}
	printerOut interface{}
}

func (sp *stubPrinter) Error(i interface{}) error {
	sp.printerErr = i
	return sp.err
}

func (sp *stubPrinter) Print(i interface{}) error {
	sp.printerOut = i
	return sp.err
}

func TestConsolePrinter_Error(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedField = "expectedField"
	var expectedAllowed1 = "allowed1"
	var expectedAllowed2 = "allowed2"
	var expectedStatus = 400
	var expectedMessage = "expectedMessage"
	var testError = task.NewErrorBuilder().
		Field(expectedField).
		Allowed(expectedAllowed1, expectedAllowed2).
		Status(expectedStatus).
		Build(expectedMessage)
	var stderrRestore = os.Stderr
	var testPrinter Printer

	defer patch.Unpatch()
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = stderrRestore
	}()

	testPrinter = NewConsolePrinter(os.Stderr, os.Stdout)
	assert.NoError(t, testPrinter.Error(testError))
	_ = w.Close()
	err, _ := io.ReadAll(r)
	actual := string(err)
	assert.Equal(t, actual, fmt.Sprintf("Error: %s\n", expectedMessage))
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestConsolePrinter_Print_Slice_Empty(t *testing.T) {
	var testError = new(bytes.Buffer)
	var testOutput = new(bytes.Buffer)
	var testPrinter = NewConsolePrinter(testError, testOutput)

	assert.NoError(t, testPrinter.Print([]string{}))
	actualError := testError.String()
	actualOutput := testOutput.String()
	assert.Empty(t, actualError)
	assert.Empty(t, actualOutput)
}

func TestConsolePrinter_Print_Slice_MarshalSliceError(t *testing.T) {
	var testError = new(bytes.Buffer)
	var testOutput = new(bytes.Buffer)
	var testPrinter = NewConsolePrinter(testError, testOutput)
	var testStruct = simpleTwoColumns{
		A:   "valueA",
		B:   "valueB",
		Err: errors.New("expected"),
	}

	assert.PanicsWithError(t, testStruct.Err.Error(), func() { _ = testPrinter.Print([]simpleTwoColumns{testStruct}) })
	actualError := testError.String()
	actualOutput := testOutput.String()
	assert.Empty(t, actualError)
	assert.NotContains(t, actualOutput, "COLUMN_A")
	assert.NotContains(t, actualOutput, testStruct.A)
	assert.NotContains(t, actualOutput, "COLUMN_B")
	assert.NotContains(t, actualOutput, testStruct.B)
}

func TestConsolePrinter_Print_Slice_RedGreen(t *testing.T) {
	var testError = new(bytes.Buffer)
	var testOutput = new(bytes.Buffer)
	var testPrinter = NewConsolePrinter(testError, testOutput)
	var testFirst = shrinkRedGreenColumns{A: "OneA", B: "OneB", C: "OneC", D: "yes"}
	var testSecond = shrinkRedGreenColumns{A: "TwoA", B: "TwoB", C: "OneC", shrink: true}
	var testThird = shrinkRedGreenColumns{A: "ThreeA", B: "ThreeB", C: "ThreeC", D: "no"}

	assert.NoError(t, testPrinter.Print([]shrinkRedGreenColumns{testFirst, testSecond, testThird}))
	actualError := testError.String()
	actualOutput := testOutput.String()
	assert.Empty(t, actualError)
	assert.Contains(t, actualOutput, "COLUMN_A")
	assert.Contains(t, actualOutput, testFirst.A)
	assert.Contains(t, actualOutput, testSecond.A)
	assert.Contains(t, actualOutput, testThird.A)
	assert.Contains(t, actualOutput, "COLUMN_B")
	assert.Contains(t, actualOutput, testFirst.B)
	assert.Contains(t, actualOutput, testSecond.B)
	assert.Contains(t, actualOutput, testThird.B)
	assert.Contains(t, actualOutput, "COLUMN_C")
	assert.Contains(t, actualOutput, testFirst.C)
	assert.Contains(t, actualOutput, testSecond.C)
	assert.Contains(t, actualOutput, testThird.C)
	assert.NotContains(t, actualOutput, "Column_D")
	assert.NotContains(t, actualOutput, "COLUMN_D")
}

func TestConsolePrinter_Print_Slice_Selector(t *testing.T) {
	var testError = new(bytes.Buffer)
	var testOutput = new(bytes.Buffer)
	var testPrinter = NewConsolePrinter(testError, testOutput)
	var testFirst = shrinkSelectedColumns{A: "OneA", B: "OneB", C: "yes"}
	var testSecond = shrinkSelectedColumns{A: "TwoA", B: "TwoB", shrink: true}
	var testThird = shrinkSelectedColumns{A: "ThreeA", B: "ThreeB", C: "no"}

	assert.NoError(t, testPrinter.Print([]shrinkSelectedColumns{testFirst, testSecond, testThird}))
	actualError := testError.String()
	actualOutput := testOutput.String()
	assert.Empty(t, actualError)
	assert.Contains(t, actualOutput, "COLUMN_A")
	assert.Contains(t, actualOutput, testFirst.A)
	assert.Contains(t, actualOutput, testSecond.A)
	assert.Contains(t, actualOutput, testThird.A)
	assert.Contains(t, actualOutput, "COLUMN_B")
	assert.Contains(t, actualOutput, testFirst.B)
	assert.Contains(t, actualOutput, testSecond.B)
	assert.Contains(t, actualOutput, testThird.B)
	assert.Contains(t, actualOutput, "SELECTOR")
	assert.Contains(t, actualOutput, "X")
}

func TestConsolePrinter_Print_NotSlice_NoMarshalSlice(t *testing.T) {
	var testError = new(bytes.Buffer)
	var testOutput = new(bytes.Buffer)
	var testPrinter = NewConsolePrinter(testError, testOutput)
	var testStruct = struct {
		A string `cli:"ColumnA"`
	}{
		A: "value",
	}

	assert.PanicsWithError(t, ErrorConsoleNoMarshal.Error(), func() { _ = testPrinter.Print(testStruct) })
	actualError := testError.String()
	actualOutput := testOutput.String()
	assert.Empty(t, actualError)
	assert.Empty(t, actualOutput)
}

func TestConsolePrinter_Print_NotSlice_NoSelector(t *testing.T) {
	var testError = new(bytes.Buffer)
	var testOutput = new(bytes.Buffer)
	var testPrinter = NewConsolePrinter(testError, testOutput)
	var testStruct = &simpleTwoColumns{
		A: "valueA",
		B: "valueB",
	}

	assert.NoError(t, testPrinter.Print(*testStruct))
	actualError := testError.String()
	actualOutput := testOutput.String()
	assert.Empty(t, actualError)
	assert.Contains(t, actualOutput, "COLUMN_A")
	assert.Contains(t, actualOutput, testStruct.A)
	assert.Contains(t, actualOutput, "COLUMN_B")
	assert.Contains(t, actualOutput, testStruct.B)
}

func TestConsolePrinter_Print_NotSlice_NothingStruct(t *testing.T) {
	var testError = new(bytes.Buffer)
	var testOutput = new(bytes.Buffer)
	var testPrinter = NewConsolePrinter(testError, testOutput)

	assert.ErrorIs(t, testPrinter.Print(struct{ A string }{A: "value"}), ErrorConsoleNothing)
	actualError := testError.String()
	actualOutput := testOutput.String()
	assert.Empty(t, actualError)
	assert.Empty(t, actualOutput)
}

func TestJsonPrinter_Error(t *testing.T) {
	var expectedField = "expectedField"
	var expectedAllowed1 = "allowed1"
	var expectedAllowed2 = "allowed2"
	var expectedStatus = 400
	var expectedMessage = "expectedMessage"
	var testError = task.NewErrorBuilder().
		Field(expectedField).
		Allowed(expectedAllowed1, expectedAllowed2).
		Status(expectedStatus).
		Build(expectedMessage)
	var output = new(bytes.Buffer)
	var testPrinter = NewJsonPrinter(output)

	assert.NoError(t, testPrinter.Error(testError))
	actual := output.String()
	assert.Contains(t, actual, expectedField)
	assert.Contains(t, actual, expectedAllowed1)
	assert.Contains(t, actual, expectedAllowed2)
	assert.Contains(t, actual, strconv.Itoa(expectedStatus))
	assert.Contains(t, actual, expectedMessage)
}

func TestJsonPrinter_Print(t *testing.T) {
	var expectedValue = "value"
	var testOutput = new(bytes.Buffer)
	var testPrinter = &jsonPrinter{
		basePrinter: basePrinter{
			output: testOutput,
		},
		indent: lang.Ref(" "),
	}

	assert.NoError(t, testPrinter.Print(struct{ A string }{A: expectedValue}))
	actual := testOutput.String()
	assert.NotContains(t, actual, "  ")
	assert.Contains(t, actual, fmt.Sprintf("\"A\": \"%s\"", expectedValue))
}

func TestJsonPrinter_Print_MarshalErr(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testPrinter = NewJsonPrinter(testOutput)

	assert.Error(t, testPrinter.Print(math.NaN()))
	assert.Empty(t, testOutput.String())
}

func TestJsonPrinter_Print_WriterErr(t *testing.T) {
	var expectedError = errors.New("expected")
	var testOutput = &io2.StubWriter{WriteError: expectedError}
	var testPrinter = NewJsonPrinter(testOutput)

	assert.ErrorIs(t, testPrinter.Print(struct{ A string }{A: "value"}), expectedError)
}

func TestHandleError(t *testing.T) {
	var testPrinter = &stubPrinter{}
	var testError = errors.New("expected")

	HandleError(testPrinter)(testError)
	assert.Same(t, testPrinter.printerErr, testError)
}

func TestHandleError_Panics(t *testing.T) {
	var expectedPanic = errors.New("panic")
	var testPrinter = &stubPrinter{err: expectedPanic}
	var testError = errors.New("expected")

	assert.Panics(t, func() { HandleError(testPrinter)(testError) })
}

func TestHandlePrint(t *testing.T) {
	var testPrinter = &stubPrinter{}
	var testStruct = struct{ A string }{A: "test"}

	HandlePrint(testPrinter)(testStruct)
	assert.Equal(t, testPrinter.printerOut, testStruct)
}

func TestHandlePrint_Panics(t *testing.T) {
	var expectedPanic = errors.New("panic")
	var testPrinter = &stubPrinter{err: expectedPanic}
	var testStruct = struct{ A string }{A: "test"}

	assert.Panics(t, func() { HandlePrint(testPrinter)(testStruct) })
}
