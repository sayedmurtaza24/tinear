package screen

import (
	"sync/atomic"
)

type Size struct {
	width  atomic.Int32
	height atomic.Int32
}

func NewSize(w, h int) *Size {
	size := &Size{
		width:  atomic.Int32{},
		height: atomic.Int32{},
	}

	size.width.Store(int32(w))
	size.height.Store(int32(h))

	return size
}

func (s *Size) Width() int {
	return int(s.width.Load())
}

func (s *Size) Height() int {
	return int(s.height.Load())
}

func (s *Size) SetSize(w, h int) {
	s.width.Store(int32(w))
	s.height.Store(int32(h))
}
