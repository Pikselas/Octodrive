package Octo

type fileDetails struct {
	Name  string   `field:"name"`
	Paths []string `field:"paths"`
	Size  uint64   `field:"size"`
}
