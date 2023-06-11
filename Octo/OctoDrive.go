package Octo

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Pikselas/Octodrive/Octo/ToOcto"
)

const (
	FileChunkSize   = 30015488 / 2
	MaxOctoRepoSize = FileChunkSize * 30 * 2
)

const (
	DefaultFileRegistry = "_Octofiles"
)

// OctoDrive stores files to GitHub
type OctoDrive struct {
	user          *ToOcto.OctoUser
	file_registry string
}

// Creates a new file
func (drive *OctoDrive) Create(src io.Reader) *OctoFile {
	file := new(OctoFile)
	file.file.Name = RandomString(10)
	file.file.ChunkSize = FileChunkSize
	file.file.MaxRepoSize = MaxOctoRepoSize
	file.file.Paths = make([]string, 0)
	file.path_index = -1
	file.src_data = src
	file.user = drive.user
	file.enc_dec = NewNilEncDec()
	return file
}

// Loads a file from path
func (drive *OctoDrive) Load(path string) (*OctoFile, error) {
	//get file details
	req, err := drive.user.MakeRequest(http.MethodGet, drive.file_registry, "Contents/"+path, nil, true)
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
	Of.enc_dec = NewNilEncDec()
	return Of, nil
}

// Saves a file to path
func (drive *OctoDrive) Save(path string, of *OctoFile) error {
	data, err := json.MarshalIndent(of.file, "", "  ")
	if err != nil {
		return err
	}
	Err := drive.user.Transfer(drive.file_registry, "Contents/"+path, bytes.NewBuffer(data))
	if Err != nil {
		return Err
	}
	return nil
}

// Updates a file at path
func (dive *OctoDrive) Update(path string, of *OctoFile) error {
	data, err := json.MarshalIndent(of.file, "", "  ")
	if err != nil {
		return err
	}
	Err := dive.user.Update(dive.file_registry, "Contents/"+path, bytes.NewBuffer(data))
	if Err != nil {
		return Err
	}
	return nil
}

// Creates a new file navigator
func (drive *OctoDrive) NewFileNavigator() (*FileNavigator, error) {
	return NewFileNavigator(drive.user, drive.file_registry, "Contents")
}

// Creates a new OctoDrive
func NewOctoDrive(user *ToOcto.OctoUser, base_repository string) (*OctoDrive, error) {
	od := new(OctoDrive)
	*od = OctoDrive{user: user, file_registry: base_repository}
	err := user.CreateRepository(base_repository, "Initial repo for OctoDrive contents")
	if err != nil {
		stat := err.StatusCode()
		if stat != http.StatusUnprocessableEntity {
			return nil, err
		}
	}
	return od, nil
}

// Enables a loaded file for writing
func EnableFileWrite(file *OctoFile, src io.Reader) error {
	if file == nil {
		return errors.New("file is nil")
	}
	if file.file.Paths == nil {
		return errors.New("invalid file")
	}
	file.path_index = len(file.file.Paths) - 1
	file.src_data = src
	part_count, err := getPartCount(file.user, file.file.Paths[file.path_index], file.file.Name)
	if err != nil {
		return err
	}

	// compare size to file chunk size
	// if size is less than file chunk size,
	// then update the last chunk with the new data

	req, err := file.user.MakeRequest(http.MethodGet, file.file.Paths[file.path_index], file.file.Name+"/"+strconv.Itoa(int(part_count-1)), nil, true)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if uint64(res.ContentLength) < file.file.ChunkSize {

		// update last chunk with chunk data + src data

		chunk_reader := NewSourceLimiter(io.MultiReader(res.Body, src), file.file.ChunkSize)
		octo_err := file.user.Update(file.file.Paths[file.path_index], file.file.Name+"/"+strconv.Itoa(int(part_count-1)), chunk_reader)
		if octo_err != nil {
			return octo_err
		}
		file.file.Size += uint64(chunk_reader.GetCurrentSize())
		if chunk_reader.GetCurrentSize() < file.file.ChunkSize {
			file.src_data = nil
			return nil
		}
	}

	// check repository size if repository
	// size is less than max repository size,
	// next chunks will be created here until max
	// repository size is reached

	repo_size := uint64(part_count) * file.file.ChunkSize
	if file.file.MaxRepoSize > repo_size {
		file.chunk_index = int(part_count)
		file.repo_limiter = NewSourceLimiter(file.src_data, file.file.MaxRepoSize-repo_size)
	}
	return nil
}
