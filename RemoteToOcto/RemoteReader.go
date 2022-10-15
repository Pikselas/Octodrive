package RemoteToOcto

import (
	"io"
)

/*
	   Reads from source -> Encodes by the encoder and returns.
	   buffer: should be the reading and writing compatible channel / buffer.
			   where the encoder writes it's data after encoding.
	   active_read_state: should be true.
*/
type RemoteReader interface {
	Read(p []byte) (int, error)
	ReadCount() int64
	EncodeCount() int64
	RemoteSourceEnded() bool
}
type reader struct {
	source            io.Reader
	encoder           io.WriteCloser
	buffer            io.Reader
	read_count        int64
	encoding_count    int64
	max_read_count    int64
	active_read_state bool
	source_ended      bool
}

func (r *reader) ReadCount() int64 {
	return r.read_count
}

func (r *reader) EncodeCount() int64 {
	return r.encoding_count
}

func (r *reader) RemoteSourceEnded() bool {
	return r.source_ended
}

func (r *reader) Read(p []byte) (int, error) {
	buff := make([]byte, 1000)
	if r.active_read_state {
		count, err := r.source.Read(buff)
		r.encoder.Write(buff[:count])
		r.read_count += int64(count)
		if err == io.EOF || r.read_count >= r.max_read_count {
			r.active_read_state = false
			r.encoder.Close()
			if err == io.EOF {
				r.source_ended = true
			}
		}
	}
	count, err := r.buffer.Read(buff)
	r.encoding_count += int64(count)
	copy(p, buff[:count])
	return count, err
}

/*
	Source: source from data should be retrieved.
	Encoder: that encodes the data into a special format.
	EncodedSource: a buffer or pipe(non blocking) where encoder writes the data after encoding.
	MaxReadCount: limit for reading from a source (could be read almost nearly the max (can be greater than max by 10-1000 bytes)). 

	It should be safe to set "MaxReadCount" lower (atleast 500 bytes less) than the actual max amount needed 
*/ 

func NewRemoteReader(Source io.Reader, Encoder io.WriteCloser, EncodedSource io.Reader, MaxReadCount int64) RemoteReader {
	return &reader{Source, Encoder, EncodedSource, 0, 0, MaxReadCount, true, false}
}
