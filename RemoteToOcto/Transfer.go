package RemoteToOcto

import (
	"io"
	"net/http"
)

func Transfer(client *http.Client, target string, token string, commiter CommiterType, body io.Reader) (int, string) {
	bodyformater := BodyFormater{0, body, commiter}
	GithubReq, err := http.NewRequest(http.MethodPut, target, &bodyformater)
	if err != nil {
		panic(err)
	}
	GithubReq.Header.Add("Accept", "application/vnd.github+json")
	GithubReq.Header.Add("Authorization", "Bearer "+token)
	GithubResp, err := client.Do(GithubReq)
	if err != nil {
		panic(err)
	}
	defer GithubResp.Body.Close()
	resp, _ := io.ReadAll(GithubResp.Body)
	return GithubResp.StatusCode, string(resp)
}
