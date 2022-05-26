package main

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yukitsune/lokirus"
)

func main() {

	// Configure the Loki hook
	opts := lokirus.NewLokiHookOptions().
		// Grafana doesn't have a "panic" level, but it does have a "critical" level
		// https://grafana.com/docs/grafana/latest/explore/logs-integration/
		WithLevelMap(lokirus.LevelMap{logrus.PanicLevel: "critical"}).
		WithStaticLabels(lokirus.Labels{
			"app":         "example",
			"environment": "development",
		})

	hook := lokirus.NewLokiHookWithOpts(
		"http://localhost:3100",
		opts,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel)

	// Configure the logger
	logger := logrus.New()
	logger.AddHook(hook)

	// Log all the things!
	levels := []logrus.Level{
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	}

	messages := []string{
		"Road work ahead? Uh yea, I sure hope it does.",
		"Merry Chrysler.",
		"Do it for the vine.",
		"It is Wednesday my dudes.",
		"A potato flew around my room before you came.",
		"Hi my name is Trey I have a basketball game tomorrow.",
		"Mother trucker dude, that hurt like a butt cheek on a stick.",
		"I could have dropped my croissant!",
		"Deez nuts, ha got em!",
	}

	kvps := []struct {
		key   string
		value string
	}{
		{"foo", "bar"},
		{"biz", "baz"},
		{"fizz", "buzz"},
		{"9+10", "21"},
		{"hotel", "trivago"},
	}

	for range time.Tick(1 * time.Second) {

		i := rand.Intn(len(levels))
		level := levels[i]

		i = rand.Intn(len(messages))
		message := messages[i]

		i = rand.Intn(len(kvps))
		kvp := kvps[i]

		logger.WithField(kvp.key, kvp.value).Log(level, message)
	}
}
