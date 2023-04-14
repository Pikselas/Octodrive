package Octo

import (
	"Octo/Octo/ToOcto"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	OctoFileRegistry = "_Octofiles"
)

const (
	FileChunkSize   = 30015488
	MaxOctoRepoSize = FileChunkSize * 30
)

type OctoDrive interface {
	Create(source io.Reader) (*OctoFile, error)
	Save(path string, of *OctoFile) error
	Load(path string) (*OctoFile, error)
	NewFileNavigator() (FileNavigator, error)
}

type octoDrive struct {
	user ToOcto.OctoUser
}

func (drive *octoDrive) Create(src io.Reader) (*OctoFile, error) {
	file := new(OctoFile)
	file.file.Name = ToOcto.RandomString(10)
	file.file.ChunkSize = FileChunkSize
	file.file.MaxRepoSize = MaxOctoRepoSize
	file.file.Paths = make([]string, 0)
	file.path_index = -1
	file.src_data = src
	file.user = drive.user

	fileKey, err := generateKey(32)
	if err != nil {
		return nil, err
	}
	fileIV, err := generateKey(16)
	if err != nil {
		return nil, err
	}
	file.encrypter = newAesEncDec(fileKey, fileIV)
	file.file.Key = append(fileKey, fileIV...)
	return file, nil
}

func (drive *octoDrive) Save(path string, of *OctoFile) error {
	data, err := json.Marshal(of.file)
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

func (drive *octoDrive) Load(path string) (*OctoFile, error) {
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
	return &OctoFile{file: FileDetails, user: drive.user}, nil
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
