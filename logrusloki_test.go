package lokirus_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/yukitsune/lokirus"
	"github.com/yukitsune/lokirus/loki"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

const testFormatterPrefix = "this is a test"

type testFormatter struct {
}

func (f *testFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s: %s", testFormatterPrefix, entry.Message)), nil
}

type mockRoundTripper struct {
	req *http.Request
}

func (r *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.req = req

	return &http.Response{
		StatusCode: http.StatusCreated,
		Body:       bodyReadCloser{bytes.NewBuffer([]byte{})},
	}, nil
}

func (r *mockRoundTripper) UnmarshalRequest(v interface{}) error {

	reqData, err := ioutil.ReadAll(r.req.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(reqData, &v)
	if err != nil {
		return err
	}

	return nil
}

type bodyReadCloser struct {
	io.Reader
}

func (rc bodyReadCloser) Close() error {
	return nil
}

func TestLokiHook_Fires(t *testing.T) {

	// Arrange
	client, roundTripper := getClient()
	hook := lokirus.NewLokiHookWithOpts(
		"",
		lokirus.NewLokiHookOptions().
			WithStaticLabels(lokirus.Labels{"test": t.Name()}).
			WithHttpClient(client))

	logger := logrus.New()
	logger.AddHook(hook)

	// Act
	logger.Info(t.Name())

	var sentBatch loki.Batch
	err := roundTripper.UnmarshalRequest(&sentBatch)
	assert.NoError(t, err)

	// Assert
	// Ensure the label and message were sent
	assert.Equal(t, sentBatch.Streams[0].Labels["test"], t.Name())
	assert.Contains(t, sentBatch.Streams[0].Entries[0][1], t.Name())
}

func TestLokiHook_SendsStaticLabels(t *testing.T) {

	// Arrange
	staticLabels := lokirus.Labels{
		"test":            t.Name(),
		"my_first_label":  "abc",
		"my_second_label": "123",
		"my_third_label":  "!@#",
	}

	client, roundTripper := getClient()
	hook := lokirus.NewLokiHookWithOpts(
		"",
		lokirus.NewLokiHookOptions().
			WithStaticLabels(staticLabels).
			WithHttpClient(client))

	logger := logrus.New()
	logger.AddHook(hook)

	// Ensure we get the same labels every time
	for i := 0; i < 10; i++ {

		// Act
		logger.Infof("test %d", i)

		var sentBatch loki.Batch
		err := roundTripper.UnmarshalRequest(&sentBatch)
		assert.NoError(t, err)

		for k, v := range staticLabels {
			sentValue, hasSentValue := sentBatch.Streams[0].Labels[k]

			// Assert
			// Ensure our static labels are all present
			assert.True(t, hasSentValue)
			assert.Equal(t, v, sentValue)
		}
	}
}

func TestLokiHook_SendsDynamicLabels(t *testing.T) {

	// Arrange

	// Every time we call the label provider, we'll return whatever counter is
	// Counter will be incremented later
	// It's a weird pattern, but it works
	counter := 0
	fn := func(entry *logrus.Entry) lokirus.Labels {
		l := lokirus.Labels{
			"count": strconv.Itoa(counter),
		}

		return l
	}

	client, roundTripper := getClient()
	hook := lokirus.NewLokiHookWithOpts(
		"",
		lokirus.NewLokiHookOptions().
			WithDynamicLabelProvider(fn).
			WithHttpClient(client))

	logger := logrus.New()
	logger.AddHook(hook)

	for counter = 0; counter < 10; counter++ {

		// Act
		logger.Infof(t.Name())

		var sentBatch loki.Batch
		err := roundTripper.UnmarshalRequest(&sentBatch)
		assert.NoError(t, err)

		sentValue, hasSentValue := sentBatch.Streams[0].Labels["count"]

		// Assert
		// Ensure the count label is different every time
		assert.True(t, hasSentValue)
		assert.Equal(t, strconv.Itoa(counter), sentValue)
	}
}

func TestLokiHook_SendsLevelLabel(t *testing.T) {

	// Arrange

	// Intentionally ignoring Fatal and Panic as they will kill the program
	// Debug and Trace aren't sent from hooks
	levels := []logrus.Level{
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}

	client, roundTripper := getClient()
	hook := lokirus.NewLokiHookWithOpts(
		"",
		lokirus.NewLokiHookOptions().
			WithHttpClient(client),
		levels...)

	logger := logrus.New()
	logger.AddHook(hook)

	for _, level := range levels {
		levelStr := level.String()

		// Act
		logger.Log(level, t.Name())

		var sentBatch loki.Batch
		err := roundTripper.UnmarshalRequest(&sentBatch)
		assert.NoError(t, err)

		// Assert
		// Ensure the level label is present and set to the correct value
		assert.Equal(t, levelStr, sentBatch.Streams[0].Labels["level"])
	}
}

func TestLokiHook_ReMapsLevels(t *testing.T) {

	// Arrange

	// Intentionally ignoring Fatal and Panic as they will kill the program
	// Debug and Trace aren't sent from hooks
	levelMap := lokirus.LevelMap{
		logrus.ErrorLevel: "my_error",
		logrus.WarnLevel:  "your_warning",
		// We'll let InfoLevel remain the same so we can test that unmapped levels use their default value
	}

	client, roundTripper := getClient()
	hook := lokirus.NewLokiHookWithOpts(
		"",
		lokirus.NewLokiHookOptions().
			WithLevelMap(levelMap).
			WithHttpClient(client))

	logger := logrus.New()
	logger.AddHook(hook)

	for logrusLevel, customLevel := range levelMap {

		// Act
		logger.Log(logrusLevel, t.Name())

		var sentBatch loki.Batch
		err := roundTripper.UnmarshalRequest(&sentBatch)
		assert.NoError(t, err)

		// Assert
		// Ensure the level label is our custom label
		assert.Equal(t, customLevel, sentBatch.Streams[0].Labels["level"])
	}

	// Test InfoLevel manually since it's not re-mapped

	// Act
	logger.Log(logrus.InfoLevel, t.Name())

	var sentBatch loki.Batch
	err := roundTripper.UnmarshalRequest(&sentBatch)
	assert.NoError(t, err)

	// Assert
	// Ensure the level label still uses the default value with the other mappings present
	assert.Equal(t, logrus.InfoLevel.String(), sentBatch.Streams[0].Labels["level"])
}

//// If any of these tests are passing, then this works...
//func TestLokiHook_UsesCustomHttpClient(t *testing.T) {
//
//}

func TestLokiHook_UsesCustomFormatter(t *testing.T) {

	// Arrange
	client, roundTripper := getClient()
	hook := lokirus.NewLokiHookWithOpts(
		"",
		lokirus.NewLokiHookOptions().
			WithFormatter(&testFormatter{}).
			WithHttpClient(client))

	logger := logrus.New()
	logger.AddHook(hook)

	// Act
	logger.Infoln(t.Name())

	var sentBatch loki.Batch
	err := roundTripper.UnmarshalRequest(&sentBatch)
	assert.NoError(t, err)

	// Assert
	// Ensure the sent log message used the formatter
	sentMessage := sentBatch.Streams[0].Entries[0][1]
	assert.Contains(t, sentMessage, testFormatterPrefix)
}

func getClient() (*http.Client, *mockRoundTripper) {
	client := &http.Client{}
	roundTripper := &mockRoundTripper{}
	client.Transport = roundTripper

	return client, roundTripper
}
