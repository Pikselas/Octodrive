package ToOcto

import "fmt"

const FILE_UPLOAD_URL = "https://api.github.com/repos/%s/%s/contents"

func getOctoURL(RepoUser string, Repo string, Path string) string {
	return fmt.Sprintf(FILE_UPLOAD_URL, RepoUser, Repo) + "/" + Path
}
