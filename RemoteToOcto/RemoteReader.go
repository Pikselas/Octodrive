package RemoteToOcto

import (
	"io"
)

type RemoteReader struct {
	source            io.Reader
	encoder           io.WriteCloser
	buffer            io.ReadWriter
	read_count        *int64
	encoding_count    *int64
	max_read_count    int64
	active_read_state bool
}

func (r *RemoteReader) Read(p []byte) (int, error) {
	buff := make([]byte, 1000)
	if r.active_read_state {
		count, err := r.source.Read(buff)
		r.encoder.Write(buff[:count])
		*r.read_count += int64(count)
		if err == io.EOF || *r.read_count >= r.max_read_count {
			r.active_read_state = false
			r.encoder.Close()
		}
	}
	count, err := r.buffer.Read(buff)
	*r.encoding_count += int64(count)
	copy(p, buff[:count])
	return count, err
}
