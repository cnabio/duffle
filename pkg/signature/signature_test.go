package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deis/duffle/pkg/bundle"
)

func TestSigner_Sign(t *testing.T) {
	is := assert.New(t)
	k := getKey(keyEmail, t)
	is.NotNil(k)

	b := &bundle.Bundle{
		Name:    "mybundle",
		Version: "1.2.3",
	}

	// Sign the bundle twice and make sure the sigs are the same.
	s := NewSigner(k)
	sig1, err := s.Clearsign(b)
	is.NoError(err)
	sig2, err := s.Clearsign(b)
	is.NoError(err)
	is.Equal(sig1, sig2)
	is.NotEmpty(sig1)

	// Verify that the returned identities are the same
	ring, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	v := NewVerifier(ring)
	signedBy, err := v.Verify(sig1)
	is.NoError(err)
	id1 := k.entity.Identities[fullKeyID]
	id2 := signedBy.entity.Identities[fullKeyID]
	is.NotNil(id1)
	is.NotNil(id2)
	is.Equal(id1.Name, id2.Name)
}

func TestSigner_Attest(t *testing.T) {
	is := assert.New(t)
	k := getKey(keyEmail, t)
	is.NotNil(k)

	b := &bundle.Bundle{
		Name:    "mybundle",
		Version: "1.2.3",
	}

	// Sign the bundle twice and make sure the sigs are the same.
	s := NewSigner(k)
	sig, err := s.Clearsign(b)
	is.NoError(err)
	attestation, err := s.Attest(sig)
	is.NoError(err)
	is.Contains(string(sig), string(attestation))
}

func getKey(keyname string, t *testing.T) *Key {
	k, err := LoadKeyRing(keyringFile)
	assert.NoError(t, err)

	key, err := k.Key(keyname)
	key.PassphraseFetcher = testPassphraseFetch
	assert.NoError(t, err)
	return key
}
