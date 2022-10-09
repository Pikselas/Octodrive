package GithubAPI

import (
	"fmt"
	"io"
)

type BodyFormater struct {
	state    int
	reader   *RemoteReader
	commiter CommiterType
}

func (r *BodyFormater) Read(p []byte) (int, error) {
	if r.state == 0 {
		r.state = 1
		return copy(p, fmt.Sprintf(`{"message":"ADDED NEW FILE","committer":{"name":"%s","email":"%s"},"content":"`, r.commiter.Name, r.commiter.Email)), nil
	} else if r.state == 1 {
		count, err := r.reader.Read(p)
		if err == io.EOF {
			r.state = 2
			return count, nil
		}
		return count, err
	} else if r.state == 2 {
		r.state = 3
		return copy(p, `"}`), io.EOF
	}
	return 0, io.EOF
}
