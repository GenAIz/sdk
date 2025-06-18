package config

import (
	"os"
	"regexp"

	"github.com/spf13/cast"

	"genaiz.com/genaiz/lang/enumz"
	"genaiz.com/genaiz/lang/panicz"
)

// Validates is a type def for a function taking a value returning valid or not valid
type Validates func(value any) bool

// validators provides Validates for various validation needs
type validators struct {
	DirCreated Validates
	DirExists  Validates
	DomainName Validates
	FileExists Validates
	FolderName Validates
}

var (
	Validation = &validators{
		DirCreated: validateDirCreated,
		DirExists:  validateDirExists,
		DomainName: stringMatches(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`),
		FileExists: validateFileExists,
		FolderName: stringMatches(`^[a-zA-Z0-9\-._\/]+$`),
	}
)

// AllFromEnumerated creates a Validates which ensures all values from a slice are valid for the provided enum
func AllFromEnumerated[T string](enum *enumz.EnumType[T]) Validates {
	return func(value any) bool {
		var list = cast.ToStringSlice(value)

		for _, arch := range list {
			if !enum.IsValid(cast.ToString(arch)) {
				return false
			}
		}

		return true
	}
}

// AnyOfEnumerated creates a Validates which indicates if a value is a string member of the provided enum
func AnyOfEnumerated[T string](enum *enumz.EnumType[T]) Validates {
	return func(value any) bool {
		return enum.IsValid(cast.ToString(value))
	}
}

// Optionally only applies the validator function if the value is not nil or the empty string
func Optionally(validates Validates) Validates {
	return func(value any) bool {
		if value != nil && cast.ToString(value) != "" {
			return validates(value)
		}

		return true
	}
}

func stringMatches(pattern string) Validates {
	return func(value any) bool {
		var matched, err = regexp.Match(pattern, []byte(cast.ToString(value)))

		panicz.PanicIfError(err)
		return matched
	}
}

// validateDirCreated creates a path if it does not exist and returns true
func validateDirCreated(path any) bool {
	var stringPath = path.(string)

	if s, _ := os.Stat(stringPath); s == nil {
		var err = os.MkdirAll(stringPath, 0750)

		return err == nil
	}

	return true
}

// validateDirExists validates that a path exists and is a directory
func validateDirExists(path any) bool {
	var s, _ = os.Stat(path.(string))

	return s != nil && s.IsDir()
}

// validateFileExists validates that a path exists and is not a directory
func validateFileExists(path any) bool {
	var s, _ = os.Stat(path.(string))

	return s != nil && !s.IsDir()
}
