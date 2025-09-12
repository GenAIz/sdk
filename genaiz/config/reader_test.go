package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestBaseReader_GetSolutionFile(t *testing.T) {
	var expectedName = "name"
	var expectedPath = "path"
	var expectedType = "type"
	var testReader = &BaseReader{
		configName: expectedName,
		configPath: expectedPath,
		configType: expectedType,
	}
	var actual = testReader.GetSolutionFile()

	assert.Equal(t, filepath.Join(expectedPath, expectedName+"."+expectedType), actual)
}

func TestBaseReader_GetSolutionPath(t *testing.T) {
	var expectedBasePath = "base"
	var expectedPath = "path"
	var testReader = &BaseReader{
		configPath: expectedBasePath,
	}
	var actual = testReader.GetSolutionPath(expectedPath)

	assert.Equal(t, filepath.Join(expectedBasePath, expectedPath), actual)
}

func TestBaseReader_Read(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedName = "Test"
	var expectedType = shared.ConfigTypeYaml

	if fd, err := os.Create(filepath.Join(expectedPath, expectedName+"."+expectedType)); err == nil {
		var testReader = &BaseReader{
			configPath: filepath.Dir(fd.Name()),
			configName: expectedName,
		}

		defer filez.CloseSilently(fd)
		_, err = testReader.Read(expectedType)
		assert.NoError(t, err)
		assert.Equal(t, expectedPath, testReader.configPath)
		assert.Equal(t, expectedType, testReader.configType)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestBaseReader_Read_FileNotExist(t *testing.T) {
	var testReader = &BaseReader{}
	var _, err = testReader.Read(shared.ConfigTypeYaml)

	assert.Error(t, err)
	assert.Empty(t, testReader.configPath)
	assert.Empty(t, testReader.configType)
}

func TestBaseReader_ReadFile(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedName = "Test"
	var expectedType = shared.ConfigTypeYaml
	var err error
	var fd *os.File

	if fd, err = os.Create(filepath.Join(expectedPath, expectedName+"."+expectedType)); err == nil {
		var testStruct = struct{ Solution string }{Solution: "invalid"}
		var testReader = &BaseReader{
			configPath: filepath.Dir(fd.Name()),
			configName: expectedName,
		}
		var data []byte

		if data, err = yaml.Marshal(testStruct); err == nil {
			if _, err = fd.Write(data); err == nil {
				var actual *broker.Solution

				defer filez.CloseSilently(fd)
				actual, err = testReader.Read(expectedType)
				assert.Error(t, err)
				assert.Empty(t, actual)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestSolutionReader_FindFunctionValues(t *testing.T) {
	var testDir = t.TempDir()
	var expectedName = "Test"
	var testLedger = NewLedger()
	var testReader = &SolutionReader{
		BaseReader: BaseReader{
			configName: expectedName,
			configPath: testDir,
			configType: shared.ConfigTypeYaml,
		},
		ledger: testLedger,
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, expectedName+".yaml")); err == nil {
		filez.CloseSilently(fd)

		if err = os.MkdirAll(filepath.Join(testDir, "function"), 0750); err == nil {
			if _, err = os.Create(filepath.Join(testDir, "function", expectedName+".yaml")); err == nil {
				defer filez.CloseSilently(fd)
				testLedger.Logger = logrus.New()
				assert.Equal(t, 1, len(testReader.FindFunctionValues()))
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestSolutionReader_FindFunction_EmptyDir(t *testing.T) {
	var testDir = t.TempDir()
	var testReader = &SolutionReader{
		BaseReader: BaseReader{
			configPath: testDir,
		},
	}

	assert.Empty(t, testReader.FindFunctionValues())
}

func TestSolutionReader_FindFunctionValues_InvalidPath(t *testing.T) {
	var testReader = &SolutionReader{}

	assert.Empty(t, testReader.FindFunctionValues())
}

func TestSolutionReader_FindFunction_NoConfigFile(t *testing.T) {
	var testDir = t.TempDir()
	var expectedName = "Test"
	var testLedger = NewLedger()
	var testReader = &SolutionReader{
		BaseReader: BaseReader{
			configName: expectedName,
			configPath: testDir,
			configType: shared.ConfigTypeYaml,
		},
		ledger: testLedger,
	}
	var err error

	if _, err = os.Create(filepath.Join(testDir, expectedName+".yaml")); err == nil {
		if err = os.MkdirAll(filepath.Join(testDir, "notAFunction"), 0666); err == nil {
			testLedger.Logger = logrus.New()
			assert.Empty(t, testReader.FindFunctionValues())
			return
		}
	}

	assert.NoError(t, err)
}

func TestSolutionReader_GetSolution(t *testing.T) {
	var testReader = &SolutionReader{}

	assert.Empty(t, testReader.GetSolution())
}

func TestSolutionReader_GetVersion(t *testing.T) {
	var testReader = &SolutionReader{}
	var expectedVersion = "version"

	assert.Empty(t, testReader.GetVersion())
	testReader.current = &broker.Solution{Version: expectedVersion}
	assert.Equal(t, expectedVersion, testReader.GetVersion())
}

func TestSolutionReader_Read(t *testing.T) {
	var expectedPath = t.TempDir()
	var expectedName = "Test"
	var expectedType = shared.ConfigTypeYaml

	if fd, err := os.Create(filepath.Join(expectedPath, expectedName+"."+expectedType)); err == nil {
		var testReader = &SolutionReader{
			BaseReader: BaseReader{
				configPath: filepath.Dir(fd.Name()),
				configName: expectedName,
			},
		}

		defer filez.CloseSilently(fd)
		assert.NoError(t, testReader.Read(expectedType))
		assert.Equal(t, expectedPath, testReader.configPath)
		assert.Equal(t, expectedType, testReader.configType)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestSolutionReader_Read_FileNotExist(t *testing.T) {
	var testReader = &SolutionReader{}
	var err = testReader.Read(shared.ConfigTypeYaml)

	assert.Error(t, err)
	assert.Empty(t, testReader.configPath)
	assert.Empty(t, testReader.configType)
}

func TestSolutionReader_WithConfigPath(t *testing.T) {
	var testReader = &SolutionReader{}
	var expectedPath = "path"

	assert.Equal(t, expectedPath, testReader.WithConfigPath(expectedPath).configPath)
}

func TestSolutionReader_WithConfigType(t *testing.T) {
	var testReader = &SolutionReader{}
	var expectedType = shared.ConfigTypeYaml

	assert.Equal(t, expectedType, testReader.WithConfigType(expectedType).configType)
}

func TestNewSolutionReader(t *testing.T) {
	var testLedger = NewLedger()
	var testReader = NewSolutionReader(testLedger)

	assert.Equal(t, testLedger.ConfigName, testReader.configName)
}
