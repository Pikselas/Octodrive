package Octo

import "io"

type SourceLimiter interface {
	IsEOF() bool
	GetCurrentSize() uint64
	Read(p []byte) (n int, err error)
}

type sourceLimiter struct {
	source      io.Reader
	maxSize     uint64
	currentSize uint64
	sourceEOF   bool
}

func (s *sourceLimiter) Read(p []byte) (n int, err error) {
	if s.currentSize >= s.maxSize {
		return 0, io.EOF
	}
	n, err = s.source.Read(p)
	s.currentSize += uint64(n)
	if err == io.EOF {
		s.sourceEOF = true
	}
	return
}

func (s *sourceLimiter) IsEOF() bool {
	return s.sourceEOF
}

func (s *sourceLimiter) GetCurrentSize() uint64 {
	return s.currentSize
}

func NewSourceLimiter(source io.Reader, maxSize uint64) SourceLimiter {
	return &sourceLimiter{
		source:  source,
		maxSize: maxSize,
	}
}
