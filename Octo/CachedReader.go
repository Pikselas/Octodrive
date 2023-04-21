package Octo

import (
	"io"
	"os"
)

type CachedReader struct {
	cached_file_name string
	src              io.Reader
	closer           io.Closer
}

func (cr *CachedReader) Dispose() {
	if cr.closer != nil {
		cr.closer.Close()
	}
	os.Remove(cr.cached_file_name)
}

func (cr *CachedReader) Reset() error {
	if cr.closer != nil {
		cr.closer.Close()
	}
	file, err := os.Open(cr.cached_file_name)
	if err != nil {
		return err
	}
	cr.src = file
	cr.closer = file
	return nil
}

func (cr *CachedReader) Read(p []byte) (int, error) {
	return cr.src.Read(p)
}

func NewCachedReader(reader io.Reader) (*CachedReader, error) {
	cr := new(CachedReader)
	cr.cached_file_name = RandomString(5)
	file, err := os.Create(cr.cached_file_name)
	if err != nil {
		return nil, err
	}
	cr.closer = file
	cr.src = io.TeeReader(reader, file)
	return cr, nil
}
