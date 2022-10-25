package main

import (
	"ToOcto/ToOcto"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	src, err := http.Get("https://octodex.github.com/images/original.png")
	if err != nil {
		panic(err)
	}
	defer src.Body.Close()
	tra := ToOcto.NewOctoUser("Pikselas",
		os.Getenv("OCTODRIVE_MAIL"),
		os.Getenv("OCTODRIVE_TOKEN")).NewMultiPartTransferer("Pikselas", "CopyPaster", "Octo.jpg", src.Body)

	st := make([]int, 0)
	go func(arr *[]int, tr *ToOcto.MultiPartTransferer) {
		len := src.ContentLength
		for {
			fmt.Print("\033[H\033[2J")
			fmt.Println(len, (*tr).ReadCount())
			for _, v := range *arr {
				fmt.Println(v)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}(&st, &tra)
	for {
		stat, _, err := tra.TransferPart()
		st = append(st, stat)
		if err == io.EOF || (stat != 201 && stat != 502) {
			break
		}
	}
}
