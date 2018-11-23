package signature

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	signedFile       = "testdata/signed.json.asc"
	failedSignedFile = "testdata/fail-signed.json.asc"
)

func TestVerifier_Verify(t *testing.T) {
	is := assert.New(t)

	data, err := ioutil.ReadFile(signedFile)
	is.NoError(err)

	keyring, err := LoadKeyRing(keyringFile, false)
	is.NoError(err)

	v := NewVerifier(keyring)
	k, err := v.Verify(data)
	is.NoError(err)
	is.NotNil(k.entity.Identities[key2Email])
}
func TestVerifier_VerifyFail(t *testing.T) {
	is := assert.New(t)

	data, err := ioutil.ReadFile(failedSignedFile)
	is.NoError(err)

	keyring, err := LoadKeyRing(keyringFile, false)
	is.NoError(err)

	// The file was signed by a user not in the keyring, so should fail.
	v := NewVerifier(keyring)
	_, err = v.Verify(data)
	is.Error(err)
	is.Contains(err.Error(), "signature made by unknown entity")
}

func TestVerifier_Extract(t *testing.T) {
	is := assert.New(t)

	data, err := ioutil.ReadFile(signedFile)
	is.NoError(err)

	keyring, err := LoadKeyRing(keyringFile, false)
	is.NoError(err)
	v := NewVerifier(keyring)

	b, k, err := v.Extract(data)
	is.NoError(err)
	is.NotNil(k.entity.Identities[key2Email])
	is.Equal("example_bundle", b.Name)
}

func TestVerifier_ExtractFail(t *testing.T) {
	is := assert.New(t)

	data, err := ioutil.ReadFile(failedSignedFile)
	is.NoError(err)

	keyring, err := LoadKeyRing(keyringFile, false)
	is.NoError(err)
	v := NewVerifier(keyring)

	_, _, err = v.Extract(data)
	is.Error(err)
	is.Contains(err.Error(), "signature made by unknown entity")
}
