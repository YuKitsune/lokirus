package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yukitsune/logrus-loki"
)

func main() {

	opts := logrusloki.NewLokiHookOptions().
		// Grafana doesn't have a panic level, but it does have a critical level
		// https://grafana.com/docs/grafana/latest/explore/logs-integration/
		WithLevelMap(logrusloki.LevelMap{logrus.PanicLevel: "critical"}).
		WithStaticLabels(logrusloki.Labels{
			"app":         "example",
			"environment": "development",
		})

	hook := logrusloki.NewLokiHookWithOpts(
		"http://localhost:3100",
		opts,
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel)

	logger := logrus.New()
	logger.AddHook(hook)
	logger.WithField("some_key", "some value").Traceln("trace")
	logger.WithField("some_key", "some value").Debugln("debug")
	logger.WithField("some_key", "some value").Infoln("info")
	logger.WithField("some_key", "some value").Warnln("warning")
	logger.WithField("some_key", "some value").Errorln("error")
	// logger.WithField("some_key", "some value").Fatalln("fatal")
	logger.WithField("some_key", "some value").Panicln("panic")
}
