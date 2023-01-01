package Octo

import (
	"Octo/Octo/ToOcto"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OctoUser interface {
	CreateFile(path string, source io.Reader) error
	NewMultiPartTransferer(Repo string, Path string, Source io.Reader) ToOcto.MultiPartTransferer
	NewMultipartReader(RepoUser string, Repo string, Path string) (OctoMultiPartReader, error)
}

type user struct {
	token    string
	commiter ToOcto.CommiterType
}

func (u *user) makeRequest(method string, repo string, path string, body io.Reader) (*http.Response, error) {
	rq, err := http.NewRequest(method, ToOcto.GetOctoURL(u.commiter.Name, repo, path), body)
	if err != nil {
		return nil, err
	}
	rq.Header.Add("Authorization", "token "+u.token)
	rq.Header.Add("Accept", "application/vnd.github.v3.raw")
	return http.DefaultClient.Do(rq)
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

func (u *user) CreateFile(path string, source io.Reader) error {
	//get the last repository use
	res, err := u.makeRequest(http.MethodGet, "_Octofiles", "LastRepo.json", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error getting last repository: %s", res.Status)
	}
	var lastRepo struct {
		Repo  *string `field:"name"`
		Usage int64   `field:"used"`
	}
	err = json.NewDecoder(res.Body).Decode(&lastRepo)
	if err != nil {
		return err
	}
	//make a new repository if it does not exists or if the last one is full
	if lastRepo.Repo == nil || lastRepo.Usage > 1000000000 {
		lastRepo.Repo = new(string)
		*lastRepo.Repo = RandomString(10)
		lastRepo.Usage = 0
		status, err := u.createRepository(*lastRepo.Repo, "Repository for OctoDrive contents")
		if err != nil {
			return err
		}
		if status != http.StatusCreated {
			return fmt.Errorf("error creating repository: %s", res.Status)
		}
	}
	//create a multipart transferer with source limiter for max repository size
	reader := SourceLimiter{Source: source, MaxSize: 1000000000 - lastRepo.Usage}
	for !reader.IsEOF() {
		//create a new multipart transferer
		//transferer := u.NewMultiPartTransferer(*lastRepo.Repo, path, &reader)
	}
	//transfer the file
	//if the file is too big, create a new repository and transfer the file
	//path = ToOcto.GetOctoURL(u.commiter.Name, "_Octofiles/Folders", path)
	return nil
}

func (u *user) NewMultiPartTransferer(Repo string, Path string, Source io.Reader) ToOcto.MultiPartTransferer {
	return ToOcto.NewMultiPartTransferer(u.commiter, Repo, Path, u.token, Source)
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
	staus, err := U.createRepository("_Octofiles", "Initial repo for OctoDrive contents")
	if err != nil {
		return nil, err
	}
	if staus == 201 {
		buf := bytes.NewBuffer([]byte(`{"name": null, "used": 0}`))
		b := bytes.Buffer{}
		enc := base64.NewEncoder(base64.StdEncoding, &b)
		encreadr := ToOcto.NewEncodedReader(buf, enc, &b, 100)
		resp, _, err := ToOcto.Transfer(http.DefaultClient, ToOcto.GetOctoURL(User, "_Octofiles", "LastRepo.json"), Token, U.commiter, encreadr)
		if err != nil {
			return nil, err
		}
		if resp != 201 {
			return nil, fmt.Errorf("error creating LastRepo.json")
		}
	}
	return &U, err
}
