package locker

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type stubSourceClient struct {
	stubClient
	createDataSourceError    error
	createDataSourceInstance *broker.DataLinkInstance
	createDataSourceProps    map[string]string
	createDataSourceResult   *broker.DataLinkInstance
	listDataSourceError      error
	listDataSourceResult     []broker.DataLinkInstance
	updateDataSourceError    error
	updateDataSourceInstance *broker.DataLinkInstance
	updateDataSourceProps    map[string]string
	updateDataSourceResult   *broker.DataLinkInstance
}

func (ssc *stubSourceClient) CreateDataSource(instance *broker.DataLinkInstance, props map[string]string) (*broker.DataLinkInstance, error) {
	ssc.createDataSourceInstance = instance
	ssc.createDataSourceProps = props
	return ssc.createDataSourceResult, ssc.createDataSourceError
}

func (ssc *stubSourceClient) ListDataSources() ([]broker.DataLinkInstance, error) {
	return ssc.listDataSourceResult, ssc.listDataSourceError
}

func (ssc *stubSourceClient) UpdateDataSource(instance *broker.DataLinkInstance, props map[string]string) (*broker.DataLinkInstance, error) {
	ssc.updateDataSourceInstance = instance
	ssc.updateDataSourceProps = props
	return ssc.updateDataSourceResult, ssc.updateDataSourceError
}

func TestNewSourceAddTask(t *testing.T) {
	var testTask = NewSourceAddTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.Nil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewSourceFindTask(t *testing.T) {
	var testTask = NewSourceFindTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewSourcePublishTask(t *testing.T) {
	var testTask = NewSourcePublishTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewSourceSyncTask(t *testing.T) {
	var testTask = NewSourceSyncTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewSourceUpdateTask(t *testing.T) {
	var testTask = NewSourceUpdateTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func Test_handleSourceAddComplete(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceAddParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: "testHandle",
			}

			assert.NoError(t, handleSourceAddComplete(testParams, testState))
			assert.NotEmpty(t, testState.Reports)
			assert.Contains(t, testState.Reports[0], testParams.SourceHandle)
			assert.Contains(t, testState.Reports[0], testParams.LockerPath)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceAddComplete_ChaChaError(t *testing.T) {
	// The only way to cause a ChaCha error is to fail add by providing the wrong password
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase = memguard.NewEnclave([]byte("notTheRightPassword"))

		if _, err = writeTestLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceAddParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: "testHandle",
			}

			assert.ErrorIs(t, handleSourceAddComplete(testParams, testState), errorLockerPassFailed)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceAddComplete_LockerAddError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testHandle",
		}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceAddParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: testLink.LockerHandle,
			}

			assert.Error(t, handleSourceAddComplete(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceAddComplete_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Output: "known",
			Logger: logrus.New(),
		}
		var testParams = &SourceAddParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			Broker: broker.Broker{
				AuthFile: testAuthFile,
				HostAddr: testUrl,
			},
		}

		assert.Error(t, handleSourceAddComplete(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceAddComplete_LockerWriteError(t *testing.T) {
	// The write will fail after a successful add if the file's permissions are no longer writeable
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceAddParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: "testHandle",
			}

			if err = os.Chmod(testLockerPath, 0400); err == nil {
				assert.Error(t, handleSourceAddComplete(testParams, testState))
				return
			}
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceAddComplete_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &SourceAddParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleSourceAddComplete(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourceAddComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleSourceAddComplete(&SourceAddParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleSourceAddComplete_UnfoldError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestJsonError(testLockerPath); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceAddParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: "someSource",
			}

			assert.Error(t, handleSourceAddComplete(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceAddContext(t *testing.T) {
	var testPath = filepath.Join(t.TempDir(), "locker.bin")
	var testState = &task.State{
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				broker.PropSpec{
					Key: "myKey",
				},
			},
		},
	}
	var testParams = &SourceAddParams{
		BaseParams: BaseParams{
			LockerPath: testPath,
			Passphrase: memguard.NewEnclave([]byte("test")),
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		defer filez.CloseSilently(fd)

		assert.NoError(t, handleSourceAddContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceAddContext_CheckOutput(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleSourceAddContext(&SourceAddParams{}, testState))
}

func Test_handleSourceAddContext_NoPassphrase(t *testing.T) {
	var testPath = filepath.Join(t.TempDir(), "locker.bin")
	var testState = &task.State{
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				broker.PropSpec{
					Key: "myKey",
				},
			},
		},
	}
	var testParams = &SourceAddParams{
		BaseParams: BaseParams{
			LockerPath: testPath,
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		defer filez.CloseSilently(fd)

		assert.ErrorIs(t, handleSourceAddContext(testParams, testState), errorLockerPassFailed)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceAddContext_NoVarSpecs(t *testing.T) {
	assert.ErrorIs(t, handleSourceAddContext(&SourceAddParams{}, &task.State{}), errorLockerDataLinkEmpty)
}

func Test_handleSourceAddContext_UnreadableLocker(t *testing.T) {
	var testPath = filepath.Join(t.TempDir(), "locker.bin")
	var testState = &task.State{
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				broker.PropSpec{
					Key: "myKey",
				},
			},
		},
	}
	var testParams = &SourceAddParams{
		BaseParams: BaseParams{
			LockerPath: testPath,
		},
	}

	assert.Error(t, handleSourceAddContext(testParams, testState))
}

func Test_handleSourceAddPretend(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLogger, testHook = test.NewNullLogger()
		var testState = &task.State{
			Output: "known",
			Logger: testLogger,
		}
		var testParams = &SourceAddParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			Broker: broker.Broker{
				AuthFile: testAuthFile,
				HostAddr: testUrl,
			},
		}

		testLogger.SetLevel(logrus.DebugLevel)
		assert.NoError(t, handleSourceAddPretend(testParams, testState))
		assert.Equal(t, 2, len(testHook.Entries))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceAddPretend_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleSourceAddPretend(&SourceAddParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleSourceAddPretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &SourceAddParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleSourceAddPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourceFindComplete(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testHandle",
			LinkOem:      "expectedOem",
			LinkHandle:   "expectedHandle",
			LinkVersion:  "expectedVersion",
		}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: testLink.LockerHandle,
				Logger: logrus.New(),
			}
			var testParams = &SourceFindParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
				SourceHandle: testLink.LockerHandle,
			}

			assert.NoError(t, handleSourceFindComplete(testParams, testState))
			assert.Equal(t, testLink.LinkOem, testParams.Oem)
			assert.Equal(t, testLink.LinkHandle, testParams.Handle)
			assert.Equal(t, testLink.LinkVersion, testParams.Version)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceFindComplete_ChaChaError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testHandle",
		}
		var testPassPhrase = memguard.NewEnclave([]byte("invalidPass"))

		if _, err = writeTestLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "invalidDataSource",
				Logger: logrus.New(),
			}
			var testParams = &SourceFindParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
				SourceHandle: testLink.LockerHandle,
			}

			assert.ErrorIs(t, handleSourceFindComplete(testParams, testState), errorLockerPassFailed)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceFindComplete_LockerLookupError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testHandle",
		}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "invalidDataSource",
				Logger: logrus.New(),
			}
			var testParams = &SourceFindParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
				SourceHandle: testLink.LockerHandle,
			}

			assert.Error(t, handleSourceFindComplete(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceFindComplete_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Output: "known",
			Logger: logrus.New(),
		}
		var testParams = &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},
		}

		assert.Error(t, handleSourceFindComplete(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceFindComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleSourceFindComplete(&SourceFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleSourceFindComplete_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &SourceFindParams{
		DataLinkParams: &broker.DataLinkParams{
			Broker: broker.Broker{
				AuthFile: filepath.Join(t.TempDir(), ".auth"),
				HostAddr: "hostAddr",
			},
		},
	}

	assert.ErrorIs(t, handleSourceFindComplete(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourceFindContext(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &SourceFindParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("test")),
		},
	}

	assert.NoError(t, handleSourceFindContext(testParams, testState))
}

func Test_handleSourceFindContext_FoundDataLink(t *testing.T) {
	var testParams = &SourceFindParams{
		DataLinkParams: &broker.DataLinkParams{
			DataLink: &broker.DataLink{
				Oem:     "expectedOem",
				Handle:  "expectedHandle",
				Version: "expectedVersion",
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleSourceFindContext(testParams, testState), errorLockerDataLinkFound)
	assert.Contains(t, testState.Output, testParams.Oem)
	assert.Contains(t, testState.Output, testParams.Handle)
	assert.Contains(t, testState.Output, testParams.Version)
}

func Test_handleSourceFindContext_OutputKnown(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleSourceFindContext(&SourceFindParams{}, testState))
}

func Test_handleSourceFindContext_NoPassphrase(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleSourceFindContext(&SourceFindParams{}, testState), errorLockerPassFailed)
}

func Test_handleSourceFindIncomplete(t *testing.T) {
	var testParams = &SourceFindParams{
		DataLinkParams: &broker.DataLinkParams{
			DataLink: &broker.DataLink{
				Oem:     "oem",
				Handle:  "handle",
				Version: "version",
			},
		},
	}

	var testState = &task.State{
		Error:  errorLockerDataLinkFound,
		Logger: logrus.New(),
	}

	assert.NoError(t, handleSourceFindIncomplete(testParams, testState))
	assert.Equal(t, 1, len(testState.Reports))
}

func Test_handleSourceFindIncomplete_TaskError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleSourceFindIncomplete(&SourceFindParams{}, testState), expectedError)
}

func Test_handleSourceFindPretend(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testHandle",
		}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testLogger, testHook = test.NewNullLogger()
			var testState = &task.State{
				Output: "known",
				Logger: testLogger,
			}
			var testParams = &SourceFindParams{
				BaseParams: BaseParams{
					LockerPath: filepath.Join(testDir, "locker.bin"),
					Passphrase: testPassPhrase,
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
			}

			testLogger.SetLevel(logrus.DebugLevel)
			assert.NoError(t, handleSourceFindPretend(testParams, testState))
			assert.Equal(t, 2, len(testHook.Entries))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceFindPretend_ChaChaError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testHandle",
		}
		var testPassPhrase = memguard.NewEnclave([]byte("invalidPass"))

		if _, err = writeTestLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "invalidDataSource",
				Logger: logrus.New(),
			}
			var testParams = &SourceFindParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
				SourceHandle: testLink.LockerHandle,
			}

			assert.ErrorIs(t, handleSourceFindPretend(testParams, testState), errorLockerPassFailed)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceFindPretend_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleSourceFindPretend(&SourceFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleSourceFindPretend_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Output: "known",
			Logger: logrus.New(),
		}
		var testParams = &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},
		}

		assert.Error(t, handleSourceFindPretend(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceFindPretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &SourceFindParams{
		DataLinkParams: &broker.DataLinkParams{
			Broker: broker.Broker{
				AuthFile: filepath.Join(t.TempDir(), ".auth"),
				HostAddr: "hostAddr",
			},
		},
	}

	assert.ErrorIs(t, handleSourceFindPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourcePublishContext(t *testing.T) {
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: broker.DataLinkInstance{},
	}
	var testParams = &SourcePublishParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("passphrase")),
		},
	}

	assert.ErrorIs(t, handleSourcePublishContext(testParams, testState), errorDataSourceExist)
}

func Test_handleSourcePublishContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleSourcePublishContext(&SourcePublishParams{}, testState))
}

func Test_handleSourcePublishContext_NoPassphrase(t *testing.T) {
	assert.ErrorIs(t, handleSourcePublishContext(&SourcePublishParams{}, &task.State{}), errorLockerPassFailed)
}

func Test_handleSourcePublishContext_NoStateInternal(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &SourcePublishParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("passphrase")),
		},
	}

	assert.ErrorIs(t, handleSourcePublishContext(testParams, testState), errorDataSourceLinkUnknown)
}

func Test_handleSourcePublishContext_NoStateLinkOrSource(t *testing.T) {
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: "notALinkOrSource",
	}
	var testParams = &SourcePublishParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("passphrase")),
		},
	}

	assert.ErrorIs(t, handleSourcePublishContext(testParams, testState), errorDataSourceLinkUnknown)
}

func Test_handleSourcePublishContext_StateDataLink(t *testing.T) {
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: broker.DataLink{},
	}
	var testParams = &SourcePublishParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("passphrase")),
		},
	}

	assert.NoError(t, handleSourcePublishContext(testParams, testState))
}

func Test_handleSourcePublishCreate(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testSource",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Internal: broker.DataLink{
					Id: new(int64(37)),
				},
			}
			var testParams = &SourcePublishParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: testLink.LockerHandle,

				client: &stubSourceClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					createDataSourceResult: &broker.DataLinkInstance{
						Id: new(int64(37)),
					},
				},
			}

			assert.NoError(t, handleSourcePublishCreate(testParams, testState))
			assert.NotEmpty(t, testState.Reports)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishCreate_CreateSourceError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testSource",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Internal: broker.DataLink{
					Id: new(int64(37)),
				},
			}
			var testParams = &SourcePublishParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: testLink.LockerHandle,

				client: &stubSourceClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					createDataSourceError: expectedError,
				},
			}

			assert.ErrorIs(t, handleSourcePublishCreate(testParams, testState), expectedError)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishCreate_GetPropError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase Enclave

		// Will just make an UnfoldError and GetProp will fail
		if testPassPhrase, err = writeTestJsonError(testLockerPath); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Internal: broker.DataLink{
					Id: new(int64(37)),
				},
			}
			var testParams = &SourcePublishParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: "someSource",
			}

			assert.Error(t, handleSourcePublishCreate(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishCreate_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Logger: logrus.New(),
			Internal: broker.DataLink{
				Id: new(int64(37)),
			},
		}
		var testParams = &SourcePublishParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			Broker: broker.Broker{
				AuthFile: testAuthFile,
				HostAddr: testUrl,
			},
		}

		assert.Error(t, handleSourcePublishCreate(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishCreate_NoStateInternal(t *testing.T) {
	assert.ErrorIs(t, handleSourcePublishCreate(&SourcePublishParams{}, &task.State{}), errorDataSourceUnknown)
}

func Test_handleSourcePublishCreate_SessionError(t *testing.T) {
	var testState = &task.State{
		Internal: broker.DataLink{},
	}
	var testParams = &SourcePublishParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleSourcePublishCreate(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourcePublishUpdate(t *testing.T) {
	var expectedError = errors.New("expected")
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testSource",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Error:  errorDataSourceExist,
				Internal: broker.DataLinkInstance{
					Id: new(int64(37)),
				},
			}
			var testParams = &SourcePublishParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: testLink.LockerHandle,

				client: &stubSourceClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					updateDataSourceResult: &broker.DataLinkInstance{
						Id: new(int64(37)),
					},
				},
			}

			assert.NoError(t, handleSourcePublishUpdate(testParams, testState), expectedError)
			assert.NotEmpty(t, testState.Reports)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishUpdate_GetPropsError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testSource",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Error:  errorDataSourceExist,
				Internal: broker.DataLinkInstance{
					Id: new(int64(37)),
				},
			}
			var testParams = &SourcePublishParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: "someSource",
			}

			assert.Error(t, handleSourcePublishUpdate(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishUpdate_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Logger: logrus.New(),
			Error:  errorDataSourceExist,
			Internal: broker.DataLinkInstance{
				Id: new(int64(37)),
			},
		}
		var testParams = &SourcePublishParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			Broker: broker.Broker{
				AuthFile: testAuthFile,
				HostAddr: testUrl,
			},
		}

		assert.Error(t, handleSourcePublishUpdate(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishUpdate_SessionError(t *testing.T) {
	var testState = &task.State{
		Error:    errorDataSourceExist,
		Internal: broker.DataLinkInstance{},
	}
	var testParams = &SourcePublishParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleSourcePublishUpdate(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourcePublishUpdate_SourceUnknown(t *testing.T) {
	var testState = &task.State{
		Error: errorDataSourceExist,
	}

	assert.ErrorIs(t, handleSourcePublishUpdate(&SourcePublishParams{}, testState), errorDataSourceUnknown)
}

func Test_handleSourcePublishUpdate_StateError(t *testing.T) {
	var testError = errors.New("expected")
	var testState = &task.State{
		Error: testError,
	}

	assert.ErrorIs(t, handleSourcePublishUpdate(&SourcePublishParams{}, testState), testError)
}

func Test_handleSourcePublishUpdate_UpdateSourceError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testSource",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Error:  errorDataSourceExist,
				Internal: broker.DataLinkInstance{
					Id: new(int64(37)),
				},
			}
			var testParams = &SourcePublishParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: testLink.LockerHandle,

				client: &stubSourceClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					updateDataSourceError: expectedError,
				},
			}

			assert.ErrorIs(t, handleSourcePublishUpdate(testParams, testState), expectedError)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishPretend(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testSource",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Error:  errorDataSourceExist,
				Internal: broker.DataLinkInstance{
					Id:         new(int64(37)),
					DataLinkId: new(int64(42)),
				},
			}
			var testParams = &SourcePublishParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: testLink.LockerHandle,

				client: &stubSourceClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					createDataSourceResult: &broker.DataLinkInstance{
						Id: new(int64(37)),
					},
				},
			}

			assert.NoError(t, handleSourcePublishPretend(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishPretend_NoStateError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testSource",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Internal: broker.DataLink{
					Id: new(int64(37)),
				},
			}
			var testParams = &SourcePublishParams{
				BaseParams: BaseParams{
					LockerPath: testLockerPath,
					Passphrase: testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
				SourceHandle: testLink.LockerHandle,

				client: &stubSourceClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					createDataSourceResult: &broker.DataLinkInstance{
						Id: new(int64(37)),
					},
				},
			}

			assert.NoError(t, handleSourcePublishPretend(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourcePublishPretend_NoStateNoSessionError(t *testing.T) {
	var testState = &task.State{
		Internal: broker.DataLink{},
	}
	var testParams = &SourcePublishParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleSourcePublishPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourcePublishPretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Internal: broker.DataLinkInstance{},
		Error:    errorDataSourceExist,
	}
	var testParams = &SourcePublishParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleSourcePublishPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourcePublishPretend_StateError(t *testing.T) {
	var expectedErrors = errors.New("expected")
	var testState = &task.State{
		Error: expectedErrors,
	}

	assert.ErrorIs(t, handleSourcePublishPretend(&SourcePublishParams{}, testState), expectedErrors)
}

func Test_handleSourceSyncComplete(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var expectedOem = "oem"
		var expectedHandle = "handle"
		var expectedVersion = "version"
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testState = &task.State{
			Logger: logrus.New(),
			Output: "output",
		}
		var testDataLink = &broker.DataLink{
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
		}
		var testParams = &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				DataLink: testDataLink,
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},

			client: &stubSourceClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataSourceResult: []broker.DataLinkInstance{
					*broker.NewDataLinkInstance(1, "name1", testDataLink),
					*broker.NewDataLinkInstance(2, "name2", &broker.DataLink{
						Oem:     "other",
						Handle:  "other",
						Version: "version",
					}),
				},
			},
		}

		assert.NoError(t, handleSourceSyncComplete(testParams, testState))
		assert.Equal(t, int64(1), *testState.Internal.(broker.DataLinkInstance).Id)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceSyncComplete_Conflict(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var expectedOem = "oem"
		var expectedHandle = "handle"
		var expectedVersion = "version"
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testState = &task.State{
			Logger: logrus.New(),
			Output: "output",
		}
		var testDataLink = &broker.DataLink{
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
		}
		var testParams = &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				DataLink: testDataLink,
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},

			client: &stubSourceClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataSourceResult: []broker.DataLinkInstance{
					*broker.NewDataLinkInstance(1, "name1", testDataLink),
					*broker.NewDataLinkInstance(2, "name2", testDataLink),
				},
			},
		}

		assert.ErrorIs(t, handleSourceSyncComplete(testParams, testState), errorLockerDataLinkConflict)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceSyncComplete_ConflictSourceName(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var expectedOem = "oem"
		var expectedHandle = "handle"
		var expectedVersion = "version"
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testState = &task.State{
			Logger: logrus.New(),
			Output: "output",
		}
		var testDataLink = &broker.DataLink{
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
		}
		var testParams = &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				DataLink: testDataLink,
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},
			SourceName: "name1",

			client: &stubSourceClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataSourceResult: []broker.DataLinkInstance{
					*broker.NewDataLinkInstance(1, "name1", testDataLink),
					*broker.NewDataLinkInstance(2, "name1", testDataLink),
				},
			},
		}

		assert.ErrorIs(t, handleSourceSyncComplete(testParams, testState), errorLockerDataSourceConflict)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceSyncComplete_ListSourcesError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testState = &task.State{
			Logger: logrus.New(),
			Output: "output",
		}
		var testParams = &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},

			client: &stubSourceClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataSourceError: expectedError,
			},
		}

		assert.ErrorIs(t, handleSourceSyncComplete(testParams, testState), expectedError)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceSyncComplete_NoSources(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testState = &task.State{
			Logger: logrus.New(),
			Output: "output",
		}
		var testParams = &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},

			client: &stubSourceClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataSourceResult: []broker.DataLinkInstance{},
			},
		}

		assert.NoError(t, handleSourceSyncComplete(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceSyncComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleSourceSyncComplete(&SourceFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleSourceSyncComplete_SessionError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &SourceFindParams{
		DataLinkParams: &broker.DataLinkParams{
			Broker: broker.Broker{
				AuthFile: filepath.Join(t.TempDir(), ".auth"),
				HostAddr: "hostAddr",
			},
		},
	}

	assert.ErrorIs(t, handleSourceSyncComplete(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourceSyncComplete_WithSourceName(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var expectedOem = "oem"
		var expectedHandle = "handle"
		var expectedVersion = "version"
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testState = &task.State{
			Logger: logrus.New(),
			Output: "output",
		}
		var testDataLink = &broker.DataLink{
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
		}
		var testParams = &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				DataLink: testDataLink,
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},
			SourceName: "name1",

			client: &stubSourceClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataSourceResult: []broker.DataLinkInstance{
					*broker.NewDataLinkInstance(1, "name1", testDataLink),
					*broker.NewDataLinkInstance(2, "name2", testDataLink),
				},
			},
		}

		assert.NoError(t, handleSourceSyncComplete(testParams, testState))
		assert.Equal(t, int64(1), *testState.Internal.(broker.DataLinkInstance).Id)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceSyncContext(t *testing.T) {
	var testParams = &SourceFindParams{
		DataLinkParams: &broker.DataLinkParams{
			DataLink: &broker.DataLink{
				Oem:     "oem",
				Handle:  "handle",
				Version: "version",
			},
		},
	}

	assert.NoError(t, handleSourceSyncContext(testParams, &task.State{}))
}

func Test_handleSourceSyncContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleSourceSyncContext(&SourceFindParams{}, testState))
}

func Test_handleSourceSyncContext_DataLinkInvalid(t *testing.T) {
	assert.ErrorIs(t, handleSourceSyncContext(&SourceFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleSourceSyncPretend(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testState = &task.State{
			Logger: logrus.New(),
			Output: "output",
		}
		var testParams = &SourceFindParams{
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},
			SourceHandle: "lockerHandle",

			client: &stubSourceClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				createDataSourceResult: &broker.DataLinkInstance{
					Id: new(int64(37)),
				},
			},
		}

		assert.NoError(t, handleSourceSyncPretend(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceSyncPretend_DataLinkInvalid(t *testing.T) {
	assert.ErrorIs(t, handleSourceSyncPretend(&SourceFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleSourceSyncPretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &SourceFindParams{
		DataLinkParams: &broker.DataLinkParams{
			Broker: broker.Broker{
				AuthFile: filepath.Join(t.TempDir(), ".auth"),
				HostAddr: "hostAddr",
			},
		},
	}

	assert.ErrorIs(t, handleSourceSyncPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSourceUpdateComplete(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testHandle",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceUpdateParams{
				SourceFindParams: &SourceFindParams{
					BaseParams: BaseParams{
						LockerPath: testLockerPath,
						Passphrase: testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
					SourceHandle: testLink.LockerHandle,
				},
				PropertyParams: PropertyParams{
					Key: "myKey",
					// Should produce a warning
				},
			}

			assert.NoError(t, handleSourceUpdateComplete(testParams, testState))
			assert.NotEmpty(t, testState.Reports)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdateComplete_ChaChaError(t *testing.T) {
	// The only way to cause a ChaCha error is to fail add by providing the wrong password
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase = memguard.NewEnclave([]byte("notTheRightPassword"))

		if _, err = writeTestLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceUpdateParams{
				SourceFindParams: &SourceFindParams{
					BaseParams: BaseParams{
						LockerPath: testLockerPath,
						Passphrase: testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
					SourceHandle: "testHandle",
				},
			}

			assert.ErrorIs(t, handleSourceUpdateComplete(testParams, testState), errorLockerPassFailed)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdateComplete_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Output: "known",
			Logger: logrus.New(),
		}
		var testParams = &SourceUpdateParams{
			SourceFindParams: &SourceFindParams{
				BaseParams: BaseParams{
					LockerPath: filepath.Join(testDir, "locker.bin"),
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
			},
		}

		assert.Error(t, handleSourceUpdateComplete(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdateComplete_LockerUpdateError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceUpdateParams{
				SourceFindParams: &SourceFindParams{
					BaseParams: BaseParams{
						LockerPath: testLockerPath,
						Passphrase: testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
					SourceHandle: "testHandle",
				},
				PropertyParams: PropertyParams{
					Key:    "myKey",
					Secret: memguard.NewEnclave([]byte("myValue")),
				},
			}

			assert.ErrorIs(t, handleSourceUpdateComplete(testParams, testState), errorLockerAccountNotFound)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdateComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleSourceUpdateComplete(&SourceUpdateParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleSourceUpdateComplete_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &SourceUpdateParams{
		SourceFindParams: &SourceFindParams{
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: filepath.Join(t.TempDir(), ".auth"),
					HostAddr: "hostAddr",
				},
			},
		},
	}

	assert.ErrorIs(t, handleSourceUpdateComplete(testParams, testState), broker.ErrorNoSession)
}

/*
This is a very nasty internal issue that should not arise: If the password to decrypt the locker file fails on
one of the source, the error returned will be a generic decryption failure.

Normally this should not happen, unless someone implemented functionality invoked in separate locker calls without
synchronizing the source with a new passphrase.
*/
func Test_handleSourceUpdateComplete_SourceChaChaError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testHandle",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testProperties = map[string]string{
			"aKey": "aValue",
		}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestChaChaError(testLockerPath, testUrl, testLink, testProperties); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceUpdateParams{
				SourceFindParams: &SourceFindParams{
					BaseParams: BaseParams{
						LockerPath: testLockerPath,
						Passphrase: testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
					SourceHandle: testLink.LockerHandle,
				},
				PropertyParams: PropertyParams{
					Key:    "myKey",
					Secret: memguard.NewEnclave([]byte("someSecret")),
				},
			}

			assert.ErrorIs(t, handleSourceUpdateComplete(testParams, testState), errorLockerPassFailed)
			assert.Empty(t, testState.Reports)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdateComplete_SourceNotFoundError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testSource",
			LinkOem:      "sourceOem",
			LinkHandle:   "sourceHandle",
			LinkVersion:  "sourceVersion",
		}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &SourceUpdateParams{
				SourceFindParams: &SourceFindParams{
					BaseParams: BaseParams{
						LockerPath: testLockerPath,
						Passphrase: testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
					SourceHandle: "notTheRightSource",
				},
				PropertyParams: PropertyParams{
					Key:   "myKey",
					Value: "myValue",
				},
			}

			assert.Error(t, handleSourceUpdateComplete(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdateContext(t *testing.T) {
	var testLocker = filepath.Join(t.TempDir(), "locker.bin")
	var expectedKey = "myKey"
	var testParams = &SourceUpdateParams{
		SourceFindParams: &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: testLocker,
				Passphrase: memguard.NewEnclave([]byte("test")),
			},
			DataLinkParams: &broker.DataLinkParams{
				DataLink: &broker.DataLink{
					Oem:     "oem",
					Handle:  "handle",
					Version: "version",
				},
			},
		},
		PropertyParams: PropertyParams{
			Key:   expectedKey,
			Value: "myValue",
		},
	}
	var testSpec = &broker.PropSpec{
		Key: expectedKey,
	}
	var testState = &task.State{
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				testSpec.VarSpec(),
			},
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testLocker); err == nil {
		defer filez.CloseSilently(fd)

		assert.NoError(t, handleSourceUpdateContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdateContext_BaseError(t *testing.T) {
	var testParams = &SourceUpdateParams{
		SourceFindParams: &SourceFindParams{
			BaseParams: BaseParams{},
		},
	}

	assert.ErrorIs(t, handleSourceUpdateContext(testParams, &task.State{}), errorLockerDataLinkEmpty)
}

func Test_handleSourceUpdateContext_InvalidValueError(t *testing.T) {
	var testLocker = filepath.Join(t.TempDir(), "locker.bin")
	var expectedKey = "myKey"
	var testParams = &SourceUpdateParams{
		SourceFindParams: &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: testLocker,
				Passphrase: memguard.NewEnclave([]byte("test")),
			},
			DataLinkParams: &broker.DataLinkParams{
				DataLink: &broker.DataLink{
					Oem:     "oem",
					Handle:  "handle",
					Version: "version",
				},
			},
		},
		PropertyParams: PropertyParams{
			Key:   expectedKey,
			Value: "myValue",
		},
	}
	var testSpec = &broker.PropSpec{
		Key:  expectedKey,
		Type: broker.PropSpecTypeInt,
	}
	var testState = &task.State{
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				testSpec.VarSpec(),
			},
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testLocker); err == nil {
		defer filez.CloseSilently(fd)

		assert.Error(t, handleSourceUpdateContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdateContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleSourceUpdateContext(&SourceUpdateParams{}, testState))
}

func Test_handleSourceUpdateContext_SecretKeyError(t *testing.T) {
	var testLocker = filepath.Join(t.TempDir(), "locker.bin")
	var expectedKey = "myKey"
	var testParams = &SourceUpdateParams{
		SourceFindParams: &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: testLocker,
				Passphrase: memguard.NewEnclave([]byte("test")),
			},
			DataLinkParams: &broker.DataLinkParams{
				DataLink: &broker.DataLink{
					Oem:     "oem",
					Handle:  "handle",
					Version: "version",
				},
			},
		},
		PropertyParams: PropertyParams{
			Key:   expectedKey,
			Value: "shouldBeSecret",
		},
	}
	var testSpec = &broker.PropSpec{
		Key: expectedKey,
	}
	var testState = &task.State{
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				testSpec.SecretSpec(),
			},
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testLocker); err == nil {
		defer filez.CloseSilently(fd)

		assert.Error(t, handleSourceUpdateContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdateContext_UnknownKeyError(t *testing.T) {
	var testLocker = filepath.Join(t.TempDir(), "locker.bin")
	var testParams = &SourceUpdateParams{
		SourceFindParams: &SourceFindParams{
			BaseParams: BaseParams{
				LockerPath: testLocker,
				Passphrase: memguard.NewEnclave([]byte("test")),
			},
			DataLinkParams: &broker.DataLinkParams{
				DataLink: &broker.DataLink{
					Oem:     "oem",
					Handle:  "handle",
					Version: "version",
				},
			},
		},
		PropertyParams: PropertyParams{
			Key:   "notTheKey",
			Value: "someValueDoesNotMatter",
		},
	}
	var testState = &task.State{
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				broker.PropSpec{
					Key: "myKey",
				},
			},
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testLocker); err == nil {
		defer filez.CloseSilently(fd)

		assert.Error(t, handleSourceUpdateContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdatePretend(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLogger, testHook = test.NewNullLogger()
		var testState = &task.State{
			Output: "known",
			Logger: testLogger,
		}
		var testParams = &SourceUpdateParams{
			SourceFindParams: &SourceFindParams{
				BaseParams: BaseParams{
					LockerPath: filepath.Join(testDir, "locker.bin"),
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
					DataLink: &broker.DataLink{
						Oem:     "myOem",
						Handle:  "myHandle",
						Version: "myVersion",
					},
				},
			},
		}

		testLogger.SetLevel(logrus.DebugLevel)
		assert.NoError(t, handleSourceUpdatePretend(testParams, testState))
		assert.Equal(t, 2, len(testHook.Entries))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleSourceUpdatePretend_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleSourceUpdatePretend(&SourceUpdateParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleSourceUpdatePretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &SourceUpdateParams{
		SourceFindParams: &SourceFindParams{
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: filepath.Join(t.TempDir(), ".auth"),
					HostAddr: "hostAddr",
				},
			},
		},
	}

	assert.ErrorIs(t, handleSourceUpdatePretend(testParams, testState), broker.ErrorNoSession)
}

func writeTestAuthSession(authPath, testUrl string) error {
	var authData = broker.NewAuthData()
	var authSession = &broker.AuthSession{
		Token:  "test",
		Expiry: -1,
	}

	return authData.Push(testUrl, authSession).Write(authPath)
}

func writeTestChaChaError(lockerPath, testUrl string, link *lockerLink, properties map[string]string) (Enclave, error) {
	var lockerState = NewSecuredLockerState(&task.State{})
	var testPassphrase = memguard.NewEnclave([]byte("test"))
	var unsyncPassphrase = memguard.NewEnclave([]byte("result"))
	var err error

	if err = lockerState.addSource(testUrl, link); err == nil {
		for k, v := range properties {
			var valueEnclave = memguard.NewEnclave([]byte(v))

			if err = lockerState.updateSource(testUrl, link.LockerHandle, k, valueEnclave, testPassphrase); err != nil {
				break
			}
		}
	}

	if err == nil {
		if err = lockerState.Write(lockerPath, unsyncPassphrase); err == nil {
			return unsyncPassphrase, nil
		}
	}

	return nil, err
}
