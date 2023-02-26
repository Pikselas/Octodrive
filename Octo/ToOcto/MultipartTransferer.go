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
	token               string
	baseurl             string
	readcount           int64
	chunksize           int64
	chunk_count         uint
	targetURL           string
	client              http.Client
	commiter            CommiterType
	source              io.Reader
	active_reader       EncodedReader
	active_cache_reader CachedReader
}

func (t *transferer) ReadCount() int64 {
	return t.readcount + t.active_reader.ReadCount()
}

func (t *transferer) TransferPart() (status_code int, resp_string string, err error) {

	if t.active_cache_reader.IsCached() {
		status_code, resp_string, err = Transfer(&t.client, t.targetURL, t.token, t.commiter, t.active_cache_reader)
		if status_code == 201 {
			t.active_cache_reader.Dispose()
		} else {
			t.active_cache_reader.ResetReadingState()
		}
		return
	}
	if !t.active_reader.SourceEnded() {
		t.targetURL = t.baseurl + "/" + strconv.Itoa(int(t.chunk_count))
		b := bytes.Buffer{}
		enc := base64.NewEncoder(base64.StdEncoding, &b)
		t.active_reader = NewEncodedReader(t.source, enc, &b, t.chunksize)
		t.active_cache_reader = NewCachedReader(t.active_reader)
		status_code, resp_string, err = Transfer(&t.client, t.targetURL, t.token, t.commiter, t.active_cache_reader)
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

func NewMultiPartTransferer(Commiter CommiterType, Repo string, Path string, Token string, chunksize int64, Source io.Reader) MultiPartTransferer {
	return &transferer{
		token:               Token,
		baseurl:             GetOctoURL(Commiter.Name, Repo, Path),
		chunksize:           chunksize,
		commiter:            Commiter,
		source:              Source,
		active_reader:       &reader{},
		active_cache_reader: &cachedReader{},
	}
}
