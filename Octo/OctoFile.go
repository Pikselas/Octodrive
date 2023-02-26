package Octo

import (
	"Octo/Octo/ToOcto"
	"fmt"
	"io"
	"net/http"
	"os"
)

func getRawFileRequest(Url string, Token string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, Url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3.raw")
	req.Header.Add("Authorization", "Bearer "+Token)
	return req, nil
}

type remoteReader struct {
	req *http.Request
	src io.ReadCloser
}

func (r *remoteReader) Read(p []byte) (n int, err error) {
	fmt.Println("Reading", r.src)
	if r.src == nil {
		fmt.Println("Requesting", r.req)
		res, err := http.DefaultClient.Do(r.req)
		fmt.Println(err)
		if err != nil {
			return 0, err
		}
		fmt.Println(res.Status)
		r.src = res.Body
	}
	c, err := r.src.Read(p)
	fmt.Println("Read", c, err)
	return c, err
}

func (r *remoteReader) Close() error {
	if r.src != nil {
		r.src.Close()
		r.src = nil
	}
	return nil
}

type octoFileReader struct {
	readers            []io.Reader
	current_read_index uint
	read_end           bool
}

type OctoFile struct {
	file       fileDetails
	user_name  string
	user_token string
	FileSize   uint64
}

func (of *OctoFile) LoadFull() (io.Reader, error) {
	return &octoFileReader{}, nil
}

func H() {
	octoFile := OctoFile{}
	octoFile.file.Size = MaxOctoRepoSize*3 + 2040320
	octoFile.LoadBytes(MaxOctoRepoSize*3+2040320, 100)
}

func (of *OctoFile) Dummy() {
	r, er := getRawFileRequest(ToOcto.GetOctoURL(of.user_name, of.file.Paths[0], of.file.Name), of.user_token)
	if er != nil {
		panic(er)
	}
	re := remoteReader{req: r}
	fi, err := os.Create("File.wmv")
	if err != nil {
		panic(err)
	}
	io.Copy(fi, &re)
}

func (of *OctoFile) LoadBytes(from uint64, to uint64) (io.Reader, error) {
	StartPathIndex := from / MaxOctoRepoSize
	EndPathIndex := to / MaxOctoRepoSize
	StartPartNo := from % MaxOctoRepoSize / FileChunkSize
	StartPartOffset := from % MaxOctoRepoSize % FileChunkSize
	EndPartNo := to % MaxOctoRepoSize / FileChunkSize
	EndPartOffset := to % MaxOctoRepoSize % FileChunkSize

	fmt.Println(StartPathIndex, EndPathIndex, StartPartNo, StartPartOffset, EndPartNo, EndPartOffset)

	Rdrs := make([]io.Reader, 0)
	if StartPathIndex == EndPathIndex && StartPartNo == EndPartNo {
		// Make a http Request to file with start and end range
		req, err := getRawFileRequest(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(StartPartNo)), of.user_token)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Range", "bytes="+fmt.Sprint(StartPartOffset)+"-"+fmt.Sprint(EndPartOffset))
		Rdrs = append(Rdrs, &remoteReader{req: req})
	} else if StartPathIndex == EndPathIndex {
		// Make a http request to first file with start range
		req, err := getRawFileRequest(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(StartPartNo)), of.user_token)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Range", "bytes="+fmt.Sprint(StartPartOffset)+"-")
		Rdrs = append(Rdrs, &remoteReader{req: req})
		// create range reader for intermediate files
		Rdrs = append(Rdrs, NewMultipartRangeReader(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name), int(StartPartNo+1), int(EndPartNo), of.user_token))
		// Make a http request to last file with end range
		req, err = getRawFileRequest(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(EndPartNo)), of.user_token)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Range", "bytes=0-"+fmt.Sprint(EndPartOffset))
		Rdrs = append(Rdrs, &remoteReader{req: req})
	} else {
		// Make http request to first file
		req, err := getRawFileRequest(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(StartPartNo)), of.user_token)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Range", "bytes="+fmt.Sprint(StartPartOffset)+"-")
		Rdrs = append(Rdrs, &remoteReader{req: req})
		// create range reader from after the first file to last
		partCount, err := getPartCount(of.user_name, of.user_token, of.file.Paths[StartPathIndex], of.file.Name)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, NewMultipartRangeReader(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name), int(StartPartNo+1), int(partCount), of.user_token))
		// create a loop from 2nd first path to 2nd last path and Range reader
		for i := StartPathIndex + 1; i < EndPathIndex; i++ {
			partCount, err := getPartCount(of.user_name, of.user_token, of.file.Paths[i], of.file.Name)
			if err != nil {
				return nil, err
			}
			Rdrs = append(Rdrs, NewMultipartReader(ToOcto.GetOctoURL(of.user_name, of.file.Paths[i], of.file.Name), int(partCount), of.user_token))
		}
		// create range reader from 0 to before the last file
		Rdrs = append(Rdrs, NewMultipartRangeReader(ToOcto.GetOctoURL(of.user_name, of.file.Paths[EndPathIndex], of.file.Name), 0, int(EndPartNo), of.user_token))
		// make http request to the last file
		req, err = getRawFileRequest(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(EndPartNo)), of.user_token)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Range", "bytes=0-"+fmt.Sprint(EndPartOffset))
		Rdrs = append(Rdrs, &remoteReader{req: req})
	}
	fmt.Println(len(Rdrs))
	return &octoFileReader{
		readers:  Rdrs,
		read_end: true,
	}, nil
}

func (r *octoFileReader) Read(p []byte) (n int, err error) {
	if r.read_end && r.current_read_index < uint(len(r.readers)) {
		r.read_end = false
	}
	if !r.read_end {
		n, err := r.readers[r.current_read_index].Read(p)
		if err == io.EOF {
			r.read_end = true
			r.current_read_index++
		} else if err != nil {
			return n, err
		}
		return n, nil
	}
	return 0, io.EOF
}
