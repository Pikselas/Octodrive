package main

import (
	"Octo/Octo"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func StreamFile(of Octo.OctoFile) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request")
		w.Header().Set("Content-Type", "video/x-ms-wmv")
		w.Header().Set("Accept-Ranges", "bytes")
		byteRange := r.Header.Get("Range")
		parsedStart := int64(0)
		bSplt := strings.Split(byteRange, "=")
		if len(bSplt) == 2 {
			bSplt = strings.Split(bSplt[1], "-")
			if len(bSplt) == 2 {
				var err error
				parsedStart, err = strconv.ParseInt(bSplt[0], 10, 64)
				if err != nil {
					panic(err)
				}
				fmt.Println(parsedStart)
			}
			w.Header().Add("Content-Range", fmt.Sprintf("bytes %d-%d/%d", parsedStart, of.GetSize(), of.GetSize()))
			w.Header().Set("Content-Length", fmt.Sprint(int64(of.GetSize())-parsedStart))
			w.WriteHeader(http.StatusPartialContent)
		} else {
			w.Header().Set("Content-Length", fmt.Sprint(of.GetSize()))
		}
		fmt.Println("Getting", uint64(parsedStart), of.GetSize(), parsedStart)
		re, err := of.GetBytes(uint64(parsedStart), of.GetSize())
		if err != nil {
			panic(err)
		}
		fmt.Println("Sending")
		io.Copy(w, re)
	}
}

func main() {
	u, err := Octo.NewOctoUser("Pikselas", os.Getenv("OCTODRIVE_MAIL"), os.Getenv("OCTODRIVE_TOKEN"))
	if err != nil {
		panic(err)
	}
	of, err := u.Load("File.mp4")
	if err != nil {
		panic(err)
	}
	f, err := os.Create("File.wmv")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r, err := of.Get()
	if err != nil {
		panic(err)
	}
	fmt.Println("Copying")
	io.Copy(f, r)
	fmt.Println("Done")
	http.HandleFunc("/f.wmv", StreamFile(of))
	http.ListenAndServe(":8080", nil)
}
