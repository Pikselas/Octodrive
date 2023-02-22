package Octo

type fileDetails struct {
	Name  string   `field:"name"`
	Paths []string `field:"paths"`
	Size  int64    `field:"size"`
}
