package locker

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
)

func TestInitParams_Validate(t *testing.T) {
	var testParams = &InitParams{
		Passphrase: memguard.NewEnclave([]byte("passworA0$")),
	}

	// Validating all special chars is the objective
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("passworB1!"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("passworC2\""))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("passworD3#"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("passworE4$"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("passworF5&"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("passworG6'"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("passworH7("))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("passworI8)"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("passworJ9*"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirK0+"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirL1,"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirM2-"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirN3."))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirO4/"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirP5:"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirQ6;"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirR7<"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirS8="))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("clazvirT9>"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbqumU0?"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbqumV1@"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbqumW2@"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbqumX3["))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbqumY4]"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbqumZ5^"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbquMA6\\"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbquNB7_"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbquNC8`"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbquND9{"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbquME0}"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbquMF1|"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbquMG2~"))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbquMH3 "))
	assert.NoError(t, testParams.Validate())
	testParams.Passphrase = memguard.NewEnclave([]byte("delbquMI4\t"))
	assert.NoError(t, testParams.Validate())
}

func TestInitParams_Validate_Empty(t *testing.T) {
	var testParams = &InitParams{}

	assert.ErrorIs(t, testParams.Validate(), errorLockerPassInvalid)
}

func TestInitParams_Validate_NoCapital(t *testing.T) {
	var testParams = &InitParams{
		Passphrase: memguard.NewEnclave([]byte("no_capital_letters_0%")),
	}

	assert.ErrorIs(t, testParams.Validate(), errorLockerPassCapLetter)
}

func TestInitParams_Validate_NoDigit(t *testing.T) {
	var testParams = &InitParams{
		Passphrase: memguard.NewEnclave([]byte("no_digits_D*")),
	}

	assert.ErrorIs(t, testParams.Validate(), errorLockerPassDigit)
}

func TestInitParams_Validate_NoSmall(t *testing.T) {
	var testParams = &InitParams{
		Passphrase: memguard.NewEnclave([]byte("NO_SMALL_LETTERS_0$*")),
	}

	assert.ErrorIs(t, testParams.Validate(), errorLockerPassSmallLetter)
}

func TestInitParams_Validate_NoSpecial(t *testing.T) {
	var testParams = &InitParams{
		Passphrase: memguard.NewEnclave([]byte("NoSpecials0")),
	}

	assert.ErrorIs(t, testParams.Validate(), errorLockerPassSpecial)
}

func TestInitParams_Validate_SameOld(t *testing.T) {
	var testParams = &InitParams{
		OldPassphrase: memguard.NewEnclave([]byte("samePass0#")),
		Passphrase:    memguard.NewEnclave([]byte("samePass0#")),
	}

	assert.ErrorIs(t, testParams.Validate(), errorLockerPassUnchanged)
}

func TestInitParams_Validate_Short(t *testing.T) {
	var testParams = &InitParams{
		Passphrase: memguard.NewEnclave([]byte("t0Sho&t")),
	}

	assert.ErrorIs(t, testParams.Validate(), errorLockerPassShort)
}

func TestNewInitLockerTask(t *testing.T) {
	var testTask = NewInitLockerTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func Test_handleLockerInitComplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: filepath.Join(t.TempDir(), "myLocker.bin.new"),
	}
	var testParams = &InitParams{
		LockerPath: filepath.Join(t.TempDir(), "myLocker.bin"),
		Passphrase: memguard.NewEnclave([]byte("password")),
	}
	var actual os.FileInfo
	var err error

	assert.NoError(t, handleLockerInitComplete(testParams, testState))

	if actual, err = os.Stat(testParams.LockerPath); err != nil {
		assert.Fail(t, "expected locker file %s", testParams.LockerPath)
	} else {
		assert.True(t, actual.Size() > 0)
	}
}

func Test_handleLockerInitComplete_NoStateOutput(t *testing.T) {
	assert.ErrorIs(t, handleLockerInitComplete(&InitParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleLockerInitComplete_PathError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: filepath.Join(t.TempDir(), "notExist", "myLocker.bin.new"),
	}
	var testParams = &InitParams{
		LockerPath: filepath.Join(t.TempDir(), "notExist", "myLocker.bin"),
		Passphrase: memguard.NewEnclave([]byte("password")),
	}

	assert.Error(t, handleLockerInitComplete(testParams, testState))
}

func Test_handleLockerInitComplete_RenameError(t *testing.T) {
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Logger: testLogger,
		Output: filepath.Join(t.TempDir(), "myLocker.bin.new"),
	}
	var testParams = &InitParams{
		LockerPath: filepath.Join(t.TempDir(), "notExist", "myLocker.bin"),
		Passphrase: memguard.NewEnclave([]byte("password")),
	}
	var actual os.FileInfo
	var err error

	testLogger.Level = logrus.DebugLevel
	assert.NoError(t, handleLockerInitComplete(testParams, testState))

	if actual, err = os.Stat(testState.Output); err != nil {
		assert.Fail(t, "expected stale locker file %s", testState.Output)
	} else {
		assert.Equal(t, 2, len(testHook.Entries))
		assert.Contains(t, testHook.Entries[1].Message, testParams.LockerPath)
		assert.True(t, actual.Size() > 0)
	}
}

func Test_handleLockerInitComplete_Update(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: filepath.Join(t.TempDir(), "myLocker.bin.new"),
	}
	var testParams = &InitParams{
		LockerPath: filepath.Join(t.TempDir(), "myLocker.bin"),
		Passphrase: memguard.NewEnclave([]byte("password")),
		Update:     true,
	}
	var actual os.FileInfo
	var err error

	assert.NoError(t, handleLockerInitComplete(testParams, testState))

	if actual, err = os.Stat(testParams.LockerPath); err != nil {
		assert.Fail(t, "expected locker file %s", testParams.LockerPath)
	} else {
		assert.Equal(t, 1, len(testState.Reports))
		assert.Contains(t, testState.Reports[0], "Updated ")
		assert.True(t, actual.Size() > 0)
	}
}

func Test_handleLockerInitContext(t *testing.T) {
	var expectedFile = filepath.Join(t.TempDir(), "myLocker.bin")
	var testParams = &InitParams{
		LockerPath:    expectedFile,
		OldPassphrase: memguard.NewEnclave([]byte("passworD1@")),
		Passphrase:    memguard.NewEnclave([]byte("passworD0#")),
		Update:        true,
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(expectedFile); err == nil {
		defer filez.CloseSilently(fd)

		assert.Error(t, handleLockerInitContext(testParams, &task.State{}), errorLockerUpdateNeeded)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleLockerInitContext_CheckOutput(t *testing.T) {
	var testState = &task.State{
		Output: "someOutput",
	}

	assert.NoError(t, handleLockerInitContext(&InitParams{}, testState))
}

func Test_handleLockerInitContext_NewLocker(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &InitParams{
		LockerPath: filepath.Join(t.TempDir(), "myLocker.bin"),
		Passphrase: memguard.NewEnclave([]byte("passworD0#")),
	}

	assert.NoError(t, handleLockerInitContext(testParams, testState))
	assert.Equal(t, testParams.LockerPath, testState.Output)
}

func Test_handleLockerInitContext_OldPasswordError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &InitParams{
		OldPassphrase: &stubEnclave{
			openError: expectedError,
		},
		Passphrase: memguard.NewEnclave([]byte("passworD0#")),
	}

	assert.Error(t, handleLockerInitContext(testParams, &task.State{}), expectedError)
}

func Test_handleLockerInitContext_OverwriteLocker(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &InitParams{
		LockerPath: filepath.Join(t.TempDir(), "myLocker.bin"),
		Passphrase: memguard.NewEnclave([]byte("passworD0#")),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.LockerPath); err == nil {
		defer filez.CloseSilently(fd)

		assert.NoError(t, handleLockerInitContext(testParams, testState))
		assert.NotEqual(t, testParams.LockerPath, testState.Output)
		assert.True(t, strings.HasSuffix(testState.Output, ".new"))
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleLockerInitContext_PasswordError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &InitParams{
		Passphrase: &stubEnclave{
			openError: expectedError,
		},
	}

	assert.Error(t, handleLockerInitContext(testParams, &task.State{}), expectedError)
}

func Test_handleLockerInitContext_PasswordInvalid(t *testing.T) {
	assert.Error(t, handleLockerInitContext(&InitParams{}, &task.State{}), errorLockerPassInvalid)
}

func Test_handleLockerInitContext_UpdateNoOldPassword(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &InitParams{
		LockerPath: filepath.Join(t.TempDir(), "myLocker.bin"),
		Passphrase: memguard.NewEnclave([]byte("passworD0#")),
		Update:     true,
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.LockerPath); err == nil {
		defer filez.CloseSilently(fd)

		assert.ErrorIs(t, handleLockerInitContext(testParams, testState), errorLockerReadNeeded)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_handleLockerInitPretend(t *testing.T) {
	var expectedFile = filepath.Join(t.TempDir(), "myLocker.bin")
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Logger: testLogger,
		Output: expectedFile,
	}
	var testParams = &InitParams{
		LockerPath: expectedFile,
	}

	testLogger.Level = logrus.DebugLevel
	assert.NoError(t, handleLockerInitPretend(testParams, testState))
	assert.Equal(t, 3, len(testHook.Entries))
	assert.Contains(t, testHook.Entries[1].Message, expectedFile)
	assert.Contains(t, testHook.Entries[1].Message, "using passphrase")
}

func Test_handleLockerInitPretend_PathInvalid(t *testing.T) {
	assert.ErrorIs(t, handleLockerInitPretend(&InitParams{}, &task.State{}), errorLockerPathInvalid)
}

func Test_handleLockerInitPretend_UpdatePath(t *testing.T) {
	var expectedFile = filepath.Join(t.TempDir(), "myLocker.bin")
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Logger: testLogger,
		Error:  errorLockerUpdateNeeded,
		Output: expectedFile + ".new",
	}
	var testParams = &InitParams{
		LockerPath: expectedFile,
	}

	testLogger.Level = logrus.DebugLevel
	assert.NoError(t, handleLockerInitPretend(testParams, testState))
	assert.Equal(t, 5, len(testHook.Entries))
	assert.Contains(t, testHook.Entries[1].Message, expectedFile)
	assert.Contains(t, testHook.Entries[2].Message, "new passphrase")
	assert.Contains(t, testHook.Entries[3].Message, testState.Output)
}

func Test_handleLockerInitUpdate(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Error:  errorLockerUpdateNeeded,
	}
	var testParams = &InitParams{
		LockerPath:    filepath.Join(t.TempDir(), "notExisting"),
		OldPassphrase: memguard.NewEnclave([]byte("password")),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.LockerPath); err == nil {
		var testBytes = []byte("cha cha cha cha cha ... Cha!")
		var testHeader *lockerHeader
		var encrypted []byte

		defer filez.CloseSilently(fd)
		testHeader = newLockerHeader()

		if encrypted, err = testHeader.Encrypt(memguard.NewEnclave(testBytes), testParams.OldPassphrase); err == nil {
			if err = writeLockerData(testHeader, fd, encrypted); err == nil {
				assert.NoError(t, handleLockerInitUpdate(testParams, testState))
				assert.False(t, testState.Completed)
				return
			}
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleLockerInitUpdate_ChaChaError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Error:  errorLockerUpdateNeeded,
	}
	var testParams = &InitParams{
		LockerPath:    filepath.Join(t.TempDir(), "notExisting"),
		OldPassphrase: memguard.NewEnclave([]byte("password")),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.LockerPath); err == nil {
		var testHeader *lockerHeader

		defer filez.CloseSilently(fd)
		testHeader = newLockerHeader()

		if err = writeLockerData(testHeader, fd, []byte("cha cha cha cha cha ... Cha!")); err == nil {
			assert.Error(t, handleLockerInitUpdate(testParams, testState))
			assert.True(t, testState.Completed)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_handleLockerInitUpdate_NoNeededError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleLockerInitUpdate(&InitParams{}, testState), expectedError)
	assert.True(t, testState.Completed)
}

func Test_handleLockerInitUpdate_ReadError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Error:  errorLockerUpdateNeeded,
	}
	var testParams = &InitParams{
		LockerPath: filepath.Join(t.TempDir(), "notExisting"),
	}

	assert.Error(t, handleLockerInitUpdate(testParams, testState))
	assert.True(t, testState.Completed)
}
