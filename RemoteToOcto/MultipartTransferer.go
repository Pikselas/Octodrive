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
	to        string
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
	token, mail, user, repo := getUserData()
	b := bytes.Buffer{}
	enc := base64.NewEncoder(base64.StdEncoding, &b)
	reader := RemoteReader{RemoteResp.Body, enc, &b, &t.readcount, &t.encoded, true}
	r := BodyFormater{0, &reader, CommiterType{user, mail}}
	targetURL := fmt.Sprintf(FILE_UPLOAD_URL+"/"+t.to, user, repo)
	GithubReq, err := http.NewRequest(http.MethodPut, targetURL, &r)
	if err != nil {
		panic(err)
	}
	GithubReq.Header.Add("Accept", "application/vnd.github+json")
	GithubReq.Header.Add("Authorization", "Bearer "+token)
	GithubResp, err := http.DefaultClient.Do(GithubReq)
	if err != nil {
		panic(err)
	}
	GithubResp.Body.Close()
	return GithubResp.StatusCode
}

func NewMultiPartTransferer(From string, To string) MultiPartTransferer {
	return &transferer{From, To, 0, 0, 0}
}
