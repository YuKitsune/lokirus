package lokirus

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// DynamicLabelProviderFunc defines the func for providing custom dynamic labels
// https://grafana.com/docs/loki/latest/best-practices/#use-dynamic-labels-sparingly
type DynamicLabelProviderFunc func(entry *logrus.Entry) Labels

var defaultDynamicLabelProvider = func(entry *logrus.Entry) Labels {
	return make(map[string]string)
}

// LevelMap defines a map between a logrus.Level and a custom level value
type LevelMap map[logrus.Level]string

// Labels defines a set of key-value pairs
type Labels map[string]string

// LokiHookOptions allows for the behaviour of the LokiHook to be customised
type LokiHookOptions interface {

	// LevelMap returns the map of logrus.Levels and custom log level values
	LevelMap() LevelMap

	// StaticLabels returns the static labels added to the loki log stream
	StaticLabels() Labels

	// DynamicLabelProvider returns the DynamicLabelProviderFunc used to populate a log entry with dynamic labels
	DynamicLabelProvider() DynamicLabelProviderFunc

	// HttpClient returns the http.Client to be used for pushing logs to loki
	HttpClient() *http.Client

	// Formatter returns the logrus.Formatter to be used when writing log entries
	Formatter() logrus.Formatter

	// BasicAuth returns the basic authentication credentials to be used
	// when connecting to Loki.
	// nil means that no credentials should be used.
	BasicAuth() *BasicAuthCredentials

	// WithLevelMap allows for logrus.Levels to be re-mapped to a different label value
	WithLevelMap(LevelMap) LokiHookOptions

	// WithStaticLabels allow for multiple static labels to be added to the loki log stream
	// These should be preferred over dynamic labels:
	// https://grafana.com/docs/loki/latest/best-practices/#static-labels-are-good
	WithStaticLabels(Labels) LokiHookOptions

	// WithDynamicLabelProvider allows for dynamic labels to be added to log entries
	// Dynamic labels should be used sparingly:
	// https://grafana.com/docs/loki/latest/best-practices/#use-dynamic-labels-sparingly
	WithDynamicLabelProvider(DynamicLabelProviderFunc) LokiHookOptions

	// WithHttpClient allows for a custom http.Client to be used for pushing logs to loki
	// By default, the http.DefaultClient will be used
	WithHttpClient(*http.Client) LokiHookOptions

	// WithFormatter allows for a custom logrus.Formatter to be used when writing log entries
	// By default, the logrus.TextFormatter will be used
	WithFormatter(logrus.Formatter) LokiHookOptions

	// WithBasicAuth allows to set a username and password to use when writing to the remote Loki host.
	WithBasicAuth(username, password string) LokiHookOptions
}

// BasicAuthCredentials is a structure that holds a username and a password.
// to be used for HTTP queries.
type BasicAuthCredentials struct {
	Username string
	Password string
}

type lokiHookOptions struct {
	levelMap             LevelMap
	staticLabels         Labels
	dynamicLabelProvider DynamicLabelProviderFunc
	httpClient           *http.Client
	formatter            logrus.Formatter
	basicAuthCredentials *BasicAuthCredentials
}

func NewLokiHookOptions() LokiHookOptions {
	return &lokiHookOptions{
		levelMap:             map[logrus.Level]string{},
		staticLabels:         map[string]string{},
		dynamicLabelProvider: defaultDynamicLabelProvider,
		httpClient:           http.DefaultClient,
		formatter:            &logrus.TextFormatter{},
		basicAuthCredentials: nil,
	}
}

func (opt *lokiHookOptions) LevelMap() LevelMap {
	return opt.levelMap
}

func (opt *lokiHookOptions) BasicAuth() *BasicAuthCredentials {
	return opt.basicAuthCredentials
}

func (opt *lokiHookOptions) StaticLabels() Labels {
	return opt.staticLabels
}

func (opt *lokiHookOptions) DynamicLabelProvider() DynamicLabelProviderFunc {
	return opt.dynamicLabelProvider
}

func (opt *lokiHookOptions) HttpClient() *http.Client {
	return opt.httpClient
}

func (opt *lokiHookOptions) Formatter() logrus.Formatter {
	return opt.formatter
}

func (opt *lokiHookOptions) WithBasicAuth(username, password string) LokiHookOptions {
	opt.basicAuthCredentials = &BasicAuthCredentials{
		Username: username,
		Password: password,
	}
	return opt
}

func (opt *lokiHookOptions) WithLevelMap(levelMap LevelMap) LokiHookOptions {
	opt.levelMap = levelMap
	return opt
}

func (opt *lokiHookOptions) WithStaticLabels(staticLabels Labels) LokiHookOptions {
	opt.staticLabels = staticLabels
	return opt
}

func (opt *lokiHookOptions) WithDynamicLabelProvider(dynamicLabelProvider DynamicLabelProviderFunc) LokiHookOptions {
	opt.dynamicLabelProvider = dynamicLabelProvider
	return opt
}

func (opt *lokiHookOptions) WithHttpClient(httpClient *http.Client) LokiHookOptions {
	opt.httpClient = httpClient
	return opt
}

func (opt *lokiHookOptions) WithFormatter(formatter logrus.Formatter) LokiHookOptions {
	opt.formatter = formatter
	return opt
}
