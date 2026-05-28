package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
)

var (
	ColorHeader   = color.New(color.Bold)
	ColorSelected = color.New(color.FgBlue)
	ColorPlain    = color.New(color.ResetBold)

	ErrorConsoleNoMarshal = task.NewError("interface{} does not implement MarshalSlice")
	ErrorConsoleNothing   = task.NewError("interface{} with no columns to print")

	StdPrinter     printerFactory
	StdCliTag      = "cli"
	StdCliRedGreen = "redGreen"
	StdCliSelected = "selected"
)

// MarshalSlice is an interface that can be inferred for when a struct needs to convert its values into a slice of strings
type MarshalSlice interface {
	MarshalSlice() ([]string, error)
}

type Printer interface {
	Error(interface{}) error

	Print(interface{}) error
}

type basePrinter struct {
	error  io.Writer
	output io.Writer
}

type consoleColumn struct {
	header   string
	hidden   bool
	redGreen func([]string) bool
	selected func([]string) string
}

type consolePrinter struct {
	basePrinter
}

func (cp consolePrinter) Error(obj interface{}) error {
	lang.HandleExit(obj)
	return nil
}

func (cp consolePrinter) Print(obj interface{}) error {
	if fieldObj := cp.getFieldObj(obj); fieldObj != nil {
		var fieldValue, fieldType, fieldMap = cp.getFieldDesc(fieldObj)

		if len(fieldMap) > 0 {
			var columns = cp.getColumns(fieldValue, fieldType, fieldMap)
			var out = tabwriter.NewWriter(cp.output, 1, 1, 3, ' ', 0)
			var columnValues, err = cp.getValueTable(obj)

			panicz.PanicIfError(err)
			columnHeaders, columnSelect := cp.getHeadersAndSelector(columns)
			_, _ = fmt.Fprintf(out, "%s\n", strings.Join(columnHeaders, "\t"))

			for _, row := range columnValues {
				var values = make([]string, len(row))
				var selectedValue string
				var rowColor = ColorPlain

				if columnSelect != nil {
					selectedValue = columnSelect.selected(row)

					if selectedValue != "" {
						rowColor = ColorSelected
					}
				}

				for j, value := range row {
					var column = columns[j]

					if column.selected != nil {
						values[j] = rowColor.Sprintf("%s", selectedValue)
					} else if column.redGreen != nil {
						if column.redGreen(row) {
							values[j] = color.RedString(value)
						} else {
							values[j] = color.GreenString(value)
						}
					} else if !column.hidden {
						values[j] = rowColor.Sprintf("%s", value)
					}
				}

				_, _ = fmt.Fprintf(out, "%s\n", strings.Join(values, "\t"))
			}

			_ = out.Flush()
			return nil
		}

		return ErrorConsoleNothing
	}

	return nil
}

func (cp consolePrinter) getColumns(val reflect.Value, typ reflect.Type, fieldMap map[string]int) []consoleColumn {
	var results []consoleColumn

	for i := 0; i < val.NumField(); i++ {
		var ft = typ.Field(i)
		var cTag = ft.Tag.Get(StdCliTag)

		if cTag != "" {
			var rgTag = ft.Tag.Get(StdCliRedGreen)
			var sTag = ft.Tag.Get(StdCliSelected)
			var rgFunc func([]string) bool
			var sdFunc func([]string) string

			if rgTag != "" {
				rgFunc = func(values []string) bool {
					if j, ok := fieldMap[rgTag]; ok && j < len(values) {
						return strings.EqualFold("yes", values[j])
					}

					return false
				}
			}

			if sTag != "" {
				sdFunc = func(values []string) string {
					if j, ok := fieldMap[cTag]; ok && j < len(values) {
						if strings.EqualFold("yes", values[j]) {
							return sTag
						}
					}

					return ""
				}
			}

			results = append(results, consoleColumn{
				header:   cTag,
				hidden:   strings.HasSuffix(cTag, "noShow"),
				redGreen: rgFunc,
				selected: sdFunc,
			})
		}
	}

	return results
}

func (cp consolePrinter) getHeadersAndSelector(columns []consoleColumn) ([]string, *consoleColumn) {
	var result []string
	var selector *consoleColumn

	for _, col := range columns {
		if !col.hidden {
			result = append(result, ColorHeader.Sprintf("%s", strings.ToUpper(col.header)))
		}

		if col.selected != nil {
			selector = &col
		}
	}

	return result, selector
}

func (cp consolePrinter) getFieldDesc(obj interface{}) (reflect.Value, reflect.Type, map[string]int) {
	var noShowSuffix = ",noShow"
	var v = reflect.ValueOf(obj)
	var t = reflect.TypeOf(obj)
	var result = map[string]int{}
	var j = 0

	for i := 0; i < v.NumField(); i++ {
		var ft = t.Field(i)
		var cTag = ft.Tag.Get(StdCliTag)

		if cTag != "" {
			var suffixIndex = strings.LastIndex(cTag, noShowSuffix)
			var cKey string

			if suffixIndex >= 0 {
				cKey = cTag[:suffixIndex]
			} else {
				cKey = cTag
			}

			result[cKey] = j
			j += 1
		}
	}

	return v, t, result
}

func (cp consolePrinter) getFieldObj(obj interface{}) interface{} {
	var v = reflect.ValueOf(obj)

	if v.Kind() == reflect.Slice {
		if v.Len() > 0 {
			return v.Index(0).Interface()
		}

		return nil
	}

	return obj
}

func (cp consolePrinter) getValueRow(val reflect.Value) ([]string, error) {
	var err error

	if ms, ok := val.Interface().(MarshalSlice); ok {
		var values []string

		if values, err = ms.MarshalSlice(); err == nil {
			return values, nil
		}

		return nil, err
	}

	return nil, ErrorConsoleNoMarshal
}

func (cp consolePrinter) getValueTable(obj interface{}) ([][]string, error) {
	var val = reflect.ValueOf(obj)
	var result [][]string
	var row []string
	var err error

	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if row, err = cp.getValueRow(val.Index(i)); err == nil {
				result = append(result, row)
			} else {
				return nil, err
			}
		}
	} else if row, err = cp.getValueRow(val); err == nil {
		result = append(result, row)
	} else {
		return nil, err
	}

	return result, nil
}

func NewConsolePrinter(errOut, output io.Writer) Printer {
	return consolePrinter{
		basePrinter: basePrinter{
			error:  errOut,
			output: output,
		},
	}
}

type jsonPrinter struct {
	basePrinter
	indent *string
}

func (jp jsonPrinter) Error(obj interface{}) error {
	// Not different with a json marshaller
	return jp.Print(obj)
}

func (jp jsonPrinter) Print(obj interface{}) error {
	var bytes []byte
	var err error

	if bytes, err = json.MarshalIndent(obj, "", jp.getIndent()); err == nil {
		if _, err = jp.output.Write(bytes); err == nil {
			_, _ = fmt.Fprintf(jp.output, "\n")
			return nil
		}
	}

	return err
}

func (jp jsonPrinter) getIndent() string {
	if jp.indent == nil {
		return "   "
	}

	return *jp.indent
}

func NewJsonPrinter(output io.Writer) Printer {
	return &jsonPrinter{
		basePrinter: basePrinter{
			output: output,
		},
	}
}

type printerFactory struct {
	consoleFactory func(io.Writer, io.Writer) Printer
	jsonFactory    func(io.Writer) Printer
}

func (pf printerFactory) JsonOrConsole(ledger *config.Ledger, jsonOption *config.BoolOption) Printer {
	if ledger.GetBool(jsonOption) {
		return pf.jsonFactory(os.Stdout)
	}

	return pf.consoleFactory(os.Stderr, os.Stdout)
}

func init() {
	StdPrinter = printerFactory{
		consoleFactory: NewConsolePrinter,
		jsonFactory:    NewJsonPrinter,
	}
}
