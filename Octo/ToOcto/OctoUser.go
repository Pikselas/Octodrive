package ToOcto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type OctoUser interface {
	CreateRepository(name string, description string) (int, error)
	MakeRequest(method string, repo string, path string, body io.Reader, is_raw bool) (*http.Request, error)
	GetContent(repo string, path string) (io.ReadCloser, error)
	Transfer(repo string, path string, body io.Reader) (resp_code int, resp_string string, err error)
	Update(repo string, path string, data io.Reader) (resp_code int, resp_string string, err error)
}

type octoUser struct {
	name   string
	email  string
	token  string
	client *http.Client
}

func (u *octoUser) MakeRequest(method string, repo string, path string, body io.Reader, is_raw bool) (*http.Request, error) {
	rq, err := http.NewRequest(method, GetOctoURL(u.name, repo, path), body)
	if err != nil {
		return nil, err
	}
	rq.Header.Add("Authorization", "bearer "+u.token)
	if is_raw {
		rq.Header.Add("Accept", "application/vnd.github.v3.raw")
	} else {
		rq.Header.Add("Accept", "application/vnd.github.v3+json")
	}
	return rq, err
}

func (u *octoUser) CreateRepository(name string, description string) (int, error) {
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

func (u *octoUser) transfer(target string, body io.Reader, sha *string) (resp_code int, resp_string string, err error) {
	body_formatter := BodyFormatter{reader: body, sha: sha, name: u.name, email: u.email}
	GithubReq, err := http.NewRequest(http.MethodPut, target, &body_formatter)
	if err != nil {
		return
	}
	GithubReq.Header.Add("Accept", "application/vnd.github+json")
	GithubReq.Header.Add("Authorization", "Bearer "+u.token)
	GithubResp, err := u.client.Do(GithubReq)
	if err != nil {
		return
	}
	defer GithubResp.Body.Close()
	resp_byte, err := io.ReadAll(GithubResp.Body)
	return GithubResp.StatusCode, string(resp_byte), err
}

func (u *octoUser) Transfer(repo string, path string, body io.Reader) (resp_code int, resp_string string, err error) {
	return u.transfer(GetOctoURL(u.name, repo, path), body, nil)
}

// Updating File is expensive , Should only be done for small files
func (u *octoUser) Update(repo string, path string, data io.Reader) (resp_code int, resp_string string, err error) {
	target := GetOctoURL(u.name, repo, path)
	//get the sha of the file
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", "Bearer "+u.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var JsonData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&JsonData)
	if err != nil {
		return
	}
	sha, ok := JsonData["sha"]
	if ok {
		SHA := sha.(string)
		return u.transfer(target, data, &SHA)
	}
	return 0, "", errors.New("no SHA found")
}

func (u *octoUser) GetContent(repo string, path string) (io.ReadCloser, error) {
	rq, err := u.MakeRequest(http.MethodGet, repo, path, nil, true)
	if err != nil {
		return nil, err
	}
	res, err := u.client.Do(rq)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func NewOctoUser(name string, email string, token string) (OctoUser, error) {
	return &octoUser{name: name, email: email, token: token, client: &http.Client{}}, nil
}
