package main

import "fmt"

func main() {
	fmt.Println("Upload to Octo : 1 | Download from Octo : 2")
	var input int
	fmt.Scanln(&input)
	if input == 1 {
		Upload()
	} else if input == 2 {
		Download()
	}
}
