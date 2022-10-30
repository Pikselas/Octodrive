package Octo

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
	token          string
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
		req, err := http.NewRequest(http.MethodGet, r.from+"/"+strconv.Itoa(r.current_count), nil)
		if err != nil {
			return 0, err
		}
		req.Header.Add("Accept", "application/vnd.github.v3.raw")
		req.Header.Add("Authorization", "Bearer "+r.token)
		res, err := r.client.Do(req)
		if err != nil {
			panic(err)
		}
		r.current_source = res.Body
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

func NewMultipartReader(from string, part_count int, token string) OctoMultiPartReader {
	return &reader{
		from:      from,
		max_count: part_count,
		token:     token,
	}
}
