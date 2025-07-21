package dirz

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoIfPathExist(t *testing.T) {
	var cwd, _ = os.Getwd()
	var expectedError = errors.New("expected")
	var testCall = func() error {
		return expectedError
	}

	assert.EqualValues(t, expectedError, DoIfPathExist(cwd, testCall))
}

func TestDoIfPathExist_NotExist(t *testing.T) {
	var testCall = func() error {
		return errors.New("expected")
	}

	assert.NoError(t, DoIfPathExist("/_not_exist", testCall))
}
