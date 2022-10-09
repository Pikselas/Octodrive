package main

import (
	"RemoteToOcto/RemoteToOcto"
	"fmt"
)

func main() {
	fmt.Println(RemoteToOcto.TransferMultiPart("https://octodex.github.com/images/original.png", "OctoBoy.jpg"))
}
