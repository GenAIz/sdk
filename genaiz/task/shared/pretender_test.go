package shared

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyValueFilePretender_PretendDelete(t *testing.T) {
	var expectedKey = "key"
	var testPretender = NewConfigPretender("config.yaml")
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	testPretender.PretendDelete(expectedKey)

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedKey)
}

func TestKeyValueFilePretender_PretendDeleteByField(t *testing.T) {
	var expectedKey = "key"
	var expectedField = "field"
	var expectedValue = "value"
	var testPretender = NewConfigPretender("config.yaml")
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	testPretender.PretendDeleteByField(expectedKey, expectedField, expectedValue)

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedKey)
	assert.Contains(t, output, expectedField)
	assert.Contains(t, output, expectedValue)
}

func TestKeyValueFilePretender_PretendSlice(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "value"
	var testPretender = NewConfigPretender("config.yaml")
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	testPretender.PretendSlice(expectedKey, []string{expectedValue})

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedKey)
	assert.Contains(t, output, expectedValue)
}

func TestKeyValueFilePretender_PretendValue(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "value"
	var testPretender = NewConfigPretender("config.yaml")
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	testPretender.PretendValue(expectedKey, expectedValue)

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedKey)
	assert.Contains(t, output, expectedValue)
}

func TestNewConfigPretender_ConfigTypeJson(t *testing.T) {
	assert.Panics(t, func() { NewConfigPretender("config.json") })
}

func TestNewConfigPretender_ConfigTypeNone(t *testing.T) {
	assert.Panics(t, func() { NewConfigPretender("") })
}

func TestNewConfigPretender_ConfigTypeToml(t *testing.T) {
	assert.Panics(t, func() { NewConfigPretender("config.toml") })
}

func TestPretendDelete(t *testing.T) {
	var expectedKey = "key"
	var testPretender = NewConfigPretender("config.yaml")
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	PretendDelete(testPretender, func() string { return "" })
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Empty(t, output)

	r, w, _ = os.Pipe()
	os.Stdout = w
	PretendDelete(testPretender, func() string { return expectedKey })
	_ = w.Close()
	b, _ = io.ReadAll(r)
	output = string(b)
	assert.Contains(t, output, expectedKey)
}

func TestPretendDeleteByField(t *testing.T) {
	var expectedField = "field"
	var expectedKey = "key"
	var expectedValue = "value"
	var testPretender = NewConfigPretender("config.yaml")
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	PretendDeleteByField(testPretender, func() (string, string, string) { return "", "", "" })
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Empty(t, output)

	r, w, _ = os.Pipe()
	os.Stdout = w
	PretendDeleteByField(testPretender, func() (string, string, string) { return expectedKey, "", "" })
	_ = w.Close()
	b, _ = io.ReadAll(r)
	output = string(b)
	assert.Empty(t, output)

	r, w, _ = os.Pipe()
	os.Stdout = w
	PretendDeleteByField(testPretender, func() (string, string, string) { return expectedKey, expectedField, "" })
	_ = w.Close()
	b, _ = io.ReadAll(r)
	output = string(b)
	assert.Empty(t, output)

	r, w, _ = os.Pipe()
	os.Stdout = w
	PretendDeleteByField(testPretender, func() (string, string, string) {
		return expectedKey, expectedField, expectedValue
	})
	_ = w.Close()
	b, _ = io.ReadAll(r)
	output = string(b)
	assert.Contains(t, output, expectedKey)
}

func TestPretendMap(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "value"
	var notExpectedKey = "notExpected"
	var testPretender = NewConfigPretender("config.yaml")
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	PretendMap(testPretender, func() map[string]string {
		return map[string]string{
			expectedKey:    expectedValue,
			notExpectedKey: "",
		}
	})
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, expectedKey)
	assert.Contains(t, output, expectedValue)
	assert.NotContains(t, output, notExpectedKey)
}

func TestPretendSlice(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "value"
	var notExpectedKey = "notExpected"
	var testPretender = NewConfigPretender("config.yaml")
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	PretendSlice(testPretender, func() (string, []string) {
		return expectedKey, []string{expectedValue}
	})
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, expectedKey)
	assert.Contains(t, output, expectedValue)

	r, w, _ = os.Pipe()
	os.Stdout = w
	PretendSlice(testPretender, func() (string, []string) {
		return notExpectedKey, nil
	})
	_ = w.Close()
	b, _ = io.ReadAll(r)
	output = string(b)
	assert.NotContains(t, output, notExpectedKey)
}

func TestPretendValue(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "value"
	var notExpectedKey = "notExpected"
	var testPretender = NewConfigPretender("config.yaml")
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	PretendValue(testPretender, func() (string, string) {
		return expectedKey, expectedValue
	})
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, expectedKey)
	assert.Contains(t, output, expectedValue)

	r, w, _ = os.Pipe()
	os.Stdout = w
	PretendValue(testPretender, func() (string, string) {
		return notExpectedKey, ""
	})
	_ = w.Close()
	b, _ = io.ReadAll(r)
	output = string(b)
	assert.NotContains(t, output, notExpectedKey)
}
