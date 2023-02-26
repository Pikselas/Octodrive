package Octo

import (
	"Octo/Octo/ToOcto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type OctoMultiPartReader interface {
	GetReadCount() int64
	Read(p []byte) (n int, err error)
}

type reader struct {
	from               string
	token              string
	max_count          int
	current_count      int
	current_read_count int
	read_count         int64
	client             http.Client
	current_source     io.ReadCloser
}

func (r *reader) GetReadCount() int64 {
	return r.read_count
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.current_count >= r.max_count {
		return 0, io.EOF
	}
	if r.current_source == nil {
		req, err := http.NewRequest(http.MethodGet, r.from+"/"+strconv.Itoa(r.current_count), nil)
		if err != nil {
			return 0, err
		}
		req.Header.Add("Accept", "application/vnd.github.v3.raw")
		req.Header.Add("Authorization", "Bearer "+r.token)
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-", r.current_read_count))
		res, err := r.client.Do(req)
		if err != nil {
			panic(err)
		}
		fmt.Println(res.Status)
		r.current_source = res.Body
	}
	n, err = r.current_source.Read(p)
	r.current_read_count += n
	if err == io.EOF {
		r.current_source.Close()
		r.current_count++
		r.current_read_count = 0
		r.current_source = nil
		if n > 0 {
			return n, nil
		}
		n, err = r.Read(p)
		r.current_read_count = n
		return
	} else if err != nil {
		r.current_source = nil
		return
	}
	return
}

func getPartCount(User string, Token string, Repo string, Path string) (uint, error) {
	url := ToOcto.GetOctoURL(User, Repo, Path)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Accept", "application/vnd.github.v3.raw")
	req.Header.Add("Authorization", "Bearer "+Token)
	res, err := http.DefaultClient.Do(req)
	fmt.Println(res.StatusCode)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	var jArr []interface{}
	json.NewDecoder(res.Body).Decode(&jArr)
	return uint(len(jArr)), nil
}

func NewMultipartReader(from string, part_count int, token string) OctoMultiPartReader {
	return &reader{
		from:      from,
		max_count: part_count,
		token:     token,
	}
}

func NewMultipartRangeReader(from string, part_start int, part_end int, token string) OctoMultiPartReader {
	return &reader{
		from:          from,
		current_count: part_start,
		max_count:     part_end,
		token:         token,
	}
}
