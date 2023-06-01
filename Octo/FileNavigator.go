package Octo

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"

	"github.com/Pikselas/Octodrive/Octo/ToOcto"
)

type ItemType struct {
	IsDir bool
	Name  string
}

var ErrorInvalidPath = errors.New("path not found")

type FileNavigator struct {
	user              *ToOcto.OctoUser
	root              string
	repository        string
	current_directory string
	dir_items         []ItemType
}

// checks if path is valid and sets dir_items
func (f *FileNavigator) checkPath(path string) error {
	req, err := f.user.MakeRequest(http.MethodGet, f.repository, f.root+"/"+path, nil, false)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ErrorInvalidPath
	}
	var jArr []interface{}
	json.NewDecoder(resp.Body).Decode(&jArr)
	ItemTypes := make([]ItemType, 0)
	for _, v := range jArr {
		itm := v.(map[string]interface{})
		ItemTypes = append(ItemTypes, ItemType{itm["type"].(string) == "dir", itm["name"].(string)})
	}
	f.dir_items = ItemTypes
	return nil
}

// returns current directory
func (f *FileNavigator) CurrentDirectory() string {
	return f.current_directory
}

// switches current directory to path
func (f *FileNavigator) GotoDirectory(path string) error {
	err := f.checkPath(path)
	if err == nil {
		f.current_directory = path
	}
	return err
}

// switches from current directory to parent directory
func (f *FileNavigator) GotoParentDirectory() error {
	p := path.Dir(f.current_directory)
	err := f.checkPath(p)
	if err == nil {
		f.current_directory = p
	}
	return err
}

// switches from current directory to given child directory
func (f *FileNavigator) GotoChildDirectory(name string) error {
	p := path.Join(f.current_directory, name)
	err := f.checkPath(p)
	if err == nil {
		f.current_directory = p
		return nil
	}
	return err
}

// returns list of items in current directory
func (f *FileNavigator) GetItemList() []ItemType {
	return f.dir_items
}

// creates a new FileNavigator
func NewFileNavigator(User *ToOcto.OctoUser, Repository string, Root string) (*FileNavigator, error) {
	f := FileNavigator{user: User, repository: Repository, root: Root, dir_items: make([]ItemType, 0)}
	err := f.checkPath("")
	if err != nil {
		return nil, errors.New("error fetching User/Repository/Root")
	}
	return &f, nil
}
