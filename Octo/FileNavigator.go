package Octo

import (
	"Octo/Octo/ToOcto"
	"encoding/json"
	"errors"
	"net/http"
	"path"
)

type ItemType struct {
	IsDir bool
	Name  string
}

type FileNavigator interface {
	CurrentDirectory() string
	GotoDirectory(path string) error
	GotoParentDirectory() error
	GotoChildDirectory(name string) error
	GetItemList() []ItemType
}

var ErrorInvalidPath = errors.New("path not found")

type fileNavigator struct {
	url               string
	token             string
	current_directory string
	dir_items         []ItemType
}

func (f *fileNavigator) checkPath(path string) error {
	req, err := http.NewRequest("GET", f.url+"/"+path, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+f.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
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

func (f *fileNavigator) CurrentDirectory() string {
	return f.current_directory
}

func (f *fileNavigator) GotoDirectory(path string) error {
	err := f.checkPath(path)
	if err == nil {
		f.current_directory = path
	}
	return err
}

func (f *fileNavigator) GotoParentDirectory() error {
	p := path.Dir(f.current_directory)
	err := f.checkPath(p)
	if err == nil {
		f.current_directory = p
	}
	return err
}

func (f *fileNavigator) GotoChildDirectory(name string) error {
	p := path.Join(f.current_directory, name)
	err := f.checkPath(p)
	if err == nil {
		f.current_directory = p
		return nil
	}
	return err
}

func (f *fileNavigator) GetItemList() []ItemType {
	return f.dir_items
}

func NewFileNavigator(RepoUser string, Repository string, token string, root string) (FileNavigator, error) {
	f := fileNavigator{ToOcto.GetOctoURL(RepoUser, Repository, root), token, "", make([]ItemType, 0)}
	err := f.checkPath("")
	if err != nil {
		return nil, errors.New("error fetching User/Repository/Root")
	}
	return &f, nil
}
