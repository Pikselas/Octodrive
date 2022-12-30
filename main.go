package main

import (
	"Octo/Octo"
	"os"
)

func main() {
	_, err := Octo.NewOctoUser("Pikselas", os.Getenv("OCTODRIVE_MAIL"), os.Getenv("OCTODRIVE_TOKEN"))
	if err != nil {
		panic(err)
	}
	f, err := Octo.NewFileNavigator("Pikselas", "CopyPaster2", os.Getenv("OCTODRIVE_TOKEN"), "")
	if err != nil {
		panic(err)
	}
	for _, v := range f.GetItemList() {
		println(v.Name, v.IsDir)
	}
	err = f.GotoDirectory("Lillian.mp4")
	if err != nil {
		panic(err)
	}
	for _, v := range f.GetItemList() {
		println(v.Name, v.IsDir)
	}
}
