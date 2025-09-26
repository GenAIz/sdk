package broker

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task"
)

type stubLoginClient struct {
	client
	session   *AuthSession
	logoutErr error
}

func (s stubLoginClient) Login(username string, password []byte) (*AuthSession, error) {
	if s.session != nil {
		return s.session, nil
	}

	return nil, errors.New("login error")
}

func (s stubLoginClient) Logout(id string) error {
	return s.logoutErr
}

func (s stubLoginClient) Session() (*Session, error) {
	if s.session == nil {
		return nil, errors.New("session error")
	}

	return &Session{
		UserId: s.UserId,
	}, nil
}

func (s stubLoginClient) SessionValid(session *Session) bool {
	if session.UserId > 0 {
		return s.client.SessionValid(session)

	}

	return true
}

func TestAuthData_Find_NoBrokerSession(t *testing.T) {
	var authData = &AuthData{}
	var _, err = authData.Find("hostAddr")

	assert.ErrorIs(t, ErrorNoSession, err)
}

func TestAuthData_Find(t *testing.T) {
	var expectedAddr = "hostAddr"
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: expectedAddr,
			},
		},
	}
	var actual, err = authData.Find(expectedAddr)

	assert.NoError(t, err)
	assert.Same(t, authData.Accounts[0], actual)
}

func TestAuthData_ForHostUser_NoHostNoSession(t *testing.T) {
	var expectedUser = "username"
	var authData = &AuthData{}
	var _, err = authData.ForHostUser("", expectedUser)

	assert.ErrorIs(t, ErrorNoSession, err)
}

func TestAuthData_ForHostUser_NoHostNoUser(t *testing.T) {
	var authData = &AuthData{}
	var _, err = authData.ForHostUser("", "")

	assert.ErrorIs(t, ErrorNoSession, err)
}

func TestAuthData_ForHostUser_NoHost(t *testing.T) {
	var expectedUser = "username"
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: "hostAddr",
				AuthSession: &AuthSession{
					Username: expectedUser,
				},
			},
		},
	}
	var actual, err = authData.ForHostUser("", expectedUser)

	assert.NoError(t, err)
	assert.Same(t, authData.Accounts[0], actual)
}

func TestAuthData_ForHostUser(t *testing.T) {
	var expectedUser = "username"
	var expectedHost = "hostAddr"
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: "hostAddr2",
				AuthSession: &AuthSession{
					Username: expectedUser,
				},
			},
			{
				HostAddr: expectedHost,
				AuthSession: &AuthSession{
					Username: expectedUser,
				},
			},
		},
	}
	var actual, err = authData.ForHostUser(expectedHost, expectedUser)

	assert.NoError(t, err)
	assert.Same(t, authData.Accounts[1], actual)
}

func TestAuthData_ForToken_NoSession(t *testing.T) {
	var expectedToken = "token"
	var authData = &AuthData{}
	var _, err = authData.ForToken(expectedToken)

	assert.ErrorIs(t, ErrorNoSession, err)
}

func TestAuthData_ForToken(t *testing.T) {
	var expectedToken = "token"
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: "hostAddr",
				AuthSession: &AuthSession{
					Token: expectedToken,
				},
			},
		},
	}
	var actual, err = authData.ForToken(expectedToken)

	assert.NoError(t, err)
	assert.Same(t, authData.Accounts[0], actual)
}

func TestAuthData_Push(t *testing.T) {
	var expectedHost = "hostAddr"
	var authSession = &AuthSession{
		Token: "token",
	}
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: expectedHost,
			},
			{
				HostAddr: "hostAddr2",
			},
		},
	}
	var actual = authData.Push(expectedHost, authSession)

	assert.Equal(t, authSession.Token, actual.Accounts[0].AuthSession.Token)
	assert.Same(t, authData.Accounts[1], actual.Accounts[1])
}

func TestAuthData_Write_PermissionDenied(t *testing.T) {
	var authData = NewAuthData()

	assert.Error(t, authData.Write("/?root"))
}

func TestAuthData_Write(t *testing.T) {
	var expectedFile = filepath.Join(t.TempDir(), "genaiz.login")
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr:    "test",
				AuthSession: &AuthSession{},
			},
		},
	}
	var actual *AuthData

	assert.NoError(t, authData.Write(expectedFile))
	assert.NoError(t, filez.IsReadable(expectedFile))
	actual = NewAuthData(expectedFile)
	assert.Equal(t, authData.Accounts[0].HostAddr, actual.Accounts[0].HostAddr)
}

func TestAuthSession_IsExpired(t *testing.T) {
	assert.True(t, NewAuthSession(&Session{Expiry: 0}, "username", "token").IsExpired())
	assert.False(t, NewAuthSession(&Session{Expiry: -1}, "username", "token").IsExpired())
}

func TestNewLoginTask(t *testing.T) {
	var actual = NewLoginTask()

	assert.NotEmpty(t, actual.Name)
	assert.NotEmpty(t, actual.OnPrepare)
	assert.NotEmpty(t, actual.OnComplete)
	assert.NotEmpty(t, actual.OnPretend)
}

func TestNewLogoutTask(t *testing.T) {
	var actual = NewLogoutTask()

	assert.NotEmpty(t, actual.Name)
	assert.NotEmpty(t, actual.OnPrepare)
	assert.NotEmpty(t, actual.OnComplete)
	assert.NotEmpty(t, actual.OnPretend)
}

func TestNewSessionTask(t *testing.T) {
	var actual = NewSessionTask()

	assert.NotEmpty(t, actual.Name)
	assert.NotEmpty(t, actual.OnPrepare)
	assert.NotEmpty(t, actual.OnComplete)
}

func Test_handleLoginContext_EmptyFile(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
		}

		assert.NoError(t, handleLoginContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLoginContext(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var authData = NewAuthData(fd.Name())
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
		}

		authData = authData.Push("hostAddr2", &AuthSession{
			Token: "token",
		})
		panicz.PanicIfError(authData.Write(fd.Name()))
		assert.NoError(t, handleLoginContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLoginCreate_LoginErr(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var restoredFactory = clientFactory.New
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
			Username: "username",
			Password: []byte("password"),
		}

		defer func() {
			clientFactory.New = restoredFactory
		}()
		clientFactory.New = func(hostAddr string) Client {
			return &stubLoginClient{}
		}

		assert.Error(t, handleLoginCreate(testParams, testState))
	}
}

func Test_handleLoginCreate_TokenPresent(t *testing.T) {
	assert.NoError(t, handleLoginCreate(&LoginParams{}, &task.State{Output: "token"}))
}

func Test_handleLoginCreate_WriteErr(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var restoredFactory = clientFactory.New
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
			Username: "username",
			Password: []byte("password"),
		}

		panicz.PanicIfError(os.Chmod(fd.Name(), 0400))

		defer func() {
			clientFactory.New = restoredFactory
		}()
		clientFactory.New = func(hostAddr string) Client {
			return &stubLoginClient{
				session: &AuthSession{
					Token:    "token",
					Username: "user",
				},
			}
		}

		assert.Error(t, handleLoginCreate(testParams, testState))
	}
}

func Test_handleLoginCreate(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var expectedHost = "hostAddr"
		var expectedUser = "username"
		var expectedToken = "token"
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: expectedHost,
			},
			Username: expectedUser,
			Password: []byte("password"),
		}
		var restoredFactory = clientFactory.New
		var actual *AuthAccount

		defer func() {
			clientFactory.New = restoredFactory
		}()
		clientFactory.New = func(hostAddr string) Client {
			return &stubLoginClient{
				session: &AuthSession{
					Token:    expectedToken,
					Username: expectedUser,
				},
			}
		}

		assert.NoError(t, handleLoginCreate(testParams, testState))

		if actual, err = NewAuthData(fd.Name()).ForHostUser(expectedHost, expectedUser); err == nil {
			assert.Equal(t, actual.Token, expectedToken)
		}

		assert.NoError(t, err)
	}
}

func Test_handleLoginDelete_NoSession(t *testing.T) {
	var testState = &task.State{}
	var testParams = &LoginParams{}

	assert.ErrorIs(t, ErrorNoSession, handleLoginDelete(testParams, testState))
}

func Test_handleLoginDelete_LogoutFailed(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedSessionId = int64(37)
		var testState = &task.State{
			Logger: logrus.New(),
			Output: cast.ToString(expectedSessionId),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
		}
		var restoredFactory = clientFactory.Get

		defer func() {
			clientFactory.Get = restoredFactory
		}()
		clientFactory.Get = func(authFile, addr string) (Client, error) {
			return &stubLoginClient{
				logoutErr: errors.New("unexpected"),
			}, nil
		}

		assert.NoError(t, handleLoginDelete(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLoginDelete(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedHost = "hostAddr"
		var expectedSessionId = int64(37)
		var testSession = &AuthSession{
			SessionId: expectedSessionId,
		}
		var testState = &task.State{
			Logger: logrus.New(),
			Output: cast.ToString(expectedSessionId),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: expectedHost,
			},
		}
		var restoredFactory = clientFactory.Get

		defer func() {
			clientFactory.Get = restoredFactory
		}()
		clientFactory.Get = func(authFile, addr string) (Client, error) {
			return &stubLoginClient{
				session: testSession,
			}, nil
		}

		panicz.PanicIfError(NewAuthData().
			Push(expectedHost, testSession).
			Push("hostAddr2", &AuthSession{}).
			Write(fd.Name()))
		assert.NoError(t, handleLoginDelete(testParams, testState))
		assert.Equal(t, 1, len(NewAuthData(fd.Name()).Accounts))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLoginPretend_AlreadyAuth(t *testing.T) {
	var testState = &task.State{
		Error:  errors.New("not_the_right_one"),
		Logger: logrus.New(),
	}
	var testParams = &LoginParams{
		Broker: &Broker{
			HostAddr: "hostAddr",
		},
	}

	assert.NoError(t, handleLoginPretend(testParams, testState))
}

func Test_handleLoginPretend(t *testing.T) {
	var expectedUser = "username"
	var expectedHost = "hostAddr"
	var testState = &task.State{
		Error:  ErrorNoAuth,
		Logger: logrus.New(),
	}
	var testParams = &LoginParams{
		Broker: &Broker{
			HostAddr: expectedHost,
		},
		Username: expectedUser,
	}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleLoginPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedUser)
	assert.Contains(t, output, expectedHost)
}

func Test_handleLogoutContext_NoAuth(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
			},
		}

		assert.ErrorIs(t, ErrorNoLogin, handleLogoutContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext_NoForHostUser(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var testSession = &AuthSession{
			Token: "token",
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
			Username: "username",
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr2", testSession).
			Write(fd.Name()))
		assert.ErrorIs(t, ErrorNoSession, handleLogoutContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext_SingleLogin(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedSessionId = int64(37)
		var testSession = &AuthSession{
			Token:     "token",
			SessionId: expectedSessionId,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
			},
		}
		var testAuth = NewAuthData().
			Push("hostAddr2", testSession)

		// Intentionally set an invalid active account
		testAuth.Active = 1
		panicz.PanicIfError(testAuth.Write(fd.Name()))
		assert.NoError(t, handleLogoutContext(testParams, testState))
		assert.Equal(t, cast.ToString(expectedSessionId), testState.Output)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext_TooManyLogins(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var testSession = &AuthSession{
			Token: "token",
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
			},
		}
		var testAuth = NewAuthData().
			Push("hostAddr2", testSession).
			Push("hostAddr3", testSession)

		// Intentionally set an invalid active account
		testAuth.Active = 2
		panicz.PanicIfError(testAuth.Write(fd.Name()))
		assert.ErrorIs(t, handleLogoutContext(testParams, testState), ErrorSessionConflict)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext_UsernameLogin(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedSessionId = int64(37)
		var expectedUser = "username"
		var testSession = &AuthSession{
			Token:     "token",
			SessionId: expectedSessionId,
			Username:  expectedUser,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
			Username: expectedUser,
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.NoError(t, handleLogoutContext(testParams, testState))
		assert.Equal(t, cast.ToString(expectedSessionId), testState.Output)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedSessionId = int64(37)
		var expectedUser = "username"
		var testSession = &AuthSession{
			Token:     "token",
			SessionId: expectedSessionId,
			Username:  expectedUser,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
			},
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.NoError(t, handleLogoutContext(testParams, testState))
		assert.Equal(t, cast.ToString(expectedSessionId), testState.Output)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutPretend_NoSession(t *testing.T) {
	var testError = errors.New("expected")
	var testState = &task.State{
		Error: testError,
	}
	var testParams = &LoginParams{}

	assert.Same(t, testError, handleLogoutPretend(testParams, testState))
}

func Test_handleLogoutPretend_NoHostSession(t *testing.T) {
	var expectedSessionId = int64(37)
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error:  errors.New("test"),
		Logger: logrus.New(),
		Output: cast.ToString(expectedSessionId),
	}
	var testParams = &LoginParams{
		Broker: &Broker{
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

	assert.ErrorIs(t, expectedError, handleLogoutPretend(testParams, testState))
}

func Test_handleLogoutPretend(t *testing.T) {
	var expectedSessionId = int64(37)
	var testError = errors.New("expected")
	var testState = &task.State{
		Error:  testError,
		Logger: logrus.New(),
		Output: cast.ToString(expectedSessionId),
	}
	var testParams = &LoginParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubLoginClient{
			session: &AuthSession{
				Token:    "token",
				Username: "user",
			},
		}, nil
	}

	assert.NoError(t, handleLogoutPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(expectedSessionId))
}

func Test_handleSessionContext_NoAuth(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testBroker = &Broker{
		AuthFile: "not_exist",
	}

	assert.ErrorIs(t, ErrorNoAuth, handleSessionContext(testBroker, testState))
}

func Test_handleSessionContext_NoHostSession(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedToken = "token"
		var testSession = &AuthSession{
			Expiry: time.Now().Add(time.Hour).UnixMilli(),
			Token:  expectedToken,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testBroker = &Broker{
			AuthFile: fd.Name(),
			HostAddr: "hostAddr2",
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.ErrorIs(t, ErrorNoSession, handleSessionContext(testBroker, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleSessionContext_ActiveExpired(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var testSession = &AuthSession{
			Expiry: 0,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testBroker = &Broker{
			AuthFile: fd.Name(),
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.ErrorIs(t, ErrorNoAuth, handleSessionContext(testBroker, testState))
	}
}

func Test_handleSessionContext_Active(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedToken = "token"
		var testSession = &AuthSession{
			Expiry: time.Now().Add(time.Hour).UnixMilli(),
			Token:  expectedToken,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testBroker = &Broker{
			AuthFile: fd.Name(),
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.NoError(t, handleSessionContext(testBroker, testState))
		assert.Equal(t, expectedToken, testState.Output)
	}
}

func Test_handleSessionContext(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedHost = "hostAddr"
		var expectedToken = "token"
		var testSession = &AuthSession{
			Expiry: time.Now().Add(time.Hour).UnixMilli(),
			Token:  expectedToken,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testBroker = &Broker{
			AuthFile: fd.Name(),
			HostAddr: expectedHost,
		}

		panicz.PanicIfError(NewAuthData().
			Push(expectedHost, testSession).
			Write(fd.Name()))
		assert.NoError(t, handleSessionContext(testBroker, testState))
		assert.Equal(t, expectedToken, testState.Output)
	}
}

func Test_handleSessionValidate_NoAuth(t *testing.T) {
	assert.ErrorIs(t, ErrorNoAuth, handleSessionValidate(&Broker{}, &task.State{}))
}

func Test_handleSessionValidate_NoHostSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error:  errors.New("test"),
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &Broker{
		AuthFile: "file",
		HostAddr: "hostAddr",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, expectedError, handleSessionValidate(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleSessionValidate_SessionError(t *testing.T) {
	var testState = &task.State{
		Error:  errors.New("test"),
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &Broker{
		AuthFile: "file",
		HostAddr: "hostAddr",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubLoginClient{}, nil
	}

	assert.Error(t, handleSessionValidate(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleSessionValidate_SessionInvalid(t *testing.T) {
	var testState = &task.State{
		Error:  errors.New("test"),
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &Broker{
		AuthFile: "file",
		HostAddr: "hostAddr",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubLoginClient{
			client: client{
				UserId: 37,
			},
			session: &AuthSession{},
		}, nil
	}

	assert.ErrorIs(t, ErrorSessionInvalid, handleSessionValidate(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleSessionValidate(t *testing.T) {
	var expectedOutput = "session"
	var testState = &task.State{
		Error:  errors.New("test"),
		Logger: logrus.New(),
		Output: expectedOutput,
	}
	var testParams = &Broker{
		AuthFile: "file",
		HostAddr: "hostAddr",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubLoginClient{
			client: client{
				UserId: 0,
			},
			session: &AuthSession{},
		}, nil
	}

	assert.NoError(t, handleSessionValidate(testParams, testState))
	assert.Equal(t, expectedOutput, testState.Output)
}
