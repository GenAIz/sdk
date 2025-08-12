package config

import (
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz/lang/enumz"
)

// Validates is a type def for a function taking a value returning valid or not valid
type Validates func(value any) bool

// validators provides Validates for various validation needs
type validators struct {
	DirCreated    Validates
	DirExists     Validates
	DomainName    Validates
	FileExists    Validates
	FolderName    Validates
	Handle        Validates
	Oem           Validates
	Name          Validates
	Version       Validates
	VersionNumber Validates
}

var (
	nameStrings   = stringMatches(`^[a-zA-Z0-9]+(?:[a-zA-Z0-9\-._][a-zA-Z0-9]+)*$`)
	versionNumber = stringMatches(`^(?:[1-9][0-9]*|0)$`)

	Validation = validators{
		DirCreated:    validateDirCreated,
		DirExists:     validateDirExists,
		DomainName:    stringMatches(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`),
		FileExists:    validateFileExists,
		FolderName:    stringMatches(`^[a-zA-Z0-9\-._\/]+$`),
		Handle:        AllOf(nameStrings, stringMaxLength(128)),
		Oem:           AllOf(nameStrings, stringMaxLength(128)),
		Name:          stringMaxLength(255),
		Version:       validateVersion,
		VersionNumber: versionNumber,
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

// AllOf creates a Validates which ensures all validation rules specified are valid for the provided value
func AllOf(validates ...Validates) Validates {
	return func(value any) bool {
		for _, validate := range validates {
			if !validate(value) {
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

// stringMatches returns a Validates which indicates whether a string value matches the specified regex
func stringMatches(pattern string) Validates {
	var regex = regexp.MustCompile(pattern)

	return func(value any) bool {
		return regex.Match([]byte(cast.ToString(value)))
	}
}

// stringMaxLength returns a Validates which indicates whether a string value exceeds the specified length
func stringMaxLength(length int) Validates {
	return func(value any) bool {
		return len(cast.ToString(value)) <= length
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

// validateVersion validates a string against the SemVer 2.0.0 semantic form, but without any specific additional tags (-RC, -Final, etc..)
func validateVersion(version any) bool {
	var stringVersion = version.(string)

	if stringVersion != "" {
		var parts = strings.SplitN(stringVersion, ".", 3)

		return len(parts) == 3 &&
			versionNumber(parts[0]) &&
			versionNumber(parts[1]) &&
			versionNumber(parts[2])
	}

	return false
}
