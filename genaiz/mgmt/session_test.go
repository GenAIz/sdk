package mgmt

import (
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task/broker"
)

func TestUserSession_MarshalSlice(t *testing.T) {
	var testUserSession = &UserSession{
		Id:      int64(37),
		UserId:  42,
		Created: int64(1),
		Expiry:  int64(2),
		Expired: false,
		Flags:   new(3),
	}

	if actual, err := testUserSession.MarshalSlice(); err == nil {
		var expectedCreated = createdFormatter.FormatMillis(testUserSession.Created)
		var expectedExpired = expiryFormatter.FormatMillis(testUserSession.Expiry)

		assert.Equal(t, 5, len(actual))
		assert.Equal(t, cast.ToString(testUserSession.Id), actual[0])
		assert.Equal(t, cast.ToString(testUserSession.UserId), actual[1])
		assert.Equal(t, expectedCreated, actual[2])
		assert.Equal(t, expectedExpired, actual[3])
		assert.Equal(t, "no", actual[4])
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserSession_MarshalSlice_NoCreated(t *testing.T) {
	var testUserSession = &UserSession{
		Id:      int64(37),
		UserId:  42,
		Expiry:  int64(1337),
		Expired: false,
		Flags:   new(3),
	}

	if actual, err := testUserSession.MarshalSlice(); err == nil {
		var expectedExpired = expiryFormatter.FormatMillis(testUserSession.Expiry)

		assert.Equal(t, 5, len(actual))
		assert.Equal(t, cast.ToString(testUserSession.Id), actual[0])
		assert.Equal(t, cast.ToString(testUserSession.UserId), actual[1])
		assert.Equal(t, "-", actual[2])
		assert.Equal(t, expectedExpired, actual[3])
		assert.Equal(t, "no", actual[4])
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestToUserSession(t *testing.T) {
	var expectedSession = &broker.Session{
		Id:     int64(37),
		UserId: 42,
		Nco:    int64(1),
		Nms:    int64(2),
		Expiry: time.Now().Add(time.Hour).UTC().UnixMilli(),
		Flags:  3,
	}

	actual := ToUserSession(expectedSession)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedSession.Id, actual.Id)
	assert.Equal(t, expectedSession.UserId, actual.UserId)
	assert.Equal(t, expectedSession.Nco, actual.Created)
	assert.Equal(t, expectedSession.Expiry, actual.Expiry)
	assert.Equal(t, expectedSession.Flags, *actual.Flags)
	assert.False(t, actual.Expired)
}

func TestToUserSession_Expired(t *testing.T) {
	var expectedSession = &broker.Session{
		Id:     int64(37),
		UserId: 42,
		Nco:    int64(1),
		Nms:    int64(2),
		Expiry: time.Now().Add(-1 * time.Minute).UTC().UnixMilli(),
		Flags:  3,
	}

	actual := ToUserSession(expectedSession)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedSession.Id, actual.Id)
	assert.Equal(t, expectedSession.UserId, actual.UserId)
	assert.Equal(t, expectedSession.Nco, actual.Created)
	assert.Equal(t, expectedSession.Expiry, actual.Expiry)
	assert.Equal(t, expectedSession.Flags, *actual.Flags)
	assert.True(t, actual.Expired)

}

func TestToUserSession_Infinite(t *testing.T) {
	var expectedSession = &broker.Session{
		Id:     int64(37),
		UserId: 42,
		Nco:    int64(1),
		Nms:    int64(2),
		Expiry: -1,
		Flags:  3,
	}

	actual := ToUserSession(expectedSession)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedSession.Id, actual.Id)
	assert.Equal(t, expectedSession.UserId, actual.UserId)
	assert.Equal(t, expectedSession.Nco, actual.Created)
	assert.Equal(t, expectedSession.Expiry, actual.Expiry)
	assert.Equal(t, expectedSession.Flags, *actual.Flags)
	assert.False(t, actual.Expired)
}
