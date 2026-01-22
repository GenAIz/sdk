package broker

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

type stubDataLinkClient struct {
	client
	findDataLink         *DataLink
	findDataLinkError    error
	listDataLink         []DataLink
	listDataLinkError    error
	publishDataLink      *DataLink
	publishDataLinkError error
}

func (sdc *stubDataLinkClient) FindDataLink(oem, handle, version string) (*DataLink, error) {
	_, _, _ = oem, handle, version
	return sdc.findDataLink, sdc.findDataLinkError
}

func (sdc *stubDataLinkClient) ListDataLinks(oem, handle string, flags int) ([]DataLink, error) {
	_, _, _ = oem, handle, flags
	return sdc.listDataLink, sdc.listDataLinkError
}

func (sdc *stubDataLinkClient) PublishDataLink(publishLinkData *DataLink) (*DataLink, error) {
	sdc.publishDataLink = publishLinkData
	return sdc.publishDataLink, sdc.publishDataLinkError
}

type stubDataLinkWriter struct {
	dataLink         *DataLink
	dataLinks        []DataLink
	dataLinksRemoved []*DataLink
	propSpecRemoved  *PropSpec
	writeError       error
	writeOutput      string
}

func (sw *stubDataLinkWriter) BuildDataLinks() (string, []DataLink) {
	return "", sw.dataLinks
}

func (sw *stubDataLinkWriter) GetDataLink(string, string, string) *DataLink {
	return sw.dataLink
}

func (sw *stubDataLinkWriter) GetLatest(string, string) *DataLink {
	return sw.dataLink

}

func (sw *stubDataLinkWriter) SyncDataLinks() []*DataLink {
	return sw.dataLinksRemoved
}

func (sw *stubDataLinkWriter) WithDataLink(dataLink *DataLink) DataLinkWriter {
	if dataLink != nil {
		if i := slices.IndexFunc(sw.dataLinks, func(link DataLink) bool {
			return link.IsEqual(dataLink.Handle, dataLink.Oem, dataLink.Version)
		}); i < 0 {
			sw.dataLinks = append(sw.dataLinks, *dataLink)
		}
	}

	return sw
}

func (sw *stubDataLinkWriter) WithPropSpecRemoved(propSpec *PropSpec) DataLinkWriter {
	sw.propSpecRemoved = propSpec
	return sw
}

func (sw *stubDataLinkWriter) Write(output string) error {
	sw.writeOutput = output
	return sw.writeError
}

func TestDataLinkParams_ToFqdn(t *testing.T) {
	var actualOem, actualHandle, actualVersion string
	var testParams = &DataLinkParams{}
	var testLink = &DataLink{
		Oem:     "expected.oem",
		Handle:  "expected-handle",
		Version: "expected.version.final",
	}

	actualOem, actualHandle, actualVersion = testParams.ToFqdn()
	assert.Empty(t, actualOem)
	assert.Empty(t, actualHandle)
	assert.Empty(t, actualVersion)
	testParams.DataLink = testLink
	actualOem, actualHandle, actualVersion = testParams.ToFqdn()
	assert.Equal(t, testLink.Oem, actualOem)
	assert.Equal(t, testLink.Handle, actualHandle)
	assert.Equal(t, testLink.Version, actualVersion)
}

func TestDataLinkParams_ToString(t *testing.T) {
	var testParams = &DataLinkParams{}
	var testLink = &DataLink{
		Oem:     "expected.oem",
		Handle:  "expected-handle",
		Version: "expected.version.final",
	}

	assert.Empty(t, testParams.ToString())
	testParams.DataLink = testLink
	assert.Equal(t, fmt.Sprintf("%s/%s:%s", testLink.Oem, testLink.Handle, testLink.Version), testParams.ToString())
}

func TestDataLinkParams_findDataLink_invalid(t *testing.T) {
	var testParams = &DataLinkParams{}
	var actual, err = testParams.findDataLink(nil)

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errDataLinkInvalid)
	testParams.DataLink = &DataLink{}
	actual, err = testParams.findDataLink(nil)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errDataLinkInvalid)
	testParams.DataLink.Oem = "oem"
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errDataLinkInvalid)
}

func TestDataLinkParams_findDataLink_latest(t *testing.T) {
	var testWriter = &stubDataLinkWriter{}
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Oem:    "oem",
			Handle: "handle",
		},
	}
	var expectedLink = &DataLink{}
	var actual, err = testParams.findDataLink(testWriter)

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errDataLinkNotFound)
	testWriter.dataLink = expectedLink
	actual, err = testParams.findDataLink(testWriter)
	assert.Same(t, expectedLink, actual)
	assert.NoError(t, err)
}

func TestDataLinkParams_findDataLink_version(t *testing.T) {
	var testWriter = &stubDataLinkWriter{}
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Oem:     "oem",
			Handle:  "handle",
			Version: "version",
		},
	}
	var expectedLink = &DataLink{}
	var actual, err = testParams.findDataLink(testWriter)

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errDataLinkNotFound)
	testWriter.dataLink = expectedLink
	actual, err = testParams.findDataLink(testWriter)
	assert.Same(t, expectedLink, actual)
	assert.NoError(t, err)
}

func TestDataLinkParams_exists(t *testing.T) {
	var testWriter = &stubDataLinkWriter{}
	var testParams = &DataLinkParams{}
	var testLink = &DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
	}

	assert.False(t, testParams.exists(testWriter))
	testParams.DataLink = testLink
	assert.False(t, testParams.exists(testWriter))
	testWriter.dataLink = testLink
	assert.True(t, testParams.exists(testWriter))
}

func TestDataLinkParam_isEqual(t *testing.T) {
	var testParams = &DataLinkParams{}
	var testLink = &DataLink{}

	assert.True(t, testParams.isEqual(nil))
	assert.False(t, testParams.isEqual(testLink))
	testParams.DataLink = &DataLink{}
	testParams.Oem = "oem"
	assert.False(t, testParams.isEqual(testLink))
	testLink.Oem = testParams.Oem
	assert.True(t, testParams.isEqual(testLink))
	testParams.Handle = "handle"
	assert.False(t, testParams.isEqual(testLink))
	testLink.Handle = testParams.Handle
	assert.True(t, testParams.isEqual(testLink))
	testParams.Version = "version"
	assert.False(t, testParams.isEqual(testLink))
	testLink.Version = testParams.Version
	assert.True(t, testParams.isEqual(testLink))
}

func TestDataLinkParam_isValid(t *testing.T) {
	var testParams = &DataLinkParams{}

	assert.False(t, testParams.isValid())
	testParams.DataLink = &DataLink{}
	testParams.Oem = "oem"
	assert.False(t, testParams.isValid())
	testParams.Handle = "handle"
	assert.False(t, testParams.isValid())
	testParams.Version = "version"
	assert.True(t, testParams.isValid())
}

func TestNewDataLinkCreateTask(t *testing.T) {
	var testTask = NewDataLinkCreateTask(nil)

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewDataLinkEditTask(t *testing.T) {
	var testTask = NewDataLinkEditTask(nil)

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewDataLinkFindTask(t *testing.T) {
	var testTask = NewDataLinkFindTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewDataLinkPublishTask(t *testing.T) {
	var testTask = NewDataLinkPublishTask(nil)

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func Test_handleDataLinkAvailableError_InternalEmpty(t *testing.T) {
	var testState = &task.State{
		Internal: []DataLink{},
		Logger:   logrus.New(),
	}
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Oem:    "expectedOem",
			Handle: "expectedHandle",
		},
	}

	if actual := handleDataLinkAvailableError(testParams, testState); actual != nil {
		actualText := actual.Error()
		assert.Contains(t, actualText, testParams.Oem)
		assert.Contains(t, actualText, testParams.Handle)
	} else {
		assert.Fail(t, "should have an error")
	}
}

func Test_handleDataLinkAvailableError_InternalInvalid(t *testing.T) {
	var testState = &task.State{
		Internal: "invalid",
		Logger:   logrus.New(),
	}
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Oem:    "expectedOem",
			Handle: "expectedHandle",
		},
	}

	if actual := handleDataLinkAvailableError(testParams, testState); actual != nil {
		actualText := actual.Error()
		assert.Contains(t, actualText, testParams.Oem)
		assert.Contains(t, actualText, testParams.Handle)
	} else {
		assert.Fail(t, "should have an error")
	}
}

func Test_handleDataLinkAvailableError_InternalNil(t *testing.T) {
	assert.NoError(t, handleDataLinkAvailableError(&DataLinkParams{}, &task.State{}))
}

func Test_handleDataLinkCreateContext(t *testing.T) {
	var testDir = t.TempDir()
	var testHandle = "testHandle"
	var testOem = "testOem"
	var testVersion = "testVersion"
	var testState = &task.State{Logger: logrus.New()}
	var testWriter = &stubDataLinkWriter{}
	var testParams = &DataLinkParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: testDir,
			ConfigName:   "Genaiz",
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
		},
		DataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}

	err := handleDataLinkCreateContext(testWriter, testParams, testState)
	assert.True(t, errorz.IsPathError(err))
}

func Test_handleDataLinkCreateContext_DirError(t *testing.T) {
	var testDir = t.TempDir()
	var testHandle = "testHandle"
	var testOem = "testOem"
	var testVersion = "testVersion"
	var testState = &task.State{Logger: logrus.New()}
	var testWriter = &stubDataLinkWriter{}
	var testParams = &DataLinkParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: testDir,
			ConfigName:   "Genaiz",
			ConfigType:   lang.Ref(shared.ConfigTypeToml),
		},
		DataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}

	if err := os.MkdirAll(filepath.Join(testDir, "Genaiz.toml"), 0750); err == nil {
		assert.ErrorIs(t, handleDataLinkCreateContext(testWriter, testParams, testState), shared.ErrorConfigFileInvalid)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleDataLinkCreateContext_FileExists(t *testing.T) {
	var testDir = t.TempDir()
	var testHandle = "testHandle"
	var testOem = "testOem"
	var testVersion = "testVersion"
	var testState = &task.State{Logger: logrus.New()}
	var testWriter = &stubDataLinkWriter{}
	var testParams = &DataLinkParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: testDir,
			ConfigName:   "Genaiz",
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
		},
		DataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}

	if fd, err := os.Create(filepath.Join(testDir, "Genaiz.json")); err == nil {
		filez.CloseSilently(fd)
		assert.ErrorIs(t, handleDataLinkCreateContext(testWriter, testParams, testState), shared.ErrorConfigFileExists)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleDataLinkCreateContext_LinkExistsError(t *testing.T) {
	var testHandle = "testHandle"
	var testOem = "testOem"
	var testVersion = "testVersion"
	var testState = &task.State{Logger: logrus.New()}
	var testWriter = &stubDataLinkWriter{
		dataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}

	assert.ErrorIs(t, handleDataLinkCreateContext(testWriter, testParams, testState), errDataLinkExists)
}

func Test_handleDataLinkCreateContext_OutputKnown(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.Nil(t, handleDataLinkCreateContext(nil, &DataLinkParams{}, testState))
}

func Test_handleDataLinkCreateComplete(t *testing.T) {
	var testDir = t.TempDir()
	var expectedOutput = filepath.Join(testDir, "Genaiz.yaml")
	var testWriter = &stubDataLinkWriter{
		dataLinksRemoved: []*DataLink{nil,
			{
				Handle:  "testHandle",
				Oem:     "testOem",
				Version: "oldVersion",
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedOutput,
	}
	var testParams = &DataLinkParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: testDir,
			ConfigName:   "Genaiz",
			ConfigType:   lang.Ref(shared.ConfigTypeYaml),
		},
		DataLink: &DataLink{
			Handle:  "testHandle",
			Oem:     "testOem",
			Version: "testVersion",
		},
	}

	assert.NoError(t, handleDataLinkCreateComplete(testWriter, testParams, testState))
	assert.Equal(t, expectedOutput, testWriter.writeOutput)
	assert.Equal(t, 1, len(testWriter.dataLinks))
	assert.Equal(t, testParams.DataLink, &testWriter.dataLinks[0])
}

func Test_handleDataLinkCreateComplete_OutputUnknown(t *testing.T) {
	assert.ErrorIs(t, handleDataLinkCreateComplete(nil, &DataLinkParams{}, &task.State{}), shared.ErrorConfigFileInvalid)
}

func Test_handleDataLinkCreateComplete_WriteError(t *testing.T) {
	var testDir = t.TempDir()
	var expectedError = errors.New("expected")
	var testWriter = &stubDataLinkWriter{writeError: expectedError}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: testDir,
	}
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Handle:  "testHandle",
			Oem:     "testOem",
			Version: "testVersion",
		},
	}

	assert.ErrorIs(t, handleDataLinkCreateComplete(testWriter, testParams, testState), expectedError)
	assert.Equal(t, testDir, testWriter.writeOutput)
	assert.Equal(t, 1, len(testWriter.dataLinks))
	assert.Equal(t, testParams.DataLink, &testWriter.dataLinks[0])
}

func Test_handleDataLinkCreateIncomplete(t *testing.T) {
	var testOutput = filepath.Join(t.TempDir(), "Genaiz.yaml")
	var err error

	if err = os.MkdirAll(filepath.Dir(testOutput), 0444); err == nil {
		var testState = &task.State{
			Logger: logrus.New(),
			Error:  &os.PathError{},
			Output: testOutput,
		}
		var stat os.FileInfo

		assert.NoError(t, handleDataLinkCreateIncomplete(&DataLinkParams{}, testState))

		if stat, err = os.Stat(testOutput); err == nil {
			assert.False(t, stat.IsDir())
			return
		}
	}

	assert.NoError(t, err)
}

func Test_handleDataLinkCreateIncomplete_CreateError(t *testing.T) {
	var testOutput = filepath.Join(t.TempDir(), "subfolder", "Genaiz.yaml")
	var err error

	if err = os.MkdirAll(filepath.Dir(testOutput), 0444); err == nil {
		var testState = &task.State{
			Logger: logrus.New(),
			Error:  &os.PathError{},
			Output: testOutput,
		}

		err = handleDataLinkCreateIncomplete(&DataLinkParams{}, testState)
		assert.Error(t, err)
		assert.Equal(t, err, testState.Error)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleDataLinkCreateIncomplete_IrrecuperableError(t *testing.T) {
	var testState = &task.State{
		Error:  errDataLinkExists,
		Output: "output",
	}

	assert.ErrorIs(t, handleDataLinkCreateIncomplete(&DataLinkParams{}, testState), testState.Error)
}

func Test_handleDataLinkCreateIncomplete_OutputUnknown(t *testing.T) {
	var testState = &task.State{
		Error: errors.New("expected"),
	}

	assert.ErrorIs(t, handleDataLinkCreateIncomplete(&DataLinkParams{}, testState), testState.Error)
}

func Test_handleDataLinkCreatePretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var removedLink = &DataLink{Handle: "removed"}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var testWriter = &stubDataLinkWriter{
		dataLinksRemoved: []*DataLink{nil, removedLink},
		dataLinks: []DataLink{
			{
				Handle: "added",
			},
			*removedLink,
		},
	}
	var testParams = &DataLinkParams{}

	defer patch.Unpatch()
	assert.NoError(t, handleDataLinkCreatePretend(testWriter, testParams, testState))
	assert.NotEmpty(t, patch.CalledWith)
	assert.Equal(t, testWriter.dataLinks[0].Handle, cast.ToStringSlice(patch.CalledWith)[2])
	assert.Equal(t, testState.Output, cast.ToStringSlice(patch.CalledWith)[3])

}

func Test_handleDataLinkCreatePretend_IrrecuperableError(t *testing.T) {
	var testState = &task.State{
		Error:  errDataLinkExists,
		Output: "output",
	}

	assert.ErrorIs(t, handleDataLinkCreatePretend(nil, &DataLinkParams{}, testState), testState.Error)
}

func Test_handleDataLinkCreatePretend_NoChanges(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var testWriter = &stubDataLinkWriter{}
	var testParams = &DataLinkParams{}

	assert.NoError(t, handleDataLinkCreatePretend(testWriter, testParams, testState))
}

func Test_handleDataLinkCreatePretend_OutputUnknown(t *testing.T) {
	var testState = &task.State{
		Error: errors.New("expected"),
	}

	assert.ErrorIs(t, handleDataLinkCreatePretend(nil, &DataLinkParams{}, testState), testState.Error)
}

func Test_handleDataLinkEditContext(t *testing.T) {
	var testDir = t.TempDir()
	var testHandle = "testHandle"
	var testOem = "testOem"
	var testVersion = "testVersion"
	var testState = &task.State{Logger: logrus.New()}
	var testWriter = &stubDataLinkWriter{
		dataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}
	var testParams = &DataLinkParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: testDir,
			ConfigName:   "Genaiz",
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
		},
		DataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}

	if fd, err := os.Create(filepath.Join(testDir, "Genaiz.json")); err == nil {
		filez.CloseSilently(fd)
		assert.NoError(t, handleDataLinkEditContext(testWriter, testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleDataLinkEditContext_DirError(t *testing.T) {
	var testDir = t.TempDir()
	var testHandle = "testHandle"
	var testOem = "testOem"
	var testVersion = "testVersion"
	var testState = &task.State{Logger: logrus.New()}
	var testWriter = &stubDataLinkWriter{
		dataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}
	var testParams = &DataLinkParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: testDir,
			ConfigName:   "Genaiz",
			ConfigType:   lang.Ref(shared.ConfigTypeToml),
		},
		DataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}

	if err := os.MkdirAll(filepath.Join(testDir, "Genaiz.toml"), 0750); err == nil {
		assert.ErrorIs(t, handleDataLinkEditContext(testWriter, testParams, testState), shared.ErrorConfigFileInvalid)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleDataLinkEditContext_LinkNotFoundError(t *testing.T) {
	var testHandle = "testHandle"
	var testOem = "testOem"
	var testVersion = "testVersion"
	var testState = &task.State{Logger: logrus.New()}
	var testWriter = &stubDataLinkWriter{}
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Handle:  testHandle,
			Oem:     testOem,
			Version: testVersion,
		},
	}

	assert.ErrorIs(t, handleDataLinkEditContext(testWriter, testParams, testState), errDataLinkNotFound)
}

func Test_handleDataLinkEditContext_OutputKnown(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.Nil(t, handleDataLinkEditContext(nil, &DataLinkParams{}, testState))
}

func Test_handleDataLinkEditComplete(t *testing.T) {
	var testDir = t.TempDir()
	var expectedOutput = filepath.Join(testDir, "Genaiz.yaml")
	var testWriter = &stubDataLinkWriter{
		dataLinksRemoved: []*DataLink{nil,
			{
				Handle:  "testHandle",
				Oem:     "testOem",
				Version: "oldVersion",
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedOutput,
	}
	var testParams = &DataLinkParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: testDir,
			ConfigName:   "Genaiz",
			ConfigType:   lang.Ref(shared.ConfigTypeYaml),
		},
		DataLink: &DataLink{
			Handle:  "testHandle",
			Oem:     "testOem",
			Version: "testVersion",
		},
	}

	assert.NoError(t, handleDataLinkEditComplete(testWriter, testParams, testState))
	assert.Equal(t, expectedOutput, testWriter.writeOutput)
	assert.Equal(t, 1, len(testWriter.dataLinks))
	assert.Equal(t, testParams.DataLink, &testWriter.dataLinks[0])
}

func Test_handleDataLinkEditComplete_OutputUnknown(t *testing.T) {
	assert.ErrorIs(t, handleDataLinkEditComplete(nil, &DataLinkParams{}, &task.State{}), shared.ErrorConfigFileInvalid)
}

func Test_handleDataLinkEditComplete_WriteError(t *testing.T) {
	var testDir = filepath.Join(t.TempDir(), "Genaiz.yaml")
	var expectedError = errors.New("expected")
	var testWriter = &stubDataLinkWriter{writeError: expectedError}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: testDir,
	}
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Handle:  "testHandle",
			Oem:     "testOem",
			Version: "testVersion",
		},
	}

	assert.ErrorIs(t, handleDataLinkEditComplete(testWriter, testParams, testState), expectedError)
	assert.Equal(t, testDir, testWriter.writeOutput)
	assert.Equal(t, 1, len(testWriter.dataLinks))
	assert.Equal(t, testParams.DataLink, &testWriter.dataLinks[0])
}

func Test_handleDataLinkEditPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHandle = "edited"
	var editedLink = &DataLink{Handle: expectedHandle}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var testWriter = &stubDataLinkWriter{
		dataLinksRemoved: []*DataLink{nil, editedLink},
		dataLinks: []DataLink{
			{
				Handle: "someOtherHandle",
			},
			{
				Handle: expectedHandle,
			},
		},
	}
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Handle: expectedHandle,
		},
	}

	defer patch.Unpatch()
	assert.NoError(t, handleDataLinkEditPretend(testWriter, testParams, testState))
	assert.NotEmpty(t, patch.CalledWith)
	assert.Equal(t, expectedHandle, cast.ToStringSlice(patch.CalledWith)[2])
	assert.Equal(t, testState.Output, cast.ToStringSlice(patch.CalledWith)[3])
}

func Test_handleDataLinkEditPretend_IrrecuperableError(t *testing.T) {
	var testState = &task.State{
		Error:  errDataLinkNotFound,
		Output: "output",
	}

	assert.ErrorIs(t, handleDataLinkEditPretend(nil, &DataLinkParams{}, testState), testState.Error)
}

func Test_handleDataLinkEditPretend_NoChanges(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output.yaml",
	}
	var testWriter = &stubDataLinkWriter{}
	var testParams = &DataLinkParams{}

	assert.NoError(t, handleDataLinkEditPretend(testWriter, testParams, testState))
}

func Test_handleDataLinkEditPretend_OutputUnknown(t *testing.T) {
	var testState = &task.State{
		Error: errors.New("expected"),
	}

	assert.ErrorIs(t, handleDataLinkEditPretend(nil, &DataLinkParams{}, testState), testState.Error)
}

func Test_handleDataLinkFindContext(t *testing.T) {
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Oem:     "oem",
			Handle:  "handle",
			Version: "version",
		},
	}

	assert.NoError(t, handleDataLinkFindContext(testParams, &task.State{}))
}

func Test_handleDataLinkFindContext_InternalSet(t *testing.T) {
	var testParams = &DataLinkParams{
		DataLink: &DataLink{
			Oem:     "oem",
			Handle:  "handle",
			Version: "version",
		},
	}
	var testTask = &task.State{
		Internal: []DataLink{},
	}

	assert.NoError(t, handleDataLinkFindContext(testParams, testTask))
}

func Test_handleDataLinkFindContext_NoValid(t *testing.T) {
	assert.Error(t, handleDataLinkFindContext(&DataLinkParams{}, &task.State{}))
}

func Test_handleDataLinkFindComplete(t *testing.T) {
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "expected-version"
	var expectedLinks = []DataLink{
		{
			Id:      int64(37),
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
			Flags:   DataLinkFlags.Active,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: &DataLink{
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataLinkClient{
			listDataLink: expectedLinks,
		}, nil
	}

	assert.NoError(t, handleDataLinkFindComplete(testParams, testState))
	assert.Equal(t, expectedLinks[0], testState.Internal)
}

func Test_handleDataLinkFindComplete_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, expectedError, handleDataLinkFindComplete(testParams, &task.State{}))
}

func Test_handleDataLinkFindComplete_ListLinkError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataLinkClient{
			listDataLinkError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataLinkFindComplete(testParams, testState), expectedError)
}

func Test_handleDataLinkFindComplete_ListLinkInactive(t *testing.T) {
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "expected-version"
	var expectedLinks = []DataLink{
		{
			Id:      int64(37),
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
			Flags:   0,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: &DataLink{
			Oem:     "oem",
			Handle:  "handle",
			Version: "version",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataLinkClient{
			listDataLink: expectedLinks,
		}, nil
	}

	if actual := handleDataLinkFindComplete(testParams, testState); actual != nil {
		actualText := actual.Error()
		assert.Contains(t, actualText, expectedOem)
		assert.Contains(t, actualText, expectedHandle)
		assert.Contains(t, actualText, expectedVersion)
	} else {
		assert.Fail(t, "should have an error")
	}
}

func Test_handleDataLinkFindComplete_ListLinkUnavailable(t *testing.T) {
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "expected-version"
	var expectedLinks = []DataLink{
		{
			Id:      int64(37),
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
			Flags:   DataLinkFlags.Active,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: &DataLink{
			Oem:     "oem",
			Handle:  "handle",
			Version: "version",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataLinkClient{
			listDataLink: expectedLinks,
		}, nil
	}

	if actual := handleDataLinkFindComplete(testParams, testState); actual != nil {
		actualText := actual.Error()
		assert.Contains(t, actualText, expectedOem)
		assert.Contains(t, actualText, expectedHandle)
		assert.Contains(t, actualText, expectedVersion)
	} else {
		assert.Fail(t, "should have an error")
	}
}

func Test_handleDataLinkFindComplete_SkipValidation(t *testing.T) {
	var testParams = &DataLinkParams{NoValidation: true}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleDataLinkFindComplete(testParams, testState))
}

func Test_handleDataLinkFindPretend(t *testing.T) {
	var expectedToken = "expectedToken"
	var expectedHostAddr = "expectedHost"
	var testParams = &DataLinkParams{
		Broker: Broker{
			HostAddr: expectedHostAddr,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w

	defer func() {
		clientFactory.Get = restoredFactory
		os.Stdout = stdoutRestore
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		var clt = &stubFunctionClient{}

		clt.AuthToken = expectedToken
		clt.HostAddr = expectedHostAddr
		return clt, nil
	}

	assert.NoError(t, handleDataLinkFindPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testParams.HostAddr)
	assert.Contains(t, output, expectedToken)
}

func Test_handleDataLinkFindPretend_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataLinkFindPretend(testParams, &task.State{}), expectedError)
}

func Test_handleDataLinkPublishContext(t *testing.T) {
	var expectedToken = "expectedToken"
	var expectedHostAddr = "expectedHost"
	var testLink = &DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testWriter = &stubDataLinkWriter{
		dataLink: testLink,
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: testLink,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		var clt = &stubDataLinkClient{
			// On any error we can not tell whether the publish call is valid or not
			findDataLinkError: errors.New("notExpected"),
		}

		clt.AuthToken = expectedToken
		clt.HostAddr = expectedHostAddr
		return clt, nil
	}

	assert.NoError(t, handleDataLinkPublishContext(testWriter, testParams, testState))
}

func Test_handleDataLinkPublishContext_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testLink = &DataLink{
		Handle: "testHandle",
		Oem:    "testOem",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testWriter = &stubDataLinkWriter{
		dataLink: testLink,
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: testLink,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataLinkPublishContext(testWriter, testParams, testState), expectedError)
}

func Test_handleDataLinkPublishContext_DataLinkConflict(t *testing.T) {
	var expectedToken = "expectedToken"
	var expectedHostAddr = "expectedHost"
	var testLink = &DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testWriter = &stubDataLinkWriter{
		dataLink: testLink,
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: testLink,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		var clt = &stubDataLinkClient{
			findDataLink: testLink,
		}

		clt.AuthToken = expectedToken
		clt.HostAddr = expectedHostAddr
		return clt, nil
	}

	assert.ErrorIs(t, handleDataLinkPublishContext(testWriter, testParams, testState), errDataLinkConflict)
}

func Test_handleDataLinkPublishContext_DataLinkError(t *testing.T) {
	var testTask = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleDataLinkPublishContext(nil, &DataLinkParams{}, testTask), errDataLinkInvalid)
}

func Test_handleDataLinkPublishContext_OutputKnown(t *testing.T) {
	var testTask = &task.State{Output: "output"}

	assert.NoError(t, handleDataLinkPublishContext(nil, &DataLinkParams{}, testTask))
}

func Test_handleDataLinkPublishComplete(t *testing.T) {
	var expectedToken = "expectedToken"
	var expectedHostAddr = "expectedHost"
	var expectedLink = &DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testWriter = &stubDataLinkWriter{
		dataLink: expectedLink,
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: expectedLink,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		var clt = &stubDataLinkClient{
			publishDataLink: expectedLink,
		}

		clt.AuthToken = expectedToken
		clt.HostAddr = expectedHostAddr
		return clt, nil
	}

	assert.NoError(t, handleDataLinkPublishComplete(testWriter, testParams, testState))
	assert.Equal(t, expectedLink, testState.Internal)
}

func Test_handleDataLinkPublishComplete_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testLink = &DataLink{
		Handle: "testHandle",
		Oem:    "testOem",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testWriter = &stubDataLinkWriter{
		dataLink: testLink,
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: testLink,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataLinkPublishComplete(testWriter, testParams, testState), expectedError)
}

func Test_handleDataLinkPublishComplete_DataLinkError(t *testing.T) {
	var testTask = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleDataLinkPublishComplete(nil, &DataLinkParams{}, testTask), errDataLinkInvalid)
}

func Test_handleDataLinkPublishComplete_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "found a DataLink id",
	}

	assert.ErrorIs(t, handleDataLinkPublishComplete(nil, &DataLinkParams{}, testState), errDataLinkExists)
}

func Test_handleDataLinkPublishComplete_PublishError(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedToken = "expectedToken"
	var expectedHostAddr = "expectedHost"
	var testLink = &DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testWriter = &stubDataLinkWriter{
		dataLink: testLink,
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: testLink,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		var clt = &stubDataLinkClient{
			publishDataLinkError: expectedError,
		}

		clt.AuthToken = expectedToken
		clt.HostAddr = expectedHostAddr
		return clt, nil
	}

	assert.ErrorIs(t, handleDataLinkPublishComplete(testWriter, testParams, testState), expectedError)
}

func Test_handleDataLinkPublishPretend(t *testing.T) {
	var expectedLink = &DataLink{
		Handle:  "testHandle",
		Oem:     "testOem",
		Version: "testVersion",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testWriter = &stubDataLinkWriter{
		dataLink: expectedLink,
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: expectedLink,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataLinkClient{}, nil
	}

	assert.NoError(t, handleDataLinkPublishPretend(testWriter, testParams, testState))
}

func Test_handleDataLinkPublishPretend_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testLink = &DataLink{
		Handle: "testHandle",
		Oem:    "testOem",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testWriter = &stubDataLinkWriter{
		dataLink: testLink,
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: testLink,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataLinkPublishPretend(testWriter, testParams, testState), expectedError)
}

func Test_handleDataLinkPublishPretend_DataLinkError(t *testing.T) {
	var testTask = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleDataLinkPublishPretend(nil, &DataLinkParams{}, testTask), errDataLinkInvalid)
}

func Test_handleDataLinkPublishPretend_OutputKnown(t *testing.T) {
	var testTask = &task.State{Output: "some existing id"}

	assert.NoError(t, handleDataLinkPublishPretend(nil, &DataLinkParams{}, testTask))
}

func Test_pretendPropSpec(t *testing.T) {
	var args []string
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {
		for _, arg := range a {
			args = append(args, cast.ToString(arg))
		}
	})
	var pretender = shared.NewConfigPretender("output.yaml")
	var testPropSpecs = []PropSpec{
		{
			Key:   "key1",
			Name:  "name1",
			Type:  "STRING",
			Value: "default1",
		},
		{
			Key:         "key2",
			Name:        "name2",
			Description: "desc2",
			Type:        "ENUM",
			Values:      []string{"VALUE1", "VALUE2"},
		},
	}

	defer patch.Unpatch()
	pretendPropSpec(pretender, "rootKey", 0, testPropSpecs)
	assert.Contains(t, args, testPropSpecs[0].Key)
	assert.Contains(t, args, testPropSpecs[1].Key)
	assert.Contains(t, args, testPropSpecs[0].Name)
	assert.Contains(t, args, testPropSpecs[1].Name)
	assert.Contains(t, args, testPropSpecs[0].Type)
	assert.Contains(t, args, testPropSpecs[1].Type)
	assert.Contains(t, args, testPropSpecs[0].Value)
	assert.Contains(t, args, testPropSpecs[1].Description)
}
