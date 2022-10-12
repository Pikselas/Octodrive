package RemoteToOcto

import (
	"bytes"
	"io"
)

type RemoteReader struct {
	source            io.ReadCloser
	encoder           io.WriteCloser
	buffer            *bytes.Buffer
	read_count        *int64
	encoding_count    *int64
	active_read_state bool
}

func (r *RemoteReader) Read(p []byte) (int, error) {
	buff := make([]byte, 1000)
	if r.active_read_state {
		count, err := r.source.Read(buff)
		r.encoder.Write(buff[:count])
		if err == io.EOF {
			r.active_read_state = false
			r.encoder.Close()
		}
		*r.read_count += int64(count)
	}
	count, err := r.buffer.Read(buff)
	*r.encoding_count += int64(count)
	copy(p, buff[:count])
	return count, err
}
