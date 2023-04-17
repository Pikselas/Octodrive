package Octo

import (
	"Octo/Octo/ToOcto"
	"bytes"
	"encoding/json"
	"errors"
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
	Update(path string, of *OctoFile) error
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
	Of := new(OctoFile)
	Of.file = FileDetails
	Of.user = drive.user
	return Of, nil
}

func (drive *octoDrive) Save(path string, of *OctoFile) error {
	data, err := json.Marshal(of.file)
	if err != nil {
		return err
	}
	Err := drive.user.Transfer(OctoFileRegistry, "Contents/"+path, bytes.NewBuffer(data))
	if Err != nil {
		return Err
	}
	return nil
}

func (dive *octoDrive) Update(path string, of *OctoFile) error {
	data, err := json.Marshal(of.file)
	if err != nil {
		return err
	}
	Err := dive.user.Update(OctoFileRegistry, "Contents/"+path, bytes.NewBuffer(data))
	if Err != nil {
		return Err
	}
	return nil
}

func (drive *octoDrive) NewFileNavigator() (FileNavigator, error) {
	return NewFileNavigator(drive.user, OctoFileRegistry, "Contents")
}

func NewOctoDrive(User string, Mail string, Token string) (OctoDrive, error) {
	oU := ToOcto.NewOctoUser(User, Mail, Token)
	od := octoDrive{user: oU}
	err := oU.CreateRepository(OctoFileRegistry, "Initial repo for OctoDrive contents")
	if err != nil {
		stat := err.StatusCode()
		if stat != http.StatusCreated && stat != http.StatusUnprocessableEntity {
			return nil, err
		}
	}
	return &od, nil
}

func EnableFileWrite(file *OctoFile, src io.Reader) error {
	if file == nil {
		return errors.New("file is nil")
	}
	if file.file.Paths == nil {
		return errors.New("invalid file")
	}
	file.path_index = len(file.file.Paths) - 1
	file.encrypter = newAesEncDec(file.file.Key[:32], file.file.Key[32:])
	file.src_data = src
	part_count, err := getPartCount(file.user, file.file.Paths[file.path_index], file.file.Name)
	if err != nil {
		return err
	}
	repo_size := uint64(part_count) * file.file.ChunkSize
	if repo_size < file.file.MaxRepoSize {
		file.chunk_index = int(part_count)
		file.repo_limiter = NewSourceLimiter(file.src_data, file.file.MaxRepoSize-repo_size)
	}
	return nil
}
