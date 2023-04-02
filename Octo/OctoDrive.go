package Octo

import (
	"Octo/Octo/ToOcto"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	OctoFileRegistry = "_Octofiles"
)

const (
	FileChunkSize   = 30015488
	MaxOctoRepoSize = FileChunkSize * 30
)

type OctoDrive interface {
	Create(path string, source io.Reader) error
	Load(path string) (OctoFile, error)
	NewFileNavigator() (FileNavigator, error)
}

type octoDrive struct {
	user ToOcto.OctoUser
}

func (drive *octoDrive) Create(path string, source io.Reader) error {

	fileID := ToOcto.RandomString(10)
	fileKey, err := generateKey(32)
	if err != nil {
		return err
	}
	fileIV, err := generateKey(16)
	if err != nil {
		return err
	}
	filePaths := make([]string, 0)
	fileSize := uint64(0)
	for {
		repoLimiter := NewSourceLimiter(source, MaxOctoRepoSize)
		//make a new repository
		Repository := ToOcto.RandomString(10)
		status, err := drive.user.CreateRepository(Repository, "Repository for OctoDrive contents")
		if err != nil {
			return err
		}
		if status != http.StatusCreated {
			return fmt.Errorf("error creating repository: %d", status)
		}
		//create files
		var fileCount int = 0
		for {
			chunkLimiter := NewSourceLimiter(repoLimiter, FileChunkSize)
			enc_dec := newAesEncDec(fileKey, fileIV)
			encrypted_reader, err := enc_dec.Encrypt(chunkLimiter)
			if err != nil {
				return err
			}
			stat, str, err := drive.user.Transfer(Repository, fileID+"/"+strconv.Itoa(fileCount), encrypted_reader)
			if err != nil {
				return err
			}
			if stat != http.StatusCreated {
				return fmt.Errorf("error creating file: %d\n\n%s", stat, str)
			}
			fileCount++
			if chunkLimiter.IsEOF() {
				break
			}
		}
		filePaths = append(filePaths, Repository)
		fileSize += repoLimiter.GetCurrentSize()
		if repoLimiter.IsEOF() {
			break
		}
	}
	key := make([]byte, 48)
	copy(key, fileKey)
	copy(key[32:], fileIV)
	//create file details
	FileDetails := fileDetails{Name: fileID, Paths: filePaths, Size: fileSize, ChunkSize: FileChunkSize, MaxRepoSize: MaxOctoRepoSize, Key: key}
	data, err := json.Marshal(FileDetails)
	if err != nil {
		return err
	}
	Stat, _, err := drive.user.Transfer(OctoFileRegistry, "Contents/"+path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	println(Stat)
	return nil
}

func (drive *octoDrive) Load(path string) (OctoFile, error) {
	//get file details
	req, err := drive.user.MakeRequest(http.MethodGet, OctoFileRegistry, "Contents/"+path, nil, true)
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
	return &octoFile{file: FileDetails, user: drive.user}, nil
}

func (drive *octoDrive) NewFileNavigator() (FileNavigator, error) {
	return NewFileNavigator(drive.user, OctoFileRegistry, "Contents")
}

func NewOctoDrive(User string, Mail string, Token string) (OctoDrive, error) {
	oU, err := ToOcto.NewOctoUser(User, Mail, Token)
	if err != nil {
		return nil, err
	}
	od := octoDrive{user: oU}
	status, err := oU.CreateRepository(OctoFileRegistry, "Initial repo for OctoDrive contents")
	if err != nil {
		return nil, err
	}
	if status != http.StatusCreated && status != http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("error creating repository: %d", status)
	}
	return &od, err
}
