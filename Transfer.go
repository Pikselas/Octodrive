package main

import (
	"ToOcto/ToOcto"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func Upload() {
	var input int
	fmt.Println("Enter the transfer mode (remote:1 | local:2):")
	fmt.Scanln(&input)
	OctoUser := ToOcto.NewOctoUser("Pikselas",
		os.Getenv("OCTODRIVE_MAIL"),
		os.Getenv("OCTODRIVE_TOKEN"))
	var source_location string
	var path string
	var repo string
	var source_size int64
	fmt.Println("Enter the source location:")
	fmt.Scanln(&source_location)
	fmt.Println("Enter the path:")
	fmt.Scanln(&path)
	fmt.Println("Enter the repo:")
	var source io.ReadCloser
	if input == 1 {
		resp, err := http.Get(source_location)
		if err != nil {
			panic(err)
		}
		source_size = resp.ContentLength
		source = resp.Body
	} else if input == 2 {
		file, err := os.Open(source_location)
		if err != nil {
			panic(err)
		}
		fst, _ := file.Stat()
		source_size = fst.Size()
		source = file
	}
	tra := OctoUser.NewMultiPartTransferer("Pikselas", repo, path, source)
	st := make([]int, 0)
	go func(src_size int64, status *[]int, tr *ToOcto.MultiPartTransferer) {
		for {
			fmt.Print("\033[H\033[2J")
			fmt.Println(src_size, (*tr).ReadCount())
			for _, v := range *status {
				fmt.Println(v)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}(source_size, &st, &tra)
	for {
		stat, _, err := tra.TransferPart()
		st = append(st, stat)
		if err == io.EOF || (stat != 201 && stat != 502) {
			break
		}
	}
	source.Close()
}

func Download() {
	fmt.Println("Enter Repo:")
	var repo string
	fmt.Scanln(&repo)
	fmt.Println("Enter Path:")
	var path string
	fmt.Scanln(&path)
	fmt.Println("Enter Part Count:")
	var part_count int
	fmt.Scanln(&part_count)
	var save_location string
	fmt.Println("Enter Save Location:")
	fmt.Scanln(&save_location)
	reader := NewOctoMultipartReader("https://raw.githubusercontent.com/Pikselas/"+repo+"/main/"+path, part_count)
	file, _ := os.Create(save_location)
	io.Copy(file, reader)
	file.Close()
}
