package gifstream

import (
	"image"
	"time"
)

type Provider interface {
	Image() *image.Paletted
}

type Streamer struct {
	Provider Provider
	Delay    time.Duration
}

func (s *Streamer) Stream(w Writer) error {
	current := s.Provider.Image()
	int_delay := int(s.Delay / (time.Second / 100))

	e := encoder{
		w:      w,
		bounds: current.Bounds(),
	}

	e.writeHeader()
	if e.err != nil {
		return e.err
	}
	for {
		e.writeImageBlock(current, int_delay)
		e.flush()
		if e.err != nil {
			return e.err
		}
		time.Sleep(s.Delay)
		current = s.Provider.Image()
	}
}
