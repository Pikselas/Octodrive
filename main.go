package main

import (
	"Octo/Octo"
	"os"
)

func main() {
	u, err := Octo.NewOctoUser("Pikselas", os.Getenv("OCTODRIVE_MAIL"), os.Getenv("OCTODRIVE_TOKEN"))
	if err != nil {
		panic(err)
	}
	u.CreateFile("test.txt", os.Stdin)
}
