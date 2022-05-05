package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yukitsune/logrus-loki"
)

func main() {

	opts := logrusloki.NewLokiHookOptions().
		WithLevelMap(logrusloki.LevelMap{logrus.PanicLevel: "critical"}).
		WithStaticLabels(logrusloki.Labels{"app": "example"})

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
