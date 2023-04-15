package main

import (
	"Octo/Octo"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func PrintIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println(ipnet.IP.String())
			}
		}
	}
}

func StreamFile(of *Octo.OctoFile, Type string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request")
		w.Header().Set("Content-Type", Type)
		w.Header().Set("Accept-Ranges", "bytes")
		byteRange := r.Header.Get("Range")
		parsedStart := int64(0)
		bSplit := strings.Split(byteRange, "=")
		if len(bSplit) == 2 {
			bSplit = strings.Split(bSplit[1], "-")
			if len(bSplit) == 2 {
				var err error
				parsedStart, err = strconv.ParseInt(bSplit[0], 10, 64)
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
		fmt.Println("Getting", parsedStart, of.GetSize(), parsedStart)
		re, err := of.GetBytes(uint64(parsedStart), of.GetSize())
		if err != nil {
			panic(err)
		}
		defer re.Close()
		fmt.Println("Sending")
		io.Copy(w, re)
		fmt.Println("Sent")
	}
}

func MakeFileServer(drive Octo.OctoDrive) {
	fn, err := drive.NewFileNavigator()
	if err != nil {
		panic(err)
	}
	fmt.Println("ALL SERVER IP on port 8080:")
	PrintIP()
	files := fn.GetItemList()
	fmt.Println("\nTOTAL FILES", len(files))
	for _, file := range files {
		if !file.IsDir {
			of, err := drive.Load(file.Name)
			if err != nil {
				panic(err)
			}
			fmt.Println("Serving", file.Name)
			http.HandleFunc("/"+file.Name, StreamFile(of, "application/octet-stream"))
		}
	}
	http.ListenAndServe(":8080", nil)
}

func main() {
	drive, err := Octo.NewOctoDrive("Pikselas", os.Getenv("OCTODRIVE_MAIL"), os.Getenv("OCTODRIVE_TOKEN"))

	if err != nil {
		panic(err)
	}
	file, err := os.Open("D:/Live pattern test.mp4")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	oFile, err := drive.Create(file)
	if err != nil {
		panic(err)
	}
	for {
		err := oFile.WriteAll()
		if err == nil || err == io.EOF {
			break
		}
		fmt.Println("\nRETRYING....\n", err)
		err = oFile.RetryWriteChunk()
		for err != nil && err != io.EOF {
			err = oFile.RetryWriteChunk()
			fmt.Println("ERR", err)
		}
	}
	err = drive.Save("LiveTex.mp4", oFile)
	if err != nil {
		panic(err)
	}
	MakeFileServer(drive)
}
