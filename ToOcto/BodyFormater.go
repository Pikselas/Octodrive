package ToOcto

import (
	"fmt"
	"io"
)

/*
 Formats giithub-API compatible body based on
 https://docs.github.com/en/rest/repos/contents#create-or-update-file-contents
*/

/*
state: represnts formation state (always use 0).
reader: from where contents can be read and injected in the "content" section of json.
commiter: author's details
*/
type BodyFormater struct {
	state    int
	reader   io.Reader
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
