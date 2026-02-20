package main

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/version"
)

func Test_main(t *testing.T) {
	var stdoutRestore = os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	os.Args = []string{"genaiz", "--version"}
	main()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	assert.Equal(t, fmt.Sprintf("genaiz version %s\n", version.GetVersion()), string(out))
}

func Test_main_errorExit(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})

	defer patch.Unpatch()
	os.Args = []string{"genaiz", "unknownCommand"}
	main()
	assert.True(t, patch.Called)
	assert.Equal(t, patch.CalledWith, 1)
}
