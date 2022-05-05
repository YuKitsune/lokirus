package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yukitsune/logrus-loki"
)

func main() {

	// Configure the Loki hook
	opts := logrusloki.NewLokiHookOptions().
		// Grafana doesn't have a "panic" level, but it does have a "critical" level
		// https://grafana.com/docs/grafana/latest/explore/logs-integration/
		WithLevelMap(logrusloki.LevelMap{logrus.PanicLevel: "critical"}).
		WithStaticLabels(logrusloki.Labels{
			"app":         "example",
			"environment": "development",
		})

	hook := logrusloki.NewLokiHookWithOpts(
		"http://localhost:3100",
		opts,
		logrus.AllLevels...)

	// Configure the logger
	logger := logrus.New()
	logger.AddHook(hook)

	// Log all the things!
	logger.WithField("some_key", "some value").Traceln("trace")
	logger.WithField("some_other_key", "some other value").Debugln("debug")
	logger.WithField("foo", "bar").Infoln("info")
	logger.WithField("fizz", "buzz").Warnln("warning")
	logger.Errorln("error")
	logger.Fatalln("fatal")
}
