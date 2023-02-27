package Octo

import (
	"Octo/Octo/ToOcto"
	"fmt"
	"io"
	"net/http"
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

type OctoFile interface {
	Get() (io.Reader, error)
	GetName() string
	GetSize() uint64
	GetBytes(from uint64, to uint64) (io.Reader, error)
}

type octoFile struct {
	file       fileDetails
	user_name  string
	user_token string
	FileSize   uint64
}

func (of *octoFile) GetName() string {
	return of.file.Name
}

func (of *octoFile) GetSize() uint64 {
	return of.FileSize
}

func (of *octoFile) Get() (io.Reader, error) {
	Rdrs := make([]io.Reader, 0)
	for _, repo := range of.file.Paths {
		c, err := getPartCount(of.user_name, of.user_token, repo, of.file.Name)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, NewMultipartReader(ToOcto.GetOctoURL(of.user_name, repo, of.file.Name), int(c), of.user_token))
		if err != nil {
			return nil, err
		}
	}
	return &octoFileReader{readers: Rdrs, read_end: true}, nil
}

func (of *octoFile) GetBytes(from uint64, to uint64) (io.Reader, error) {
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
		Rdrs = append(Rdrs, io.LimitReader(&delayedReader{req: req, ignoreBytes: StartPartOffset}, int64(EndPartOffset-StartPartOffset)))
	} else if StartPathIndex == EndPathIndex {
		// Make a http request to first file with start range
		url := ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(StartPartNo))
		req, err := getRawFileRequest(url, of.user_token)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, &delayedReader{req: req, ignoreBytes: StartPartOffset})
		// create range reader for intermediate files
		Rdrs = append(Rdrs, NewMultipartRangeReader(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name), int(StartPartNo+1), int(EndPartNo), of.user_token))
		// Make a http request to last file with end range
		req, err = getRawFileRequest(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(EndPartNo)), of.user_token)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, &remoteReader{req: req})
	} else {
		// Make http request to first file
		req, err := getRawFileRequest(ToOcto.GetOctoURL(of.user_name, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(StartPartNo)), of.user_token)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, &delayedReader{req: req, ignoreBytes: StartPartOffset})
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
		Rdrs = append(Rdrs, io.LimitReader(&remoteReader{req: req}, int64(EndPartOffset)))
	}
	return &octoFileReader{
		readers:  Rdrs,
		read_end: true,
	}, nil
}
