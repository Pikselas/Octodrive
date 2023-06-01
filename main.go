package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Pikselas/Octodrive/Octo"
	"github.com/Pikselas/Octodrive/Octo/ToOcto"
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
		re, err := of.GetSeekReader()
		if err != nil {
			panic(err)
		}
		re.Seek(parsedStart, io.SeekStart)
		defer re.Close()
		fmt.Println("Sending")
		n, err := io.Copy(w, re)
		fmt.Println("Sent", n, err)
	}
}

func MakeFileServer(drive *Octo.OctoDrive) {
	fn, err := drive.NewFileNavigator()
	if err != nil {
		panic(err)
	}
	fmt.Println("ALL SERVER IP on port 8080:")
	PrintIP()
	files := fn.GetItemList()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>OctoDrive</h1><ul>")
		for _, file := range files {
			if !file.IsDir {
				fmt.Fprintf(w, "<li><a href=\"/%s\">%s</a></li>", file.Name, file.Name)
			}
		}
		fmt.Fprint(w, "</ul></body></html>")
	})
	fmt.Println("\nTOTAL FILES", len(files))
	for _, file := range files {
		if !file.IsDir {
			of, err := drive.Load(file.Name)
			b := of.GetUserData()
			if len(b) > 0 {
				enc_dec := Octo.NewAesEncDecFrom(b)
				of.SetEncDec(enc_dec)
			}
			if err != nil {
				panic(err)
			}
			fmt.Println("Serving", file.Name)
			http.HandleFunc("/"+file.Name, StreamFile(of, "application/octet-stream"))
		}
	}
	http.ListenAndServe(":8080", nil)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type CheckSize struct {
	n   int64
	src io.Reader
}

func (c *CheckSize) Read(p []byte) (int, error) {
	n, err := c.src.Read(p)
	c.n += int64(n)
	return n, err
}

func (c *CheckSize) GetSize() int64 {
	return c.n
}

func UploadFile(name, path string, drive *Octo.OctoDrive) {
	file, err := os.Open(path)
	check(err)
	file_stat, err := file.Stat()
	check(err)
	size := file_stat.Size() / (1024 * 1024)
	defer file.Close()
	cF := CheckSize{src: file}
	f := drive.Create(&cF)
	enc_dec, err := Octo.NewAesEncDec()
	check(err)
	f.SetEncDec(enc_dec)
	f.SetUserData(enc_dec.GetKey())
	ch := make(chan struct{})
	go func() {
		for {
			select {
			case <-ch:
				return
			default:
				fmt.Println("\033[H\033[2J")
				fmt.Println(cF.GetSize()/(1024*1024), "MB /", size, "MB")
			}
		}
	}()
	for {
		err = f.WriteAll()
		if err == nil || err == io.EOF {
			break
		}
		for {
			err = f.RetryWriteChunk()
			if err == nil || err == io.EOF {
				break
			}
		}
	}
	close(ch)
	err = drive.Save(name, f)
	check(err)
}
func main() {
	octoUser, err := ToOcto.NewOctoUser(
		"Pikselas",
		"",
		"")

	if err != nil {
		panic(err)
	}
	drive, err2 := Octo.NewOctoDrive(octoUser)
	check(err2)
	MakeFileServer(drive)
}
