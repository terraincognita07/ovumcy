package security

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var (
	errNegativeLength = errors.New("length must be non-negative")
	errEmptyAlphabet  = errors.New("alphabet must not be empty")
)

// RandomString returns a cryptographically secure, unbiased string of the requested length.
func RandomString(length int, alphabet string) (string, error) {
	if length < 0 {
		return "", errNegativeLength
	}
	if length == 0 {
		return "", nil
	}
	if len(alphabet) == 0 {
		return "", errEmptyAlphabet
	}

	limit := big.NewInt(int64(len(alphabet)))
	value := make([]byte, length)
	for index := range value {
		position, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		value[index] = alphabet[position.Int64()]
	}

	return string(value), nil
}
