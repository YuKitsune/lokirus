package loki

import (
	"strconv"
	"time"
)

type stream struct {
	Labels  map[string]string `json:"stream"`
	Entries [][]string        `json:"values"`
}

func NewStream() *stream {
	return &stream{
		map[string]string{},
		[][]string{},
	}
}

func (s *stream) AddLabel(key string, value string) {
	s.Labels[key] = value
}

func (s *stream) AddEntry(t time.Time, entry string) {
	timeStr := strconv.FormatInt(t.UnixNano(), 10)
	s.Entries = append(s.Entries, []string{timeStr, entry})
}
