package loader

import (
	"path/filepath"
	"testing"

	"github.com/scothis/ruffle/pkg/signature"

	"github.com/stretchr/testify/assert"
)

var (
	testFixtures     = filepath.Join("..", "signature", "testdata")
	testPublicRing   = filepath.Join(testFixtures, "public.gpg")
	testSigned       = filepath.Join(testFixtures, "signed.json.asc")
	testFailedSigned = filepath.Join(testFixtures, "fail-signed.json.asc")
)

func TestSecureLoader(t *testing.T) {
	is := assert.New(t)
	kr, err := signature.LoadKeyRing(testPublicRing)
	if err != nil {
		t.Fatal(err)
	}

	loader := NewSecureLoader(kr)
	b, err := loader.Load(testSigned)
	is.NoError(err)
	is.Equal(b.Name, "example_bundle")
}

func TestSecureLoader_FailSignature(t *testing.T) {
	is := assert.New(t)
	kr, err := signature.LoadKeyRing(testPublicRing)
	if err != nil {
		t.Fatal(err)
	}

	loader := NewSecureLoader(kr)
	_, err = loader.Load(testFailedSigned)
	is.Error(err)
}
