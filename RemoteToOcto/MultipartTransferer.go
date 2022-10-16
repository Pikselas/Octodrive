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
	TransferMultiPart() int
}

type transferer struct {
	from          string
	repo          string
	path          string
	total         int64
	encoded       int64
	readcount     int64
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

func (t *transferer) TransferMultiPart() int {
	RemoteResp, err := http.Get(t.from)
	if err != nil {
		panic(err)
	}
	defer RemoteResp.Body.Close()
	t.total = RemoteResp.ContentLength
	token, mail, user := getUserData()
	count := 0
	for !t.active_reader.RemoteSourceEnded() {
		b := bytes.Buffer{}
		enc := base64.NewEncoder(base64.StdEncoding, &b)
		t.active_reader = NewRemoteReader(RemoteResp.Body, enc, &b, 20000000)
		targetURL := fmt.Sprintf(FILE_UPLOAD_URL+"/"+t.path+"/"+strconv.Itoa(count), user, t.repo)
		Transfer(targetURL, token, user, mail, t.active_reader)
		t.readcount += t.active_reader.ReadCount()
		t.encoded += t.active_reader.EncodeCount()
		count++
	}
	return 1
}

func NewMultiPartTransferer(From string, Repo string, Path string) MultiPartTransferer {
	return &transferer{From, Repo, Path, 0, 0, 0, &reader{}}
}
