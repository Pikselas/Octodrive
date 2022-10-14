package main

import (
	"RemoteToOcto/RemoteToOcto"
	"fmt"
	"time"
)

func main() {
	t := RemoteToOcto.NewMultiPartTransferer("https://octodex.github.com/images/original.png", "CopyPaster", "Octo.jpg")
	cmp := 0
	go func(c *int) {
		*c = t.TransferMultiPart()
	}(&cmp)
	for cmp == 0 {
		fmt.Print("\033[H\033[2J")
		fmt.Println(t.RawTransferSize(), t.ReadCount(), t.EncodedSize())
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Println(cmp)
}
