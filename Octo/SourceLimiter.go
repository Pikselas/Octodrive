package Octo

import "io"

type SourceLimiter interface {
	IsEOF() bool
	GetCurrentSize() int64
	Read(p []byte) (n int, err error)
}

type sourceLimiter struct {
	source      io.Reader
	maxSize     int64
	currentSize int64
	sourceEOF   bool
}

func (s *sourceLimiter) Read(p []byte) (n int, err error) {
	if s.currentSize >= s.maxSize {
		return 0, io.EOF
	}
	n, err = s.source.Read(p)
	s.currentSize += int64(n)
	if err == io.EOF {
		s.sourceEOF = true
	}
	return
}

func (s *sourceLimiter) IsEOF() bool {
	return s.sourceEOF
}

func (s *sourceLimiter) GetCurrentSize() int64 {
	return s.currentSize
}

func NewSourceLimiter(source io.Reader, maxSize int64) SourceLimiter {
	return &sourceLimiter{
		source:  source,
		maxSize: maxSize,
	}
}
