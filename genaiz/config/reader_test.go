package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

func TestDataLinksReader_GetDataLink(t *testing.T) {
	var expectedLink = &broker.DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testReader = &DataLinksReader{current: []broker.DataLink{*expectedLink}}

	assert.Nil(t, testReader.GetDataLink("notOem", "notHandle", "notVersion"))
	assert.Equal(t, expectedLink, testReader.GetDataLink(expectedLink.Oem, expectedLink.Handle, expectedLink.Version))
}

func TestDataLinksReader_GetLatest(t *testing.T) {
	var expectedHandle = "expectedHandle"
	var expectedOem = "expectedOem"
	var expectedLinks = []broker.DataLink{
		{
			Handle:  expectedHandle,
			Oem:     expectedOem,
			Version: "firstVersion",
		},
		{
			Handle:  expectedHandle,
			Oem:     expectedOem,
			Version: "secondVersion",
		},
	}
	var testReader = &DataLinksReader{}

	assert.Nil(t, testReader.GetLatest("notOem", "notHandle"))
	testReader.WithCurrent(expectedLinks)
	assert.Equal(t, &expectedLinks[1], testReader.GetLatest(expectedOem, expectedHandle))
}

func TestDataLinksReader_Read(t *testing.T) {
	var testInput = filepath.Join(t.TempDir(), "Genaiz.yaml")
	var testLedger = NewBuilder().WithViper(viper.New()).Build()
	var testReader = &DataLinksReader{}
	var fd *os.File
	var err error

	if fd, err = os.Create(testInput); err == nil {
		var testBytes []byte
		var expectedHandle = "expectedHandle"
		var testLinks = map[string][]broker.DataLink{
			"datalinks": {
				{
					Handle: expectedHandle,
				},
			},
		}

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(testLinks); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				filez.CloseSilently(fd)
				testLedger.InitLogging()
				testReader.Read(testLedger, testInput)
				assert.NotEmpty(t, testReader.current)
				assert.Equal(t, expectedHandle, testReader.current[0].Handle)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestDataLinkReader_Read_FileNotFound(t *testing.T) {
	var testDir = filepath.Join(t.TempDir(), "notExist")
	var testLedger = NewBuilder().WithViper(viper.New()).Build()
	var testReader = &DataLinksReader{}

	testLedger.InitLogging()
	testReader.Read(testLedger, testDir)
	assert.Empty(t, testReader.current)
}

func TestDataLinksReader_ReadFile_MarshallError(t *testing.T) {
	var testInput = filepath.Join(t.TempDir(), "Genaiz.yaml")
	var testReader = &DataLinksReader{}
	var fd *os.File
	var err error

	if fd, err = os.Create(testInput); err == nil {
		var testBytes []byte
		var testHandle = "testHandle"
		var testLinks = map[string]string{
			"datalinks": testHandle,
		}

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(testLinks); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				var actual []broker.DataLink

				filez.CloseSilently(fd)
				actual, err = testReader.ReadFile(testInput)
				assert.Nil(t, actual)
				assert.Error(t, err)
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

	if _, err = os.Create(filepath.Join(testDir, expectedName+"."+shared.ConfigTypeYaml)); err == nil {
		if err = os.MkdirAll(filepath.Join(testDir, "notAFunction"), 0666); err == nil {
			testLedger.Logger = logrus.New()
			assert.Empty(t, testReader.FindFunctionValues())
			return
		}
	}

	assert.NoError(t, err)
}

func TestSolutionReader_Find(t *testing.T) {
	var testName = "testName"
	var expectedName = testName + "." + shared.ConfigTypeYaml
	var testDir = t.TempDir()
	var testReader = &SolutionReader{
		BaseReader: BaseReader{
			configPath: testDir,
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, expectedName)); err == nil {
		filez.CloseSilently(fd)
		assert.NoError(t, testReader.Find(testName))
		assert.Equal(t, shared.ConfigTypeYaml, testReader.GetConfigType())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestSolutionReader_FindInvalidPath(t *testing.T) {
	var testReader = &SolutionReader{
		BaseReader: BaseReader{
			configPath: "/_not_exist",
		},
	}

	assert.Error(t, testReader.Find("testName"))
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
