package ToOcto

import (
	"io"
	"math/rand"
	"os"
	"time"
)

/*
Dispose: Disposes the cached data.
ResetReadingState: Resets the current reading state of the cached data.
Read: implements the io.Reader interface.
*/
type CachedReader interface {
	Dispose()
	ReadCount() int64
	IsCached() bool
	ResetReadingState()
	Read(p []byte) (int, error)
}

type cachedReader struct {
	cached              bool
	read_count          int64
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

func (cr *cachedReader) ReadCount() int64 {
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
	cr.read_count += int64(count)
	return count, err
}

func NewCachedReader(reader io.Reader) CachedReader {
	const s = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz123456789"
	rand.Seed(time.Now().UnixNano())
	rand_byte := make([]byte, 5)
	for i := range rand_byte {
		rand_byte[i] = s[byte(rand.Intn(len(s)))]
	}
	rand_str := string(rand_byte)
	file, _ := os.Create(rand_str)
	return &cachedReader{false, 0, rand_str, reader, file, nil}
}
