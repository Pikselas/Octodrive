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

const (
	MaxOctoRepoSize = 100000000
)

type OctoUser interface {
	CreateFile(path string, source io.Reader) error
	LoadFile(path string, dest io.Writer) error
}

type user struct {
	token    string
	commiter ToOcto.CommiterType
}

func (u *user) makeRequest(method string, repo string, path string, body io.Reader, is_raw bool) (*http.Response, error) {
	rq, err := http.NewRequest(method, ToOcto.GetOctoURL(u.commiter.Name, repo, path), body)
	if err != nil {
		return nil, err
	}
	rq.Header.Add("Authorization", "token "+u.token)
	if is_raw {
		rq.Header.Add("Accept", "application/vnd.github.v3.raw")
	} else {
		rq.Header.Add("Accept", "application/vnd.github.v3+json")
	}
	return http.DefaultClient.Do(rq)
}

func (u *user) createRepository(name string, description string) (int, error) {
	data := bytes.NewBufferString(fmt.Sprintf(`{"name": "%s",
	"description": "%s",
	"homepage": null,
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
	var Repository string
	//make a new repository
	Repository = RandomString(10)
	status, err := u.createRepository(Repository, "Repository for OctoDrive contents")
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("error creating repository: %d", status)
	}
	fileID := RandomString(10)
	paths := make([]string, 0)
	paths = append(paths, Repository)
	//create a multipart transferer with source limiter for max repository size
	var reader SourceLimiter
	var ContentSize int64 = 0
	for {
		reader = NewSourceLimiter(source, MaxOctoRepoSize)
		//create a new multipart transferer
		transferer := u.NewMultiPartTransferer(Repository, fileID, reader)
		err = nil
		for err != io.EOF {
			_, _, err = transferer.TransferPart()
			fmt.Println(err)
		}
		ContentSize += reader.GetCurrentSize()
		if !reader.IsEOF() {
			print("Creating new repository")
			newRepo := RandomString(10)
			status, err := u.createRepository(newRepo, "Repository for OctoDrive contents")
			if err != nil {
				return err
			}
			if status != http.StatusCreated {
				return fmt.Errorf("error creating repository: %d", status)
			}
			paths = append(paths, newRepo)
			Repository = newRepo
		} else {
			break
		}
	}
	//create file details
	var FileDetails fileDetails
	FileDetails.Name = fileID
	FileDetails.Paths = paths
	FileDetails.Size = ContentSize
	data, err := json.Marshal(FileDetails)
	if err != nil {
		return err
	}
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	_, Str, err := ToOcto.Transfer(http.DefaultClient, ToOcto.GetOctoURL(u.commiter.Name, "_Octofiles", "Contents/"+path), u.token, u.commiter, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	println(Str)
	return nil
}

func (u *user) LoadFile(path string, w io.Writer) error {
	//get file details
	res, err := u.makeRequest(http.MethodGet, "_Octofiles", "Contents/"+path, nil, true)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var FileDetails fileDetails
	json.NewDecoder(res.Body).Decode(&FileDetails)
	//get file parts
	for _, repo := range FileDetails.Paths {
		r, err := u.NewMultiPartFileReader(repo, FileDetails.Name)
		if err != nil {
			return err
		}
		io.Copy(w, r)
	}
	return nil
}

func (u *user) NewMultiPartTransferer(Repo string, Path string, Source io.Reader) ToOcto.MultiPartTransferer {
	return ToOcto.NewMultiPartTransferer(u.commiter, Repo, Path, u.token, Source)
}

func (u *user) NewMultiPartFileReader(Repo string, Path string) (OctoMultiPartReader, error) {
	url := ToOcto.GetOctoURL(u.commiter.Name, Repo, Path)
	fmt.Println(url)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Accept", "application/vnd.github.v3.raw")
	req.Header.Add("Authorization", "Bearer "+u.token)
	res, err := http.DefaultClient.Do(req)
	fmt.Println(res.StatusCode)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var jArr []interface{}
	json.NewDecoder(res.Body).Decode(&jArr)
	fmt.Println(jArr)
	return NewMultipartReader(url, len(jArr), u.token), nil
}

func NewOctoUser(User string, Mail string, Token string) (OctoUser, error) {
	U := user{Token, ToOcto.CommiterType{Name: User, Email: Mail}}
	status, err := U.createRepository("_Octofiles", "Initial repo for OctoDrive contents")
	if err != nil {
		return nil, err
	}
	if status != http.StatusCreated && status != http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("error creating repository: %d", status)
	}
	return &U, err
}
