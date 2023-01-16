package main

import (
	"Octo/Octo"
	"net/http"
	"os"
)

func main() {
	u, err := Octo.NewOctoUser("Pikselas", os.Getenv("OCTODRIVE_MAIL"), os.Getenv("OCTODRIVE_TOKEN"))
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://vdownload-45.sb-cd.com/1/2/12185534-720p.mp4?secure=_XDdZyB4ogwzO2uqqIamQg,1673920678&m=45&d=1&_tid=12185534", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	sl := Octo.SourceLimiter{Source: res.Body, MaxSize: 200000000}
	err = u.CreateFile("test.mp4", &sl)
	if err != nil {
		panic(err)
	}
}
