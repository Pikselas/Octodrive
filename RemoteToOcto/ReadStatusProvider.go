package RemoteToOcto

type readStatusProvider interface {
	ReadCount() int64
}
