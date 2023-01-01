package Octo

import "io"

type SourceLimiter struct {
	Source      io.Reader
	MaxSize     int64
	CurrentSize int64
	SourceEOF   bool
}

func (s *SourceLimiter) Read(p []byte) (n int, err error) {
	if s.CurrentSize >= s.MaxSize {
		return 0, io.EOF
	}
	n, err = s.Source.Read(p)
	s.CurrentSize += int64(n)
	if err == io.EOF {
		s.SourceEOF = true
	}
	return
}

func (s *SourceLimiter) IsEOF() bool {
	return s.SourceEOF
}
