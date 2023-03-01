package ToOcto

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
)

type MultiPartTransferer interface {
	ReadCount() int64
	TransferPart() (int, string, error)
}
type transferer struct {
	repo                string
	path                string
	readcount           int64
	chunksize           int64
	chunk_count         uint
	user                OctoUser
	source              io.Reader
	active_reader       EncodedReader
	active_cache_reader CachedReader
}

func (t *transferer) ReadCount() int64 {
	return t.readcount + t.active_reader.ReadCount()
}

func (t *transferer) TransferPart() (status_code int, resp_string string, err error) {

	if t.active_cache_reader.IsCached() {
		status_code, resp_string, err = t.user.Transfer(t.repo, t.path+"/"+strconv.Itoa(int(t.chunk_count)), t.active_cache_reader)
		if status_code == http.StatusCreated {
			t.active_cache_reader.Dispose()
		} else {
			t.active_cache_reader.ResetReadingState()
		}
		return
	}
	if !t.active_reader.SourceEnded() {
		b := bytes.Buffer{}
		enc := base64.NewEncoder(base64.StdEncoding, &b)
		t.active_reader = NewEncodedReader(t.source, enc, &b, t.chunksize)
		t.active_cache_reader = NewCachedReader(t.active_reader)
		status_code, resp_string, err = t.user.Transfer(t.repo, t.path+"/"+strconv.Itoa(int(t.chunk_count)), t.active_cache_reader)
		if status_code != http.StatusCreated {
			t.active_cache_reader.ResetReadingState()
		} else {
			t.active_cache_reader.Dispose()
		}
		t.readcount += t.active_reader.ReadCount()
		t.chunk_count++
		return
	}
	return 0, "", io.EOF
}

func NewMultiPartTransferer(User OctoUser, Repo string, Path string, chunksize int64, Source io.Reader) MultiPartTransferer {
	return &transferer{
		repo:                Repo,
		path:                Path,
		user:                User,
		chunksize:           chunksize,
		source:              Source,
		active_reader:       &reader{},
		active_cache_reader: &cachedReader{},
	}
}
