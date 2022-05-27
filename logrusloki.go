package lokirus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/yukitsune/lokirus/loki"
)

const pushLogsPath = "/loki/api/v1/push"
const levelKey = "level"

// LokiHook sends logs to Loki via HTTP.
type LokiHook struct {
	host   string
	levels []logrus.Level
	opts   LokiHookOptions
}

// NewLokiHook creates a Loki hook for logrus
// if no levels are provided, then logrus.AllLevels is used
func NewLokiHook(host string, levels ...logrus.Level) *LokiHook {
	return NewLokiHookWithOpts(host, NewLokiHookOptions(), levels...)
}

// NewLokiHookWithOpts creates a Loki hook for logrus with the specified LokiHookOptions
// if no levels are provided, then logrus.AllLevels is used
func NewLokiHookWithOpts(host string, opts LokiHookOptions, levels ...logrus.Level) *LokiHook {
	if len(levels) == 0 {
		levels = logrus.AllLevels
	}

	return &LokiHook{
		host:   host,
		levels: levels,
		opts:   opts,
	}
}

// Fire sends a log entry to Loki
func (hook *LokiHook) Fire(entry *logrus.Entry) error {

	messageData, err := hook.formatMessage(entry)
	if err != nil {
		return fmt.Errorf("error formatting message: %s", err.Error())
	}

	stream := loki.NewStream()
	stream.AddEntry(entry.Time, string(messageData))
	hook.applyDynamicLabels(entry, stream)
	hook.applyStaticLabels(stream)
	hook.applyLevelLabel(entry, stream)

	batch := loki.NewBatch()
	batch.AddStream(stream)

	req, err := hook.buildRequest(batch)
	if err != nil {
		return fmt.Errorf("error building request: %s", err.Error())
	}

	err = hook.sendRequest(req)
	return err
}

// Levels returns the levels for which Fire will be called
func (hook *LokiHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *LokiHook) formatMessage(entry *logrus.Entry) ([]byte, error) {
	formatter := hook.opts.Formatter()
	return formatter.Format(entry)
}

func (hook *LokiHook) applyDynamicLabels(entry *logrus.Entry, stream *loki.Stream) {
	fn := hook.opts.DynamicLabelProvider()
	dynamicLabels := fn(entry)

	for k, v := range dynamicLabels {
		stream.AddLabel(k, v)
	}
}

func (hook *LokiHook) applyStaticLabels(stream *loki.Stream) {
	for k, v := range hook.opts.StaticLabels() {
		stream.AddLabel(k, v)
	}
}

func (hook *LokiHook) applyLevelLabel(entry *logrus.Entry, stream *loki.Stream) {
	level := hook.getLevel(entry.Level)
	stream.AddLabel(levelKey, level)
}

func (hook *LokiHook) getLevel(level logrus.Level) string {
	altValue, hasAltValue := hook.opts.LevelMap()[level]
	if hasAltValue {
		return altValue
	}

	return level.String()
}

func (hook *LokiHook) buildRequest(batch *loki.Batch) (*http.Request, error) {

	endpoint := fmt.Sprintf("%s%s", hook.host, pushLogsPath)

	data, err := json.Marshal(batch)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if hook.opts.BasicAuth() != nil {
		req.SetBasicAuth(hook.opts.BasicAuth().Username, hook.opts.BasicAuth().Password)
	}

	return req, nil
}

func (hook *LokiHook) sendRequest(req *http.Request) error {

	resp, err := hook.opts.HttpClient().Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if !isSuccessStatusCode(resp.StatusCode) {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response (%s): %s", resp.Status, err.Error())
		}

		return fmt.Errorf("error posting loki batch (%s): %s", resp.Status, string(data))
	}

	return nil
}

func isSuccessStatusCode(code int) bool {
	return code >= 200 && code <= 299
}
