package loader

import (
	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/signature"
)

type SecureLoader struct {
	keyring *signature.KeyRing
	fetcher Fetcher
}

func NewSecureLoader(keyring *signature.KeyRing, fetcher Fetcher) *SecureLoader {
	return &SecureLoader{
		keyring: keyring,
		fetcher: fetcher,
	}
}

func (s *SecureLoader) Load() (*bundle.Bundle, error) {
	b := &bundle.Bundle{}
	data, err := s.fetcher.Bytes()
	if err != nil {
		return b, err
	}
	verifier := signature.NewVerifier(s.keyring)
	b, _, err = verifier.Extract(data)
	return b, err
}
