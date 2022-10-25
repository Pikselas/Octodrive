package ToOcto

import "io"

type OctoUser interface {
	NewMultiPartTransferer(RepoUser string, Repo string, Path string, Source io.Reader) MultiPartTransferer
}

type user struct {
	token    string
	commiter CommiterType
}

func (u *user) NewMultiPartTransferer(RepoUser string, Repo string, Path string, Source io.Reader) MultiPartTransferer {
	return NewMultiPartTransferer(u.commiter, RepoUser, Repo, Path, u.token, Source)
}

func NewOctoUser(User string, Mail string, Token string) OctoUser {
	return &user{Token, CommiterType{User, Mail}}
}
