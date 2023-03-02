package Octo

import (
	"Octo/Octo/ToOcto"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type OctoMultiPartReader interface {
	GetReadCount() uint64
	Read(p []byte) (n int, err error)
}

type reader struct {
	repo               string
	path               string
	user               ToOcto.OctoUser
	max_count          int
	current_count      int
	current_read_count int
	read_count         uint64
	current_source     io.ReadCloser
}

func (r *reader) GetReadCount() uint64 {
	return r.read_count
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.current_count >= r.max_count {
		return 0, io.EOF
	}
	if r.current_source == nil {
		r.current_source, err = r.user.GetContent(r.repo, r.path+"/"+strconv.Itoa(r.current_count))
		if err != nil {
			return 0, err
		}
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

func getPartCount(User ToOcto.OctoUser, Repo string, Path string) (uint, error) {
	req, err := User.MakeRequest(http.MethodGet, Repo, Path, nil, false)
	if err != nil {
		return 0, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	var jArr []interface{}
	json.NewDecoder(res.Body).Decode(&jArr)
	return uint(len(jArr)), nil
}

func NewMultipartReader(User ToOcto.OctoUser, Repo string, Path string, part_count int) OctoMultiPartReader {
	return &reader{
		repo:      Repo,
		path:      Path,
		user:      User,
		max_count: part_count,
	}
}

func NewMultipartRangeReader(User ToOcto.OctoUser, Repo string, Path string, part_start int, part_end int) OctoMultiPartReader {
	return &reader{
		repo:          Repo,
		path:          Path,
		user:          User,
		current_count: part_start,
		max_count:     part_end,
	}
}
