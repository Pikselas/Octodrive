package main

import (
	"Octo/Octo"
	"io"
	"os"
)

func main() {
	user := Octo.NewOctoUser("Pikselas", os.Getenv("OCTODRIVE_MAIL"), os.Getenv("OCTODRIVE_TOKEN"))
	reader := user.NewMultipartReader("Pikselas", "CopyPaster", "Octo.jpg")
	fl, _ := os.Create("Octo.jpg")
	io.Copy(fl, reader)
	fl.Close()
}
