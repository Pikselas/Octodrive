package ToOcto

import (
	"io"
	"net/http"
)

func Transfer(client *http.Client, target string, token string, commiter CommiterType, body io.Reader) (resp_code int, resp_string string, err error) {
	bodyformater := BodyFormater{0, body, commiter}
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
