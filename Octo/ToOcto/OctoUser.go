package ToOcto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

/*
 Interface that wraps the basic methods of a user.
*/

type OctoUser struct {
	name   string
	email  string
	token  string
	client *http.Client
}

// Returns *http.Request for the given method, repo and path.
// If is_raw is true, the request is made for raw data.
func (u *OctoUser) MakeRequest(method string, repo string, path string, body io.Reader, is_raw bool) (*http.Request, error) {
	rq, err := http.NewRequest(method, getOctoURL(u.name, repo, path), body)
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

// Creates a new repository with the given name and description.
func (u *OctoUser) CreateRepository(name string, description string) *Error {
	data := bytes.NewBufferString(fmt.Sprintf(`{"name": "%s",
	"description": "%s",
	"homepage": null,
	"private": true}`, name, description))
	rq, err := http.NewRequest(http.MethodPost, "https://api.github.com/user/repos", data)
	if err != nil {
		return NewError(ErrorCreatingRepository, 0, nil, err)
	}
	rq.Header.Add("Authorization", "token "+u.token)
	res, err := http.DefaultClient.Do(rq)
	if err != nil {
		return NewError(ErrorCreatingRepository, 0, nil, err)
	}
	if res.StatusCode != http.StatusCreated {
		return NewError(ErrorCreatingRepository, res.StatusCode, res.Body, nil)
	}
	res.Body.Close()
	return nil
}

// transfers data to the target path.
// If sha is not nil, the data is transferred to the given sha.
func (u *OctoUser) transfer(target string, body io.Reader, sha *string) *Error {
	b64reader := NewEncodedReader(body)
	body_formatter := BodyFormatter{reader: b64reader, sha: sha, name: u.name, email: u.email}
	GithubReq, err := http.NewRequest(http.MethodPut, target, &body_formatter)
	if err != nil {
		return NewError(ErrorTransferring, 0, nil, err)
	}
	GithubReq.Header.Add("Accept", "application/vnd.github+json")
	GithubReq.Header.Add("Authorization", "Bearer "+u.token)
	GithubResp, err := u.client.Do(GithubReq)
	if err != nil {
		return NewError(ErrorTransferring, 0, nil, err)
	}
	if GithubResp.StatusCode != http.StatusCreated && GithubResp.StatusCode != http.StatusOK {
		return NewError(ErrorTransferring, GithubResp.StatusCode, GithubResp.Body, nil)
	}
	GithubResp.Body.Close()
	return nil
}

// Creates a new file in the given repository.
func (u *OctoUser) Transfer(repo string, path string, body io.Reader) *Error {
	return u.transfer(getOctoURL(u.name, repo, path), body, nil)
}

// Updates the file in the given repository.
// Updating File is expensive , Should only be done for small files
func (u *OctoUser) Update(repo string, path string, data io.Reader) *Error {
	target := getOctoURL(u.name, repo, path)
	//get the sha of the file
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return NewError(ErrorUpdating, 0, nil, err)
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", "Bearer "+u.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NewError(ErrorUpdating, 0, nil, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return NewError(ErrorUpdating, resp.StatusCode, resp.Body, nil)
	}
	var JsonData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&JsonData)
	if err != nil {
		return NewError(ErrorUpdating, 0, nil, err)
	}
	sha, ok := JsonData["sha"]
	if ok {
		SHA := sha.(string)
		return u.transfer(target, data, &SHA)
	}
	return NewError(ErrorUpdating, 0, nil, errors.New("SHA not found"))
}

// Gets the content of the file from the given repository.
func (u *OctoUser) GetContent(repo string, path string) (io.ReadCloser, *Error) {
	rq, err := u.MakeRequest(http.MethodGet, repo, path, nil, true)
	if err != nil {
		return nil, NewError(ErrorGettingContent, 0, nil, err)
	}
	res, err := u.client.Do(rq)
	if err != nil {
		return nil, NewError(ErrorGettingContent, 0, nil, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, NewError(ErrorGettingContent, res.StatusCode, res.Body, nil)
	}
	return res.Body, nil
}

// Creates a new user with the given name, email and token.
func NewOctoUser(name string, email string, token string) *OctoUser {
	user := new(OctoUser)
	*user = OctoUser{name: name, email: email, token: token, client: &http.Client{}}
	return user
}
