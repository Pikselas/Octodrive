package main

import (
	"RemoteToOcto/RemoteToOcto"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	src, err := http.Get("https://octodex.github.com/images/original.png")
	if err != nil {
		panic(err)
	}
	defer src.Body.Close()
	tra := RemoteToOcto.NewOctoUser("Pikselas",
		os.Getenv("OCTODRIVE_MAIL"),
		os.Getenv("OCTODRIVE_TOKEN")).NewMultiPartTransferer("Pikselas", "CopyPaster", "Temp", src.Body)
	for {
		stat, _, err := tra.TransferPart()
		if err == io.EOF {
			break
		}
		fmt.Println(stat)
	}
}
