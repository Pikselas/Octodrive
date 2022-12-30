package Octo

import (
	"Octo/Octo/ToOcto"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OctoUser interface {
	NewMultiPartTransferer(RepoUser string, Repo string, Path string, Source io.Reader) ToOcto.MultiPartTransferer
	NewMultipartReader(RepoUser string, Repo string, Path string) (OctoMultiPartReader, error)
}

type user struct {
	token    string
	commiter ToOcto.CommiterType
}

func (u *user) createRepository(name string, description string) (int, error) {
	data := bytes.NewBufferString(fmt.Sprintf(`{"name": "%s",
	"description": "%s",
	"homepage": "https://github.com",
	"private": true}`, name, description))
	rq, err := http.NewRequest(http.MethodPost, "https://api.github.com/user/repos", data)
	if err != nil {
		return 0, err
	}
	rq.Header.Add("Authorization", "token "+u.token)
	res, err := http.DefaultClient.Do(rq)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	return res.StatusCode, nil
}

func (u *user) NewMultiPartTransferer(RepoUser string, Repo string, Path string, Source io.Reader) ToOcto.MultiPartTransferer {
	return ToOcto.NewMultiPartTransferer(u.commiter, RepoUser, Repo, Path, u.token, Source)
}

func (u *user) NewMultipartReader(RepoUser string, Repo string, Path string) (OctoMultiPartReader, error) {
	url := ToOcto.GetOctoURL(RepoUser, Repo, Path)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Accept", "application/vnd.github.v3.raw")
	req.Header.Add("Authorization", "Bearer "+u.token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var jArr []interface{}
	json.NewDecoder(res.Body).Decode(&jArr)
	return NewMultipartReader(url, len(jArr), u.token), nil
}

func NewOctoUser(User string, Mail string, Token string) (OctoUser, error) {
	U := user{Token, ToOcto.CommiterType{Name: User, Email: Mail}}
	_, err := U.createRepository("_Octofiles", "Initial repo for OctoDrive contents")
	return &U, err
}
