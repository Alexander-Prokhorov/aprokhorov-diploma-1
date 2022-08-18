package hasher

type Hasher interface {
	RandomKey() (string, error)
	GenerateToken() (string, error)
	GetHash(password string, key string) string
}
