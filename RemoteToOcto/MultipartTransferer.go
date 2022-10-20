package RemoteToOcto

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
)

type MultiPartTransferer interface {
	EncodedSize() int64
	ReadCount() int64
	RawTransferSize() int64
	TransferMultiPart(From string, Repo string, Path string) int
}

type transferer struct {
	token         string
	total         int64
	encoded       int64
	readcount     int64
	chunksize     int64
	commiter      CommiterType
	active_reader RemoteReader
}

func (t *transferer) EncodedSize() int64 {
	return t.encoded + t.active_reader.EncodeCount()
}

func (t *transferer) ReadCount() int64 {
	return t.readcount + t.active_reader.ReadCount()
}

func (t *transferer) RawTransferSize() int64 {
	return t.total
}

func (t *transferer) TransferMultiPart(From string, Repo string, Path string) int {
	RemoteResp, err := http.Get(From)
	if err != nil {
		panic(err)
	}
	defer RemoteResp.Body.Close()
	t.total = RemoteResp.ContentLength
	count := 0
	for !t.active_reader.RemoteSourceEnded() {
		b := bytes.Buffer{}
		enc := base64.NewEncoder(base64.StdEncoding, &b)
		t.active_reader = NewRemoteReader(RemoteResp.Body, enc, &b, t.chunksize)
		targetURL := fmt.Sprintf(FILE_UPLOAD_URL+"/"+Path+"/"+strconv.Itoa(count), t.commiter.Name, Repo)
		cacheR := NewCachedReader(t.active_reader)
		for Transfer(targetURL, t.token, t.commiter, cacheR) != 201 {
			fmt.Println("Cache Reading")
			cacheR.ResetReadingState()
		}
		cacheR.Dispose()
		t.readcount += t.active_reader.ReadCount()
		t.encoded += t.active_reader.EncodeCount()
		count++
	}
	return 1
}

func NewMultiPartTransferer(Commiter CommiterType, Token string) MultiPartTransferer {
	return &transferer{Token, 0, 0, 0, 30000000,Commiter,&reader{}}
}
