package main

import (
	"RemoteToOcto/RemoteToOcto"
	"fmt"
)

func main() {
	t := RemoteToOcto.NewMultiPartTransferer("https://octodex.github.com/images/original.png", "OctoCat.png")
	cmp := 0
	go func(c *int) {
		*c = t.TransferMultiPart()
	}(&cmp)
	for cmp == 0 {
		fmt.Print("\033[H\033[2J")
		fmt.Println(t.RawTransferSize(), t.ReadCount(), t.EncodedSize())
	}
	fmt.Println(cmp)
}
