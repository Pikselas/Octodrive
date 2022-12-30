package main

import (
	"Octo/Octo"
	"os"
)

func main() {
	_, err := Octo.NewOctoUser("Pikselas", os.Getenv("OCTO_MAIL"), os.Getenv("OCTO_TOKEN"))
	if err != nil {
		panic(err)
	}
}
