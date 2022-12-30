package Octo

import (
	"Octo/Octo/ToOcto"
	"errors"
	"net/http"
)

type FileNavigator interface {
	CurrentDirectory() string
	GotoDirectory(path string) (bool, error)
	GotoParentDirectory() (bool, error)
	GotoChildDirectory(name string) (bool, error)
	GetItemList() ([]struct {
		IsDir bool
		Name  string
	}, error)
}

type fileNavigator struct {
	url               string
	token             string
	current_directory string
}

func (f *fileNavigator) checkPath(path string) (bool, error) {
	req, err := http.NewRequest("GET", f.url+"/"+path, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", "Bearer "+f.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

func (f *fileNavigator) CurrentDirectory() string {
	return f.current_directory
}

func (f *fileNavigator) GotoDirectory(path string) (bool, error) {
	ok, err := f.checkPath(path)
	if err != nil {
		return false, err
	}
	if ok {
		f.current_directory = path
	}
	return ok, nil
}

func (f *fileNavigator) GotoParentDirectory() (bool, error) {
	return false, nil
}

func (f *fileNavigator) GotoChildDirectory(name string) (bool, error) {
	return false, nil
}

func (f *fileNavigator) GetItemList() ([]struct {
	IsDir bool
	Name  string
}, error) {
	return nil, nil
}

func NewFileNavigator(RepoUser string, Repository string, token string, root string) (FileNavigator, error) {
	OctoUrl := ToOcto.GetOctoURL(RepoUser, Repository, root)
	req, err := http.NewRequest("GET", OctoUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("invalid user/repository/root => " + OctoUrl)
	}
	return &fileNavigator{OctoUrl, token, ""}, nil
}
