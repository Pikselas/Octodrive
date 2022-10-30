package Octo

import (
	"Octo/Octo/ToOcto"
	"encoding/json"
	"io"
	"net/http"
)

type OctoUser interface {
	NewMultiPartTransferer(RepoUser string, Repo string, Path string, Source io.Reader) ToOcto.MultiPartTransferer
	NewMultipartReader(RepoUser string, Repo string, Path string) OctoMultiPartReader
}

type user struct {
	token    string
	commiter ToOcto.CommiterType
}

func (u *user) NewMultiPartTransferer(RepoUser string, Repo string, Path string, Source io.Reader) ToOcto.MultiPartTransferer {
	return ToOcto.NewMultiPartTransferer(u.commiter, RepoUser, Repo, Path, u.token, Source)
}

func (u *user) NewMultipartReader(RepoUser string, Repo string, Path string) OctoMultiPartReader {
	url := ToOcto.GetOctoURL(RepoUser, Repo, Path)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Accept", "application/vnd.github.v3.raw")
	req.Header.Add("Authorization", "Bearer "+u.token)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	var jArr []interface{}
	json.NewDecoder(res.Body).Decode(&jArr)
	return NewMultipartReader(url, len(jArr), u.token)
}

func NewOctoUser(User string, Mail string, Token string) OctoUser {
	return &user{Token, ToOcto.CommiterType{Name: User, Email: Mail}}
}
