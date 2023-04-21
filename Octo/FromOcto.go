package Octo

import (
	"Octo/Octo/ToOcto"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type Decrypter interface {
	Decrypt(io.Reader) (io.Reader, error)
}

type OctoMultiPartReader interface {
	GetReadCount() uint64
	Read(p []byte) (n int, err error)
	Close() error
}

type reader struct {
	repo               string
	path               string
	user               *ToOcto.OctoUser
	max_count          int
	current_count      int
	current_read_count int
	read_count         uint64
	current_source     io.Reader
	decrypter          Decrypter
	source_closer      io.Closer
}

func (r *reader) GetReadCount() uint64 {
	return r.read_count
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.current_count >= r.max_count {
		return 0, io.EOF
	}
	if r.current_source == nil {
		var Err *ToOcto.Error
		r.current_source, Err = r.user.GetContent(r.repo, r.path+"/"+strconv.Itoa(r.current_count))
		if Err != nil {
			return 0, Err
		}
		if r.decrypter != nil {
			r.current_source, err = r.decrypter.Decrypt(r.current_source)
			if err != nil {
				return 0, err
			}
		}
	}
	n, err = r.current_source.Read(p)
	r.current_read_count += n
	if err == io.EOF {
		r.Close()
		r.source_closer = nil
		r.current_count++
		r.current_read_count = 0
		r.current_source = nil
		return n, nil
	} else if err != nil {
		r.current_source = nil
		return
	}
	return
}

func (r *reader) Close() error {
	if r.current_source != nil && r.source_closer != nil {
		return r.source_closer.Close()
	}
	return nil
}

func getPartCount(User *ToOcto.OctoUser, Repo string, Path string) (uint, error) {
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

func NewMultipartReader(User *ToOcto.OctoUser, Repo string, Path string, part_count int, dec Decrypter) OctoMultiPartReader {
	r := new(reader)
	*r = reader{
		repo:      Repo,
		path:      Path,
		user:      User,
		max_count: part_count,
		decrypter: dec,
	}
	return r
}

func NewMultipartRangeReader(User *ToOcto.OctoUser, Repo string, Path string, part_start int, part_end int, dec Decrypter) OctoMultiPartReader {
	r := new(reader)
	*r = reader{
		repo:          Repo,
		path:          Path,
		user:          User,
		current_count: part_start,
		max_count:     part_end,
		decrypter:     dec,
	}
	return r
}
