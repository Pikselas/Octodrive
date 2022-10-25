package ToOcto

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type MultiPartTransferer interface {
	ReadCount() int64
	TransferPart() (int, string, error)
}
type transferer struct {
	token               string
	baseurl             string
	readcount           int64
	chunksize           int64
	chunk_count         uint
	client              http.Client
	commiter            CommiterType
	source              io.Reader
	active_reader       EncodedReader
	active_cache_reader CachedReader
}

func (t *transferer) ReadCount() int64 {
	return t.readcount + t.active_cache_reader.ReadCount()
}

func (t *transferer) TransferPart() (status_code int, resp_string string, err error) {

	targetURL := t.baseurl + "/" + strconv.Itoa(int(t.chunk_count))

	if t.active_cache_reader.IsCached() {
		status_code, resp_string, err = Transfer(&t.client, targetURL, t.token, t.commiter, t.active_cache_reader)
		if status_code == 201 {
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
		status_code, resp_string, err = Transfer(&t.client, targetURL, t.token, t.commiter, t.active_cache_reader)
		if status_code != 201 {
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

func NewMultiPartTransferer(Commiter CommiterType, RepoUser string, Repo string, Path string, Token string, Source io.Reader) MultiPartTransferer {
	url := fmt.Sprintf(FILE_UPLOAD_URL+"/"+Path, RepoUser, Repo)
	reader := reader{}
	tra := http.Transport{}
	client := http.Client{Transport: &tra}
	return &transferer{Token, url, 0, 30000000, 0, client, Commiter, Source, &reader, &cachedReader{}}
}
