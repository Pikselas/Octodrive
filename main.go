package main

import (
	"RemoteToOcto/RemoteToOcto"
	"fmt"
	"os"
	"time"
)

func main() {
	t := RemoteToOcto.NewMultiPartTransferer(
		RemoteToOcto.CommiterType{
			Name:  "Pikselas",
			Email: os.Getenv("OCTODRIVE_MAIL")},
		os.Getenv("OCTODRIVE_TOKEN"))

	cmp := 0
	go func(c *int) {
		*c = t.TransferMultiPart("https://octodex.github.com/images/original.png", "CopyPaster", "Octo.jpg")
	}(&cmp)
	for cmp == 0 {
		fmt.Print("\033[H\033[2J")
		fmt.Println(t.RawTransferSize(), t.ReadCount(), t.EncodedSize())
		time.Sleep(50 * time.Millisecond)
	}
}
