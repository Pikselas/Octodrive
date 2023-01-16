package ToOcto

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func transfer(client *http.Client, target string, token string, commiter CommiterType, body io.Reader, sha *string) (resp_code int, resp_string string, err error) {
	bodyformater := BodyFormater{0, body, commiter, sha}
	GithubReq, err := http.NewRequest(http.MethodPut, target, &bodyformater)
	if err != nil {
		return
	}
	GithubReq.Header.Add("Accept", "application/vnd.github+json")
	GithubReq.Header.Add("Authorization", "Bearer "+token)
	GithubResp, err := client.Do(GithubReq)
	if err != nil {
		return
	}
	defer GithubResp.Body.Close()
	resp_byte, err := io.ReadAll(GithubResp.Body)
	return GithubResp.StatusCode, string(resp_byte), err
}

func Transfer(client *http.Client, target string, token string, commiter CommiterType, body io.Reader) (resp_code int, resp_string string, err error) {
	return transfer(client, target, token, commiter, body, nil)
}

// Updating File is expnsive , Should only be done for small files
func Update(target string, token string, commiter CommiterType, data io.Reader) (resp_code int, resp_string string, err error) {
	//get the sha of the file
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", "Bearer "+token)
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
		return transfer(http.DefaultClient, target, token, commiter, data, &SHA)
	}
	return 0, "", errors.New("no SHA found")
}
