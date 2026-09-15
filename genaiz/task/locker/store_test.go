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

type stubStoreClient struct {
	stubClient
	createDataStoreError    error
	createDataStoreInstance *broker.DataLinkInstance
	createDataStoreProps    map[string]string
	createDataStoreResult   *broker.DataLinkInstance
	listDataStoreError      error
	listDataStoreResult     []broker.DataLinkInstance
	updateDataStoreError    error
	updateDataStoreInstance *broker.DataLinkInstance
	updateDataStoreProps    map[string]string
	updateDataStoreResult   *broker.DataLinkInstance
}

func (ssc *stubStoreClient) CreateDataStore(instance *broker.DataLinkInstance, props map[string]string) (*broker.DataLinkInstance, error) {
	ssc.createDataStoreInstance = instance
	ssc.createDataStoreProps = props
	return ssc.createDataStoreResult, ssc.createDataStoreError
}

func (ssc *stubStoreClient) ListDataStores() ([]broker.DataLinkInstance, error) {
	return ssc.listDataStoreResult, ssc.listDataStoreError
}

func (ssc *stubStoreClient) UpdateDataStore(instance *broker.DataLinkInstance, props map[string]string) (*broker.DataLinkInstance, error) {
	ssc.updateDataStoreInstance = instance
	ssc.updateDataStoreProps = props
	return ssc.updateDataStoreResult, ssc.updateDataStoreError
}

func TestNewStoreAddTask(t *testing.T) {
	var testTask = NewStoreAddTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.Nil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewStoreFindTask(t *testing.T) {
	var testTask = NewStoreFindTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewStorePublishTask(t *testing.T) {
	var testTask = NewStorePublishTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewStoreSyncTask(t *testing.T) {
	var testTask = NewStoreSyncTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewStoreUpdateTask(t *testing.T) {
	var testTask = NewStoreUpdateTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func Test_handleStoreAddComplete(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &StoreAddParams{
				BaseParams: BaseParams{
					LockerHandle: "testHandle",
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			}

			assert.NoError(t, handleStoreAddComplete(testParams, testState))
			assert.NotEmpty(t, testState.Reports)
			assert.Contains(t, testState.Reports[0], testParams.Handle)
			assert.Contains(t, testState.Reports[0], testParams.LockerPath)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreAddComplete_ChaChaError(t *testing.T) {
	// The only way to cause a ChaCha error is to fail add by providing the wrong password
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase = memguard.NewEnclave([]byte("notTheRightPassword"))

		if _, err = writeTestStoreLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &StoreAddParams{
				BaseParams: BaseParams{
					LockerHandle: "testHandle",
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			}

			assert.ErrorIs(t, handleStoreAddComplete(testParams, testState), errorLockerPassFailed)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreAddComplete_LockerAddError(t *testing.T) {
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

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &StoreAddParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			}

			assert.Error(t, handleStoreAddComplete(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreAddComplete_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Output: "known",
			Logger: logrus.New(),
		}
		var testParams = &StoreAddParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			Broker: broker.Broker{
				AuthFile: testAuthFile,
				HostAddr: testUrl,
			},
		}

		assert.Error(t, handleStoreAddComplete(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreAddComplete_LockerWriteError(t *testing.T) {
	// The write will fail after a successful add if the file's permissions are no longer writeable
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &StoreAddParams{
				BaseParams: BaseParams{
					LockerHandle: "testHandle",
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			}

			if err = os.Chmod(testLockerPath, 0400); err == nil {
				assert.Error(t, handleStoreAddComplete(testParams, testState))
				return
			}
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreAddComplete_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &StoreAddParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleStoreAddComplete(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStoreAddComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleStoreAddComplete(&StoreAddParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleStoreAddComplete_UnfoldError(t *testing.T) {
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
			var testParams = &StoreAddParams{
				BaseParams: BaseParams{
					LockerHandle: "someStore",
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			}

			assert.Error(t, handleStoreAddComplete(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreAddContext(t *testing.T) {
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
	var testParams = &StoreAddParams{
		BaseParams: BaseParams{
			LockerPath: testPath,
			Passphrase: memguard.NewEnclave([]byte("test")),
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		defer filez.CloseSilently(fd)

		assert.NoError(t, handleStoreAddContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreAddContext_CheckOutput(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleStoreAddContext(&StoreAddParams{}, testState))
}

func Test_handleStoreAddContext_NoPassphrase(t *testing.T) {
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
	var testParams = &StoreAddParams{
		BaseParams: BaseParams{
			LockerPath: testPath,
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testPath); err == nil {
		defer filez.CloseSilently(fd)

		assert.ErrorIs(t, handleStoreAddContext(testParams, testState), errorLockerPassFailed)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreAddContext_NoVarSpecs(t *testing.T) {
	assert.ErrorIs(t, handleStoreAddContext(&StoreAddParams{}, &task.State{}), errorLockerDataLinkEmpty)
}

func Test_handleStoreAddContext_UnreadableLocker(t *testing.T) {
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
	var testParams = &StoreAddParams{
		BaseParams: BaseParams{
			LockerPath: testPath,
		},
	}

	assert.Error(t, handleStoreAddContext(testParams, testState))
}

func Test_handleStoreAddPretend(t *testing.T) {
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
		var testParams = &StoreAddParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			Broker: broker.Broker{
				AuthFile: testAuthFile,
				HostAddr: testUrl,
			},
		}

		testLogger.SetLevel(logrus.DebugLevel)
		assert.NoError(t, handleStoreAddPretend(testParams, testState))
		assert.Equal(t, 2, len(testHook.Entries))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreAddPretend_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleStoreAddPretend(&StoreAddParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleStoreAddPretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &StoreAddParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleStoreAddPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStoreFindComplete(t *testing.T) {
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

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: testLink.LockerHandle,
				Logger: logrus.New(),
			}
			var testParams = &StoreFindParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
			}

			assert.NoError(t, handleStoreFindComplete(testParams, testState))
			assert.Equal(t, testLink.LinkOem, testParams.Oem)
			assert.Equal(t, testLink.LinkHandle, testParams.Handle)
			assert.Equal(t, testLink.LinkVersion, testParams.Version)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreFindComplete_ChaChaError(t *testing.T) {
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

		if _, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "invalidDataStore",
				Logger: logrus.New(),
			}
			var testParams = &StoreFindParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
			}

			assert.ErrorIs(t, handleStoreFindComplete(testParams, testState), errorLockerPassFailed)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreFindComplete_LockerLookupError(t *testing.T) {
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

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "invalidDataStore",
				Logger: logrus.New(),
			}
			var testParams = &StoreFindParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
			}

			assert.Error(t, handleStoreFindComplete(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreFindComplete_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Output: "known",
			Logger: logrus.New(),
		}
		var testParams = &StoreFindParams{
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

		assert.Error(t, handleStoreFindComplete(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreFindComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleStoreFindComplete(&StoreFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleStoreFindComplete_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &StoreFindParams{
		DataLinkParams: &broker.DataLinkParams{
			Broker: broker.Broker{
				AuthFile: filepath.Join(t.TempDir(), ".auth"),
				HostAddr: "hostAddr",
			},
		},
	}

	assert.ErrorIs(t, handleStoreFindComplete(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStoreFindContext(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &StoreFindParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("test")),
		},
	}

	assert.NoError(t, handleStoreFindContext(testParams, testState))
}

func Test_handleStoreFindContext_FoundDataLink(t *testing.T) {
	var testParams = &StoreFindParams{
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

	assert.ErrorIs(t, handleStoreFindContext(testParams, testState), errorLockerDataLinkFound)
	assert.Contains(t, testState.Output, testParams.Oem)
	assert.Contains(t, testState.Output, testParams.Handle)
	assert.Contains(t, testState.Output, testParams.Version)
}

func Test_handleStoreFindContext_OutputKnown(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleStoreFindContext(&StoreFindParams{}, testState))
}

func Test_handleStoreFindContext_NoPassphrase(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleStoreFindContext(&StoreFindParams{}, testState), errorLockerPassFailed)
}

func Test_handleStoreFindIncomplete(t *testing.T) {
	var testParams = &StoreFindParams{
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

	assert.NoError(t, handleStoreFindIncomplete(testParams, testState))
	assert.Equal(t, 1, len(testState.Reports))
}

func Test_handleStoreFindIncomplete_TaskError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleStoreFindIncomplete(&StoreFindParams{}, testState), expectedError)
}

func Test_handleStoreFindPretend(t *testing.T) {
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

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testLogger, testHook = test.NewNullLogger()
			var testState = &task.State{
				Output: "known",
				Logger: testLogger,
			}
			var testParams = &StoreFindParams{
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
			assert.NoError(t, handleStoreFindPretend(testParams, testState))
			assert.Equal(t, 2, len(testHook.Entries))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreFindPretend_ChaChaError(t *testing.T) {
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

		if _, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "invalidDataStore",
				Logger: logrus.New(),
			}
			var testParams = &StoreFindParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				DataLinkParams: &broker.DataLinkParams{
					Broker: broker.Broker{
						AuthFile: testAuthFile,
						HostAddr: testUrl,
					},
				},
			}

			assert.ErrorIs(t, handleStoreFindPretend(testParams, testState), errorLockerPassFailed)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreFindPretend_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleStoreFindPretend(&StoreFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleStoreFindPretend_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Output: "known",
			Logger: logrus.New(),
		}
		var testParams = &StoreFindParams{
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

		assert.Error(t, handleStoreFindPretend(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreFindPretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &StoreFindParams{
		DataLinkParams: &broker.DataLinkParams{
			Broker: broker.Broker{
				AuthFile: filepath.Join(t.TempDir(), ".auth"),
				HostAddr: "hostAddr",
			},
		},
	}

	assert.ErrorIs(t, handleStoreFindPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStorePublishContext(t *testing.T) {
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: broker.DataLinkInstance{},
	}
	var testParams = &StorePublishParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("passphrase")),
		},
	}

	assert.ErrorIs(t, handleStorePublishContext(testParams, testState), errorDataStoreExist)
}

func Test_handleStorePublishContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleStorePublishContext(&StorePublishParams{}, testState))
}

func Test_handleStorePublishContext_NoPassphrase(t *testing.T) {
	assert.ErrorIs(t, handleStorePublishContext(&StorePublishParams{}, &task.State{}), errorLockerPassFailed)
}

func Test_handleStorePublishContext_NoStateInternal(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &StorePublishParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("passphrase")),
		},
	}

	assert.ErrorIs(t, handleStorePublishContext(testParams, testState), errorDataStoreLinkUnknown)
}

func Test_handleStorePublishContext_NoStateLinkOrStore(t *testing.T) {
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: "notALinkOrStore",
	}
	var testParams = &StorePublishParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("passphrase")),
		},
	}

	assert.ErrorIs(t, handleStorePublishContext(testParams, testState), errorDataStoreLinkUnknown)
}

func Test_handleStorePublishContext_StateDataLink(t *testing.T) {
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: broker.DataLink{},
	}
	var testParams = &StorePublishParams{
		BaseParams: BaseParams{
			Passphrase: memguard.NewEnclave([]byte("passphrase")),
		},
	}

	assert.NoError(t, handleStorePublishContext(testParams, testState))
}

func Test_handleStorePublishCreate(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testStore",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Internal: broker.DataLink{
					Id: new(int64(37)),
				},
			}
			var testParams = &StorePublishParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},

				client: &stubStoreClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					createDataStoreResult: &broker.DataLinkInstance{
						Id: new(int64(37)),
					},
				},
			}

			assert.NoError(t, handleStorePublishCreate(testParams, testState))
			assert.NotEmpty(t, testState.Reports)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishCreate_CreateStoreError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testStore",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Internal: broker.DataLink{
					Id: new(int64(37)),
				},
			}
			var testParams = &StorePublishParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},

				client: &stubStoreClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					createDataStoreError: expectedError,
				},
			}

			assert.ErrorIs(t, handleStorePublishCreate(testParams, testState), expectedError)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishCreate_GetPropError(t *testing.T) {
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
			var testParams = &StorePublishParams{
				BaseParams: BaseParams{
					LockerHandle: "someStore",
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			}

			assert.Error(t, handleStorePublishCreate(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishCreate_LockerReadError(t *testing.T) {
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
		var testParams = &StorePublishParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			Broker: broker.Broker{
				AuthFile: testAuthFile,
				HostAddr: testUrl,
			},
		}

		assert.Error(t, handleStorePublishCreate(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishCreate_NoStateInternal(t *testing.T) {
	assert.ErrorIs(t, handleStorePublishCreate(&StorePublishParams{}, &task.State{}), errorDataStoreUnknown)
}

func Test_handleStorePublishCreate_SessionError(t *testing.T) {
	var testState = &task.State{
		Internal: broker.DataLink{},
	}
	var testParams = &StorePublishParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleStorePublishCreate(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStorePublishUpdate(t *testing.T) {
	var expectedError = errors.New("expected")
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testStore",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Error:  errorDataStoreExist,
				Internal: broker.DataLinkInstance{
					Id: new(int64(37)),
				},
			}
			var testParams = &StorePublishParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},

				client: &stubStoreClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					updateDataStoreResult: &broker.DataLinkInstance{
						Id: new(int64(37)),
					},
				},
			}

			assert.NoError(t, handleStorePublishUpdate(testParams, testState), expectedError)
			assert.NotEmpty(t, testState.Reports)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishUpdate_GetPropsError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testStore",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Error:  errorDataStoreExist,
				Internal: broker.DataLinkInstance{
					Id: new(int64(37)),
				},
			}
			var testParams = &StorePublishParams{
				BaseParams: BaseParams{
					LockerHandle: "someStore",
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			}

			assert.Error(t, handleStorePublishUpdate(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishUpdate_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Logger: logrus.New(),
			Error:  errorDataStoreExist,
			Internal: broker.DataLinkInstance{
				Id: new(int64(37)),
			},
		}
		var testParams = &StorePublishParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			Broker: broker.Broker{
				AuthFile: testAuthFile,
				HostAddr: testUrl,
			},
		}

		assert.Error(t, handleStorePublishUpdate(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishUpdate_SessionError(t *testing.T) {
	var testState = &task.State{
		Error:    errorDataStoreExist,
		Internal: broker.DataLinkInstance{},
	}
	var testParams = &StorePublishParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleStorePublishUpdate(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStorePublishUpdate_StoreUnknown(t *testing.T) {
	var testState = &task.State{
		Error: errorDataStoreExist,
	}

	assert.ErrorIs(t, handleStorePublishUpdate(&StorePublishParams{}, testState), errorDataStoreUnknown)
}

func Test_handleStorePublishUpdate_StateError(t *testing.T) {
	var testError = errors.New("expected")
	var testState = &task.State{
		Error: testError,
	}

	assert.ErrorIs(t, handleStorePublishUpdate(&StorePublishParams{}, testState), testError)
}

func Test_handleStorePublishUpdate_UpdateStoreError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testStore",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Error:  errorDataStoreExist,
				Internal: broker.DataLinkInstance{
					Id: new(int64(37)),
				},
			}
			var testParams = &StorePublishParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},

				client: &stubStoreClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					updateDataStoreError: expectedError,
				},
			}

			assert.ErrorIs(t, handleStorePublishUpdate(testParams, testState), expectedError)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishPretend(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testStore",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Error:  errorDataStoreExist,
				Internal: broker.DataLinkInstance{
					Id:         new(int64(37)),
					DataLinkId: new(int64(42)),
				},
			}
			var testParams = &StorePublishParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},

				client: &stubStoreClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					createDataStoreResult: &broker.DataLinkInstance{
						Id: new(int64(37)),
					},
				},
			}

			assert.NoError(t, handleStorePublishPretend(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishPretend_NoStateError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testClient, _ = broker.GetClient(testAuthFile, testUrl)
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testStore",
			LinkOem:      "testOem",
			LinkHandle:   "testHandle",
			LinkVersion:  "testVersion",
		}
		var testMappings = map[string]string{"key": "value"}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, testMappings); err == nil {
			var testState = &task.State{
				Logger: logrus.New(),
				Internal: broker.DataLink{
					Id: new(int64(37)),
				},
			}
			var testParams = &StorePublishParams{
				BaseParams: BaseParams{
					LockerHandle: testLink.LockerHandle,
					LockerPath:   testLockerPath,
					Passphrase:   testPassPhrase,
				},
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},

				client: &stubStoreClient{
					stubClient: stubClient{
						decoratedClient: testClient,
					},
					createDataStoreResult: &broker.DataLinkInstance{
						Id: new(int64(37)),
					},
				},
			}

			assert.NoError(t, handleStorePublishPretend(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStorePublishPretend_NoStateNoSessionError(t *testing.T) {
	var testState = &task.State{
		Internal: broker.DataLink{},
	}
	var testParams = &StorePublishParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleStorePublishPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleSorePublishPretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Internal: broker.DataLinkInstance{},
		Error:    errorDataStoreExist,
	}
	var testParams = &StorePublishParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleStorePublishPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStorePublishPretend_StateError(t *testing.T) {
	var expectedErrors = errors.New("expected")
	var testState = &task.State{
		Error: expectedErrors,
	}

	assert.ErrorIs(t, handleStorePublishPretend(&StorePublishParams{}, testState), expectedErrors)
}

func Test_handleStoreSyncComplete(t *testing.T) {
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
		var testParams = &StoreFindParams{
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

			client: &stubStoreClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataStoreResult: []broker.DataLinkInstance{
					*broker.NewDataLinkInstance(1, "name1", testDataLink),
					*broker.NewDataLinkInstance(2, "name2", &broker.DataLink{
						Oem:     "other",
						Handle:  "other",
						Version: "version",
					}),
				},
			},
		}

		assert.NoError(t, handleStoreSyncComplete(testParams, testState))
		assert.Equal(t, int64(1), *testState.Internal.(broker.DataLinkInstance).Id)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreSyncComplete_Conflict(t *testing.T) {
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
		var testParams = &StoreFindParams{
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

			client: &stubStoreClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataStoreResult: []broker.DataLinkInstance{
					*broker.NewDataLinkInstance(1, "name1", testDataLink),
					*broker.NewDataLinkInstance(2, "name2", testDataLink),
				},
			},
		}

		assert.ErrorIs(t, handleStoreSyncComplete(testParams, testState), errorLockerDataStoreConflict)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreSyncComplete_ConflictStoreName(t *testing.T) {
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
		var testParams = &StoreFindParams{
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
			StoreName: "name1",

			client: &stubStoreClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataStoreResult: []broker.DataLinkInstance{
					*broker.NewDataLinkInstance(1, "name1", testDataLink),
					*broker.NewDataLinkInstance(2, "name1", testDataLink),
				},
			},
		}

		assert.ErrorIs(t, handleStoreSyncComplete(testParams, testState), errorLockerDataStoreConflict)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreSyncComplete_ListStoresError(t *testing.T) {
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
		var testParams = &StoreFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},

			client: &stubStoreClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataStoreError: expectedError,
			},
		}

		assert.ErrorIs(t, handleStoreSyncComplete(testParams, testState), expectedError)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreSyncComplete_NoStores(t *testing.T) {
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
		var testParams = &StoreFindParams{
			BaseParams: BaseParams{
				LockerPath: filepath.Join(testDir, "locker.bin"),
			},
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},

			client: &stubStoreClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataStoreResult: []broker.DataLinkInstance{},
			},
		}

		assert.NoError(t, handleStoreSyncComplete(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreSyncComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleStoreSyncComplete(&StoreFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleStoreSyncComplete_SessionError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &StoreFindParams{
		DataLinkParams: &broker.DataLinkParams{
			Broker: broker.Broker{
				AuthFile: filepath.Join(t.TempDir(), ".auth"),
				HostAddr: "hostAddr",
			},
		},
	}

	assert.ErrorIs(t, handleStoreSyncComplete(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStoreSyncComplete_WithStoreName(t *testing.T) {
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
		var testParams = &StoreFindParams{
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
			StoreName: "name1",

			client: &stubStoreClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				listDataStoreResult: []broker.DataLinkInstance{
					*broker.NewDataLinkInstance(1, "name1", testDataLink),
					*broker.NewDataLinkInstance(2, "name2", testDataLink),
				},
			},
		}

		assert.NoError(t, handleStoreSyncComplete(testParams, testState))
		assert.Equal(t, int64(1), *testState.Internal.(broker.DataLinkInstance).Id)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreSyncContext(t *testing.T) {
	var testParams = &StoreFindParams{
		DataLinkParams: &broker.DataLinkParams{
			DataLink: &broker.DataLink{
				Oem:     "oem",
				Handle:  "handle",
				Version: "version",
			},
		},
	}

	assert.NoError(t, handleStoreSyncContext(testParams, &task.State{}))
}

func Test_handleStoreSyncContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleStoreSyncContext(&StoreFindParams{}, testState))
}

func Test_handleStoreSyncContext_DataLinkInvalid(t *testing.T) {
	assert.ErrorIs(t, handleStoreSyncContext(&StoreFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleStoreSyncPretend(t *testing.T) {
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
		var testParams = &StoreFindParams{
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: testAuthFile,
					HostAddr: testUrl,
				},
			},

			client: &stubStoreClient{
				stubClient: stubClient{
					decoratedClient: testClient,
				},
				createDataStoreResult: &broker.DataLinkInstance{
					Id: new(int64(37)),
				},
			},
		}

		assert.NoError(t, handleStoreSyncPretend(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreSyncPretend_DataLinkInvalid(t *testing.T) {
	assert.ErrorIs(t, handleStoreSyncPretend(&StoreFindParams{}, &task.State{}), errorLockerDataLinkInvalid)
}

func Test_handleStoreSyncPretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &StoreFindParams{
		DataLinkParams: &broker.DataLinkParams{
			Broker: broker.Broker{
				AuthFile: filepath.Join(t.TempDir(), ".auth"),
				HostAddr: "hostAddr",
			},
		},
	}

	assert.ErrorIs(t, handleStoreSyncPretend(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStoreUpdateComplete(t *testing.T) {
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

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &StoreUpdateParams{
				StoreFindParams: &StoreFindParams{
					BaseParams: BaseParams{
						LockerHandle: testLink.LockerHandle,
						LockerPath:   testLockerPath,
						Passphrase:   testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
				},
				PropertyParams: PropertyParams{
					Key: "myKey",
					// Should produce a warning
				},
			}

			assert.NoError(t, handleStoreUpdateComplete(testParams, testState))
			assert.NotEmpty(t, testState.Reports)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdateComplete_ChaChaError(t *testing.T) {
	// The only way to cause a ChaCha error is to fail add by providing the wrong password
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase = memguard.NewEnclave([]byte("notTheRightPassword"))

		if _, err = writeTestStoreLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &StoreUpdateParams{
				StoreFindParams: &StoreFindParams{
					BaseParams: BaseParams{
						LockerHandle: "testHandle",
						LockerPath:   testLockerPath,
						Passphrase:   testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
				},
			}

			assert.ErrorIs(t, handleStoreUpdateComplete(testParams, testState), errorLockerPassFailed)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdateComplete_LockerReadError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testState = &task.State{
			Output: "known",
			Logger: logrus.New(),
		}
		var testParams = &StoreUpdateParams{
			StoreFindParams: &StoreFindParams{
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

		assert.Error(t, handleStoreUpdateComplete(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdateComplete_LockerUpdateError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, nil, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &StoreUpdateParams{
				StoreFindParams: &StoreFindParams{
					BaseParams: BaseParams{
						LockerHandle: "testHandle",
						LockerPath:   testLockerPath,
						Passphrase:   testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
				},
				PropertyParams: PropertyParams{
					Key:    "myKey",
					Secret: memguard.NewEnclave([]byte("myValue")),
				},
			}

			assert.ErrorIs(t, handleStoreUpdateComplete(testParams, testState), errorLockerAccountNotFound)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdateComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleStoreUpdateComplete(&StoreUpdateParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleStoreUpdateComplete_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &StoreUpdateParams{
		StoreFindParams: &StoreFindParams{
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: filepath.Join(t.TempDir(), ".auth"),
					HostAddr: "hostAddr",
				},
			},
		},
	}

	assert.ErrorIs(t, handleStoreUpdateComplete(testParams, testState), broker.ErrorNoSession)
}

func Test_handleStoreUpdateComplete_StoreChaChaError(t *testing.T) {
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

		if testPassPhrase, err = writeTestStoreChaChaError(testLockerPath, testUrl, testLink, testProperties); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &StoreUpdateParams{
				StoreFindParams: &StoreFindParams{
					BaseParams: BaseParams{
						LockerHandle: testLink.LockerHandle,
						LockerPath:   testLockerPath,
						Passphrase:   testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
				},
				PropertyParams: PropertyParams{
					Key:    "myKey",
					Secret: memguard.NewEnclave([]byte("someSecret")),
				},
			}

			assert.ErrorIs(t, handleStoreUpdateComplete(testParams, testState), errorLockerPassFailed)
			assert.Empty(t, testState.Reports)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdateComplete_StoreNotFoundError(t *testing.T) {
	var testDir = t.TempDir()
	var testAuthFile = filepath.Join(testDir, ".auth")
	var testUrl = "testUrl"
	var err error

	if err = writeTestAuthSession(testAuthFile, testUrl); err == nil {
		var testLockerPath = filepath.Join(testDir, "locker.bin")
		var testLink = &lockerLink{
			LockerHandle: "testStore",
			LinkOem:      "storeOem",
			LinkHandle:   "storeHandle",
			LinkVersion:  "storeVersion",
		}
		var testPassPhrase Enclave

		if testPassPhrase, err = writeTestStoreLocker(testLockerPath, testUrl, testLink, nil); err == nil {
			var testState = &task.State{
				Output: "known",
				Logger: logrus.New(),
			}
			var testParams = &StoreUpdateParams{
				StoreFindParams: &StoreFindParams{
					BaseParams: BaseParams{
						LockerHandle: "notTheRightStore",
						LockerPath:   testLockerPath,
						Passphrase:   testPassPhrase,
					},
					DataLinkParams: &broker.DataLinkParams{
						Broker: broker.Broker{
							AuthFile: testAuthFile,
							HostAddr: testUrl,
						},
					},
				},
				PropertyParams: PropertyParams{
					Key:   "myKey",
					Value: "myValue",
				},
			}

			assert.Error(t, handleStoreUpdateComplete(testParams, testState))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdateContext(t *testing.T) {
	var testLocker = filepath.Join(t.TempDir(), "locker.bin")
	var expectedKey = "myKey"
	var testParams = &StoreUpdateParams{
		StoreFindParams: &StoreFindParams{
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

		assert.NoError(t, handleStoreUpdateContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdateContext_BaseError(t *testing.T) {
	var testParams = &StoreUpdateParams{
		StoreFindParams: &StoreFindParams{
			BaseParams: BaseParams{},
		},
	}

	assert.ErrorIs(t, handleStoreUpdateContext(testParams, &task.State{}), errorLockerDataLinkEmpty)
}

func Test_handleStoreUpdateContext_InvalidValueError(t *testing.T) {
	var testLocker = filepath.Join(t.TempDir(), "locker.bin")
	var expectedKey = "myKey"
	var testParams = &StoreUpdateParams{
		StoreFindParams: &StoreFindParams{
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

		assert.Error(t, handleStoreUpdateContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdateContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleStoreUpdateContext(&StoreUpdateParams{}, testState))
}

func Test_handleStoreUpdateContext_SecretKeyError(t *testing.T) {
	var testLocker = filepath.Join(t.TempDir(), "locker.bin")
	var expectedKey = "myKey"
	var testParams = &StoreUpdateParams{
		StoreFindParams: &StoreFindParams{
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

		assert.Error(t, handleStoreUpdateContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdateContext_UnknownKeyError(t *testing.T) {
	var testLocker = filepath.Join(t.TempDir(), "locker.bin")
	var testParams = &StoreUpdateParams{
		StoreFindParams: &StoreFindParams{
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

		assert.Error(t, handleStoreUpdateContext(testParams, testState))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdatePretend(t *testing.T) {
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
		var testParams = &StoreUpdateParams{
			StoreFindParams: &StoreFindParams{
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
		assert.NoError(t, handleStoreUpdatePretend(testParams, testState))
		assert.Equal(t, 2, len(testHook.Entries))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleStoreUpdatePretend_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleStoreUpdatePretend(&StoreUpdateParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleStoreUpdatePretend_SessionError(t *testing.T) {
	var testState = &task.State{
		Output: "known",
		Logger: logrus.New(),
	}
	var testParams = &StoreUpdateParams{
		StoreFindParams: &StoreFindParams{
			DataLinkParams: &broker.DataLinkParams{
				Broker: broker.Broker{
					AuthFile: filepath.Join(t.TempDir(), ".auth"),
					HostAddr: "hostAddr",
				},
			},
		},
	}

	assert.ErrorIs(t, handleStoreUpdatePretend(testParams, testState), broker.ErrorNoSession)
}

func writeTestStoreChaChaError(lockerPath, testUrl string, link *lockerLink, properties map[string]string) (Enclave, error) {
	var lockerState = NewSecuredLockerState(&task.State{})
	var testPassphrase = memguard.NewEnclave([]byte("test"))
	var unsyncPassphrase = memguard.NewEnclave([]byte("result"))
	var err error

	if err = lockerState.addStore(testUrl, link); err == nil {
		for k, v := range properties {
			var valueEnclave = memguard.NewEnclave([]byte(v))

			if err = lockerState.updateStore(testUrl, link.LockerHandle, k, valueEnclave, testPassphrase); err != nil {
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
