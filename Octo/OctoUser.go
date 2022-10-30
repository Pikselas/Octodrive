package Octo

import (
	"Octo/Octo/ToOcto"
	"io"
)

type OctoUser interface {
	NewMultiPartTransferer(RepoUser string, Repo string, Path string, Source io.Reader) ToOcto.MultiPartTransferer
	NewMultipartReader(RepoUser string, Repo string, Path string, PartCount int) OctoMultiPartReader
}

type user struct {
	token    string
	commiter ToOcto.CommiterType
}

func (u *user) NewMultiPartTransferer(RepoUser string, Repo string, Path string, Source io.Reader) ToOcto.MultiPartTransferer {
	return ToOcto.NewMultiPartTransferer(u.commiter, RepoUser, Repo, Path, u.token, Source)
}

func (u *user) NewMultipartReader(RepoUser string, Repo string, Path string, PartCount int) OctoMultiPartReader {
	return NewMultipartReader(ToOcto.GetOctoURL(RepoUser, Repo, Path), PartCount, u.token)
}

func NewOctoUser(User string, Mail string, Token string) OctoUser {
	return &user{Token, ToOcto.CommiterType{Name: User, Email: Mail}}
}
