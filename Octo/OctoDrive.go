package Octo

import (
	"Octo/Octo/ToOcto"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	FileChunkSize   = 30015488
	MaxOctoRepoSize = FileChunkSize * 30
)

type OctoDrive interface {
	Create(path string, source io.Reader) error
	Load(path string) (OctoFile, error)
}

type user struct {
	user ToOcto.OctoUser
}

func (u *user) Create(path string, source io.Reader) error {
	var Repository string
	//make a new repository
	Repository = RandomString(10)
	status, err := u.user.CreateRepository(Repository, "Repository for OctoDrive contents")
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("error creating repository: %d", status)
	}
	fileID := RandomString(10)
	paths := make([]string, 0)
	paths = append(paths, Repository)
	//create a multipart transferer with source limiter for max repository size
	var reader SourceLimiter
	var ContentSize int64 = 0
	for {
		reader = NewSourceLimiter(source, MaxOctoRepoSize)
		//create a new multipart transferer
		transferer := ToOcto.NewMultiPartTransferer(u.user, Repository, fileID, FileChunkSize, reader)
		err = nil
		for err != io.EOF {
			_, _, err = transferer.TransferPart()
			fmt.Println(err)
		}
		ContentSize += reader.GetCurrentSize()
		if !reader.IsEOF() {
			print("Creating new repository")
			newRepo := RandomString(10)
			status, err := u.user.CreateRepository(newRepo, "Repository for OctoDrive contents")
			if err != nil {
				return err
			}
			if status != http.StatusCreated {
				return fmt.Errorf("error creating repository: %d", status)
			}
			paths = append(paths, newRepo)
			Repository = newRepo
		} else {
			break
		}
	}
	//create file details
	var FileDetails fileDetails
	FileDetails.Name = fileID
	FileDetails.Paths = paths
	FileDetails.Size = ContentSize
	data, err := json.Marshal(FileDetails)
	if err != nil {
		return err
	}
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	_, Str, err := u.user.Transfer("_Octofiles", "Contents/"+path, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	println(Str)
	return nil
}

func (u *user) Load(path string) (OctoFile, error) {
	//get file details
	req, err := u.user.MakeRequest(http.MethodGet, "_Octofiles", "Contents/"+path, nil, true)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var FileDetails fileDetails
	json.NewDecoder(res.Body).Decode(&FileDetails)
	return &octoFile{file: FileDetails, user: u.user, FileSize: uint64(FileDetails.Size)}, nil
}

func NewOctoDrive(User string, Mail string, Token string) (OctoDrive, error) {
	oU, err := ToOcto.NewOctoUser(User, Mail, Token)
	if err != nil {
		return nil, err
	}
	U := user{user: oU}
	status, err := oU.CreateRepository("_Octofiles", "Initial repo for OctoDrive contents")
	if err != nil {
		return nil, err
	}
	if status != http.StatusCreated && status != http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("error creating repository: %d", status)
	}
	return &U, err
}
