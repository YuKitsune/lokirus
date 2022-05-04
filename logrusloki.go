package logrusloki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yukitsune/logrus-loki/internal/loki"
	"io/ioutil"
	"net/http"
)

const pushLogsPath = "/loki/api/v1/push"

// LokiHookOptions allows the bahviour of the LokiHook to be customised
type LokiHookOptions struct {
	// LevelMap allows logrus.Levels to be re-mapped to a different label
	LevelMap map[logrus.Level]string

	// Labels allow for custom labels to be added to the loki logs
	Labels map[string]string

	// HttpClient allows for a custom http.Client to be used when pushing logs to loki
	HttpClient *http.Client

	// Todo: Add batching
}

func newDefaultOptions() *LokiHookOptions {
	return &LokiHookOptions{
		map[logrus.Level]string{},
		map[string]string{},
		http.DefaultClient,
	}
}

// LokiHook sends logs to Loki via HTTP.
type LokiHook struct {
	endpoint string
	levels   []logrus.Level
	opts     *LokiHookOptions
}

// NewLokiHook creates a Loki hook for logrus
func NewLokiHook(host string, levels ...logrus.Level) *LokiHook {
	return NewLokiHookWithOpts(host, levels, newDefaultOptions())
}

// NewLokiHookWithOpts creates a Loki hook for logrus with the specified LokiHookOptions
func NewLokiHookWithOpts(host string, levels []logrus.Level, opts *LokiHookOptions) *LokiHook {
	endpoint := fmt.Sprintf("%s%s", host, pushLogsPath)
	return &LokiHook{
		endpoint: endpoint,
		levels:   levels,
		opts:     opts,
	}
}

// Fire sends a log entry to Loki
func (hook *LokiHook) Fire(entry *logrus.Entry) error {

	// Build the stream
	stream := loki.NewStream()

	// Todo: Allow custom message formatter

	// Add the log message and time
	stream.AddEntry(entry.Time, entry.Message)

	// Add the fields as labels
	// Todo: Look into how the message formatter will affect this
	for k, v := range entry.Data {
		stream.AddLabel(k, fmt.Sprintf("%s", v))
	}

	// Add any custom labels we may have configured
	if hook.opts.Labels != nil {
		for k, v := range hook.opts.Labels {
			stream.AddLabel(k, v)
		}
	}

	// Add the log level label
	level := hook.getLevel(entry.Level)
	stream.AddLabel("level", level)

	// Build the batch
	batch := loki.NewBatch()
	batch.AddStream(stream)

	// Convert to JSON
	data, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	// Build the request
	req, err := http.NewRequest("POST", hook.endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := hook.getClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !isSuccessStatusCode(resp.StatusCode) {
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("error posting loki batch (%s): %s", resp.Status, string(data))
	}

	return nil
}

// Levels returns the levels for which Fire will be called
func (hook *LokiHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *LokiHook) getLevel(level logrus.Level) string {
	if hook.opts.LevelMap != nil {
		altValue, hasAltValue := hook.opts.LevelMap[level]
		if hasAltValue {
			return altValue
		}
	}

	return level.String()
}

func (hook *LokiHook) getClient() *http.Client {
	if hook.opts.HttpClient != nil {
		return hook.opts.HttpClient
	}

	return http.DefaultClient
}

func isSuccessStatusCode(code int) bool {
	return code >= 200 && code <= 299
}
