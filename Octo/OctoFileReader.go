package Octo

import (
	"fmt"
	"io"
	"net/http"
)

type delayedReader struct {
	req         *http.Request
	src         io.ReadCloser
	ignoreBytes uint64
}

func (r *delayedReader) Read(p []byte) (n int, err error) {
	if r.src == nil {
		res, err := http.DefaultClient.Do(r.req)
		if err != nil {
			return 0, err
		}
		r.src = res.Body
		n, err := io.CopyN(io.Discard, r.src, int64(r.ignoreBytes))
		if err != nil {
			return 0, err
		}
		fmt.Println("Ignored:", n)
	}
	return r.src.Read(p)
}

func (r *delayedReader) Close() error {
	if r.src != nil {
		r.src.Close()
		r.src = nil
	}
	return nil
}

type remoteReader struct {
	req *http.Request
	src io.ReadCloser
}

func (r *remoteReader) Read(p []byte) (n int, err error) {
	if r.src == nil {
		res, err := http.DefaultClient.Do(r.req)
		if err != nil {
			return 0, err
		}
		r.src = res.Body
	}
	c, err := r.src.Read(p)
	return c, err
}

func (r *remoteReader) Close() error {
	if r.src != nil {
		r.src.Close()
		r.src = nil
	}
	return nil
}

type octoFileReader struct {
	readers            []io.ReadCloser
	current_read_index uint
	read_end           bool
}

func (r *octoFileReader) Read(p []byte) (n int, err error) {
	if r.read_end && r.current_read_index < uint(len(r.readers)) {
		r.read_end = false
	}
	if !r.read_end {
		n, err := r.readers[r.current_read_index].Read(p)
		if err == io.EOF {
			r.readers[r.current_read_index].Close()
			r.read_end = true
			r.current_read_index++
		} else if err != nil {
			return n, err
		}
		return n, nil
	}
	return 0, io.EOF
}

func (r *octoFileReader) Close() error {
	if r.current_read_index < uint(len(r.readers)) {
		r.readers[r.current_read_index].Close()
		r.current_read_index = uint(len(r.readers))
	}
	return nil
}
