package enumz

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"genaiz.com/genaiz/lang/panicz"
)

type EnumFactory[T string] interface {
	FromString(label string) (EnumType[T], error)
}

type EnumType[T string] struct {
	Values []T
}

// AllFromStrings returns a slice of enum values corresponding to the provided labels if they are all valid, otherwise an error is returned
func (et EnumType[T]) AllFromStrings(labels *[]string) ([]T, error) {
	var result []T

	panicz.RequiresNotNil("labels", labels)

	for i, l := range *labels {
		if value, err := et.FromString(l); err == nil {
			result = append(result, *value)
		} else {
			return nil, fmt.Errorf("label position %d is invalid", i)
		}
	}

	return result, nil
}

// FromString returns the equivalent type to the provided label or an error if it doesn't match any enum values
func (et EnumType[T]) FromString(label string) (*T, error) {
	var valueMap = map[string]*T{}

	for _, t := range et.Values {
		valueMap[string(t)] = &t
	}

	if value, ok := valueMap[label]; ok {
		return value, nil
	}

	return nil, errors.New("enum not found")
}

// FromOrdinal returns the equivalent type to its enumerated index or an error if there is no such enum value
func (et EnumType[T]) FromOrdinal(ordinal int) (*T, error) {
	if ordinal >= len(et.Values) {
		return nil, errors.New("enum not found")
	}

	return &et.Values[ordinal], nil
}

// IsValid indicates whether a label corresponds to an enum value
func (et EnumType[T]) IsValid(label string) bool {
	var stringValues []string

	for _, t := range et.Values {
		stringValues = append(stringValues, string(t))
	}

	return slices.Contains(stringValues, strings.ToLower(label))
}

// NewEnumType builds a new enum type using the provided variadic array of string aliases
func NewEnumType[T string](enums ...T) *EnumType[T] {
	return &EnumType[T]{Values: enums}
}
