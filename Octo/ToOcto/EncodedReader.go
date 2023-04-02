package ToOcto

import (
	"bytes"
	"encoding/base64"
	"io"
)

/*
Reads from source -> Encodes by the encoder and returns.
*/
type EncodedReader interface {
	Read(p []byte) (int, error)
}
type reader struct {
	source            io.Reader
	encoder           io.WriteCloser
	buffer            io.ReadWriter
	active_read_state bool
}

func (r *reader) Read(p []byte) (int, error) {
	if r.active_read_state {
		count, err := r.source.Read(p)
		r.encoder.Write(p[:count])
		if err == io.EOF {
			r.active_read_state = false
			r.encoder.Close()
		}
	}
	return r.buffer.Read(p)
}

/*
	Source: source from data should be retrieved.
	Encoder: that encodes the data into a special format.
*/

func NewEncodedReader(Source io.Reader) EncodedReader {
	buffer := bytes.Buffer{}
	Encoder := base64.NewEncoder(base64.StdEncoding, &buffer)
	return &reader{Source, Encoder, &buffer, true}
}
