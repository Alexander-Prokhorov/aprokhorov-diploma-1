package hasher

type Hasher interface {
	RandomKey() (string, error)
	GetHash(string, string) string
}
