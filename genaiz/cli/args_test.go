package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/config"
)

func TestArgsHostAndPort(t *testing.T) {
	var expectedHost = "0.0.0.0"
	var expectedPort = 65535
	var host, port, err = ArgsHostAndPort(fmt.Sprintf("%s:%d", expectedHost, expectedPort))

	assert.Equal(t, expectedHost, host)
	assert.Equal(t, expectedPort, port)
	assert.NoError(t, err)
}

func TestArgsHostAndPort_Any(t *testing.T) {
	var expectedHost = "*"
	var host, port, err = ArgsHostAndPort(expectedHost)

	assert.Equal(t, expectedHost, host)
	assert.Equal(t, 0, port)
	assert.NoError(t, err)
}

func TestArgsHostAndPort_PortRange(t *testing.T) {
	var host, port, err = ArgsHostAndPort("a:65536")

	assert.Empty(t, host)
	assert.Equal(t, -1, port)
	assert.Error(t, err)
	assert.Equal(t, "invalid port", err.Error())
}

func TestArgsHostAndPort_PortStr(t *testing.T) {
	var host, port, err = ArgsHostAndPort("a:a")

	assert.Empty(t, host)
	assert.Equal(t, -1, port)
	assert.Error(t, err)
}

func TestArgsHostAndPort_SplitError(t *testing.T) {
	var host, port, err = ArgsHostAndPort("a")

	assert.Empty(t, host)
	assert.Equal(t, -1, port)
	assert.Error(t, err)
}

func TestArgsOptionalFolder(t *testing.T) {
	var validator = ArgsOptionalFolder("test", 1, config.Validation.Handle)

	assert.NoError(t, validator(nil, []string{"valid"}))
}

func TestArgsOptionalFolder_InvalidDir(t *testing.T) {
	if fd, err := os.CreateTemp("", "genaiz-args-invalid-file*.yaml"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		var validator = ArgsOptionalFolder("test", 1, config.Validation.Handle)

		err = validator(nil, []string{fd.Name()})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fd.Name())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestArgsOptionalFolder_InvalidName(t *testing.T) {
	var expectedType = "testType"
	var expectedInvalid = "..invalid"
	var validator = ArgsOptionalFolder(expectedType, 1, config.Validation.Handle)
	var err = validator(nil, []string{expectedInvalid})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedType)
	assert.Contains(t, err.Error(), expectedInvalid)
}

func TestArgsSingleFolder(t *testing.T) {
	var expectedArg = "arg"
	var expectedArg2 = "arg2"

	assert.Equal(t, expectedArg, ArgsOptionalSingle([]string{expectedArg}))
	assert.Equal(t, strings.Join([]string{expectedArg, expectedArg2}, " "), ArgsOptionalSingle([]string{expectedArg, expectedArg2}))
}

func TestArgsFolderValidator_MultiArgs(t *testing.T) {
	var expectedType = "testType"
	var validator = ArgsOptionalFolder(expectedType, 1, config.Validation.Handle)

	assert.NoError(t, validator(nil, []string{"valid", "not--a-folder-argument"}))
}

func TestArgsFolderValidator_NoArgs(t *testing.T) {
	var expectedType = "testType"
	var validator = ArgsOptionalFolder(expectedType, 1, config.Validation.Handle)

	assert.NoError(t, validator(nil, []string{}))
}
