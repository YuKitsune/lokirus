package loki

type batch struct {
	Streams []stream `json:"streams"`
}

func NewBatch() *batch {
	return &batch{[]stream{}}
}

func (b *batch) AddStream(s *stream) {
	b.Streams = append(b.Streams, *s)
}
