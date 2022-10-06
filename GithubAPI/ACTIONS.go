package GithubAPI

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func getUserData() (string, string, string, string) {
	type User struct {
		Token string
		Mail  string
		User  string
		Repo  string
	}
	jsonFile, err := os.Open("ENV_KEY.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	var user User
	json.Unmarshal(byteValue, &user)
	return os.Getenv(user.Token), os.Getenv(user.Mail), user.User, user.Repo
}

func Transfer(From string, To string) int{
	RemoteResp, err := http.Get(From)
	if err != nil {
		panic(err)
	}
	defer RemoteResp.Body.Close()
	token, mail, user, repo := getUserData()
	targetURL := fmt.Sprintf(FILE_UPLOAD_URL+"/"+To, user, repo)
	RemoteData, err := io.ReadAll(RemoteResp.Body)
	reqjson := fmt.Sprintf(`{"message":"ADDED NEW FILE","committer":{"name":"%s","email":"%s"},"content":"%s"}`, user, mail, base64.StdEncoding.EncodeToString(RemoteData))

	GithubReq, err := http.NewRequest(http.MethodPut, targetURL, bytes.NewBuffer([]byte(reqjson)))
	if err != nil {
		panic(err)
	}
	GithubReq.Header.Add("Accept", "application/vnd.github+json")
	GithubReq.Header.Set("Authorization", "Bearer "+token)
	GithubResp, err := http.DefaultClient.Do(GithubReq)
	if err != nil {
		panic(err)
	}
	defer GithubResp.Body.Close()
	return GithubResp.StatusCode
}
