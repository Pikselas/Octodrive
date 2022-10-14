package RemoteToOcto

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
)

type MultiPartTransferer interface {
	EncodedSize() int64
	ReadCount() int64
	RawTransferSize() int64
	TransferMultiPart() int
}

type transferer struct {
	from      string
	repo      string
	path      string
	encoded   int64
	total     int64
	readcount int64
}

func (t *transferer) EncodedSize() int64 {
	return t.encoded
}

func (t *transferer) ReadCount() int64 {
	return t.readcount
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
	b := bytes.Buffer{}
	enc := base64.NewEncoder(base64.StdEncoding, &b)
	reader := RemoteReader{RemoteResp.Body, enc, &b, &t.readcount, &t.encoded, 50000000, true, false}
	targetURL := fmt.Sprintf(FILE_UPLOAD_URL+"/"+t.path, user, t.repo)
	return Transfer(targetURL, token, user, mail, &reader)
}

func NewMultiPartTransferer(From string, Repo string, Path string) MultiPartTransferer {
	return &transferer{From, Repo, Path, 0, 0, 0}
}
