package RemoteToOcto

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
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
	commiter            CommiterType
	source              io.Reader
	active_reader       EncodedReader
	active_cache_reader CachedReader
}

func (t *transferer) ReadCount() int64 {
	return t.readcount + t.active_cache_reader.ReadCount()
}

func (t *transferer) TransferPart() (int, string, error) {

	targetURL := t.baseurl + "/" + strconv.Itoa(int(t.chunk_count))

	if t.active_cache_reader.IsCached() {
		stat, resp := Transfer(targetURL, t.token, t.commiter, t.active_cache_reader)
		if stat == 201 {
			t.active_cache_reader.Dispose()
		} else {
			t.active_cache_reader.ResetReadingState()
		}
		return stat, resp, nil
	}
	if !t.active_reader.SourceEnded() {
		b := bytes.Buffer{}
		enc := base64.NewEncoder(base64.StdEncoding, &b)
		t.active_reader = NewEncodedReader(t.source, enc, &b, t.chunksize)
		t.active_cache_reader = NewCachedReader(t.active_reader)
		stat, resp := Transfer(targetURL, t.token, t.commiter, t.active_cache_reader)
		if stat != 201 {
			t.active_cache_reader.ResetReadingState()
		} else {
			t.active_cache_reader.Dispose()
		}
		t.readcount += t.active_reader.ReadCount()
		t.chunk_count++
		return stat, resp, nil
	}
	return 0, "", io.EOF
}

func NewMultiPartTransferer(Commiter CommiterType, RepoUser string, Repo string, Path string, Token string, Source io.Reader) MultiPartTransferer {
	url := fmt.Sprintf(FILE_UPLOAD_URL+"/"+Path, RepoUser, Repo)
	reader := reader{}
	return &transferer{Token, url, 0, 30000000, 0, Commiter, Source, &reader, &cachedReader{}}
}
