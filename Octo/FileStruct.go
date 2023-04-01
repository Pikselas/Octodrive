package Octo

type fileDetails struct {
	Name        string
	Paths       []string
	Size        uint64
	ChunkSize   uint64
	MaxRepoSize uint64
	Key         []byte
}
