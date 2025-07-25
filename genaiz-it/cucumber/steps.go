package cucumber

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"

	messages "github.com/cucumber/messages/go/v24"
	"github.com/iancoleman/strcase"
)

var (
	regexArgs         = regexp.MustCompile(`(".*?")`)
	regexPunctuation  = regexp.MustCompile(`[:;,.\-_\\/|]`)
	regexVariables    = regexp.MustCompile(`(<.*?>)`)
	supportedKeywords = []string{"given", "and", "when", "then"}
)

type Steps interface {
	Visit([]*messages.Step) error
}

type steps struct {
	definitions interface{}
}

func (s *steps) getMethodArgs(text string) []string {
	var matches = regexArgs.FindAllString(text, -1)

	if len(matches) > 0 {
		var result []string

		for _, match := range matches {
			if match != "" {
				result = append(result, match[1:len(match)-1])
			}
		}

		return result
	}

	return nil
}

func (s *steps) getMethodName(text string) string {
	var noArgs = regexArgs.ReplaceAll([]byte(text), []byte(" "))
	var noPunctuation = regexPunctuation.ReplaceAll(noArgs, nil)
	var noVariables = regexVariables.ReplaceAll(noPunctuation, nil)
	var allWords = strings.Split(string(noVariables), " ")
	var builder = strings.Builder{}

	for _, word := range allWords {
		builder.Write([]byte(strcase.ToCamel(word)))
	}

	return builder.String()
}

func (s *steps) getMethodTable(table *messages.DataTable) any {
	if table != nil {
		var rowsCount = len(table.Rows)

		if rowsCount == 2 {
			var headers = table.Rows[0].Cells
			var result = make(map[string]string)

			for i, cell := range table.Rows[1].Cells {
				var name = headers[i].Value

				if name != "" {
					result[name] = cell.Value
				}
			}

			return result
		} else if rowsCount > 2 {
			if len(table.Rows[0].Cells) == 1 {
				var result []string

				for _, row := range table.Rows {
					result = append(result, row.Cells[0].Value)
				}

				return result
			}
		}

		if rowsCount != 0 {
			panic(errors.New("unsupported data table format"))
		}
	}

	return nil
}

func (s *steps) Visit(steps []*messages.Step) error {
	var err error

	for _, step := range steps {
		var cleanKeyword = strings.ToLower(strings.TrimSpace(step.Keyword))

		if slices.Contains(supportedKeywords, cleanKeyword) {
			var methodName = s.getMethodName(step.Text)
			var methodArgs = s.getMethodArgs(step.Text)
			var methodTable = s.getMethodTable(step.DataTable)
			var method = reflect.ValueOf(s.definitions).MethodByName(methodName)
			var params []reflect.Value

			if method.IsValid() {
				for _, arg := range methodArgs {
					params = append(params, reflect.ValueOf(arg))
				}

				if methodTable != nil {
					params = append(params, reflect.ValueOf(methodTable))
				}

				method.Call(params)
			} else {
				return fmt.Errorf("step definition for step line [%d] is not valid", step.Location.Line)
			}
		}
	}

	return err
}

func NewSteps(definitions interface{}) Steps {
	return &steps{
		definitions: definitions,
	}
}
