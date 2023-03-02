package ToOcto

import (
	"io"
	"os"
)

/*
Dispose: Disposes the cached data.
ResetReadingState: Resets the current reading state of the cached data.
Read: implements the io.Reader interface.
*/
type CachedReader interface {
	Dispose()
	ReadCount() uint64
	IsCached() bool
	ResetReadingState()
	Read(p []byte) (int, error)
}

type cachedReader struct {
	cached              bool
	read_count          uint64
	temp_data_name      string
	current_data_source io.Reader
	place_to_write      io.WriteCloser
	file_closer         io.Closer
}

func (cr *cachedReader) Dispose() {
	if cr.IsCached() {
		if cr.file_closer != nil {
			cr.file_closer.Close()
		}
		os.Remove(cr.temp_data_name)
		cr.cached = false
	}
}

func (cr *cachedReader) IsCached() bool {
	return cr.cached
}

func (cr *cachedReader) ReadCount() uint64 {
	return cr.read_count
}

func (cr *cachedReader) ResetReadingState() {
	if cr.IsCached() {
		if cr.file_closer != nil {
			cr.file_closer.Close()
		}
		cr.read_count = 0
		f, _ := os.Open(cr.temp_data_name)
		cr.current_data_source = f
		cr.file_closer = f
	}
}

func (cr *cachedReader) Read(p []byte) (int, error) {
	count, err := cr.current_data_source.Read(p)
	if !cr.IsCached() {
		cr.place_to_write.Write(p[:count])
		if err == io.EOF {
			cr.place_to_write.Close()
			cr.cached = true
		}
	}
	cr.read_count += uint64(count)
	return count, err
}

func NewCachedReader(reader io.Reader) CachedReader {
	rand_str := RandomString(5)
	file, _ := os.Create(rand_str)
	return &cachedReader{false, 0, rand_str, reader, file, nil}
}
