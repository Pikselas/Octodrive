package main

import (
	"io"
	"net/http"
	"strconv"
)

type OctoMultiPartReader interface {
	GetReadCount() int64
	Read(p []byte) (n int, err error)
}

type reader struct {
	from           string
	max_count      int
	current_count  int
	read_count     int64
	client         http.Client
	current_source io.ReadCloser
}

func (r *reader) GetReadCount() int64 {
	return r.read_count
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.current_count > r.max_count {
		return 0, io.EOF
	}
	if r.current_source == nil {
		rq, err := r.client.Get(r.from + "/" + strconv.Itoa(r.current_count))
		if err != nil {
			panic(err)
		}
		r.current_source = rq.Body
	}
	n, err = r.current_source.Read(p)
	if err == io.EOF {
		r.current_source.Close()
		r.current_count++
		r.current_source = nil
		if n > 0 {
			return n, nil
		}
		return r.Read(p)
	}
	return
}

func NewOctoMultipartReader(from string, part_count int) OctoMultiPartReader {
	return &reader{
		from:      from,
		max_count: part_count,
	}
}
