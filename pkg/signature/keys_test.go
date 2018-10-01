package signature

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	test1pw     = "password"
	keyringFile = "testdata/keyring.gpg"
	fullKeyID   = "Test One (Signer) <test1@example.com>"
	keyEmail    = "test1@example.com"
	key2Email   = "test2@example.com"
	fullExtraID = "Extra Key (Signer) <extra1@example.com>"
)

func TestLoadKeyRing(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	is.Len(k.entities, 2)
	is.Equal(k.entities[0].Identities[fullKeyID].UserId.Email, keyEmail)
	is.NotNil(k.entities[0].PrivateKey)
}

func TestKeyring_Key(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)

	key, err := k.Key("test1@example.com")
	is.NoError(err)

	is.Equal(key.entity.Identities[fullKeyID].UserId.Email, keyEmail)
}

func TestKeyring_MultipleKeys(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)

	_, err = k.Key("test")
	is.Error(err)
	is.Contains(err.Error(), "multiple matching keys found")
}

func TestKeyring_KeyByID(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)

	key, err := k.Key("6EFB02A2F77D9682")
	is.NoError(err)
	is.Equal(key.entity.Identities[fullKeyID].UserId.Email, keyEmail)

	key, err = k.Key("123A4002462DC23B")
	is.NoError(err)
	is.Equal(key.entity.Identities[key2Email].Name, key2Email)
}

func TestKeyRing_Add(t *testing.T) {
	is := assert.New(t)
	extras, err := os.Open("testdata/extra.gpg")
	is.NoError(err)
	kr, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	is.NoError(kr.Add(extras))

	k, err := kr.Key("extra1@example.com")
	is.NoError(err)
	is.Equal(k.entity.Identities[fullExtraID].Name, fullExtraID)
}

func TestKeyRing_Save(t *testing.T) {
	is := assert.New(t)
	kr, err := LoadKeyRingFetcher(keyringFile, testPassphraseFetch)
	is.NoError(err)

	is.Error(kr.Save("testdata/noclobber.empty", false))

	dirname, err := ioutil.TempDir("", "signature-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		is.NoError(os.RemoveAll(dirname))
	}()

	newfile := filepath.Join(dirname, "save.gpg")
	is.NoError(kr.Save(newfile, true))

	// Finally, we test loading the newly saved keyring
	kr2, err := LoadKeyRing(newfile)
	is.NoError(err)
	is.Len(kr2.entities, len(kr.entities))

	// Test that a known key exists.
	kk, err := kr2.Key("123A4002462DC23B")
	is.NoError(err)
	is.Equal(kk.entity.Identities[key2Email].Name, key2Email)
}

func TestKey(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	key, err := k.Key("test1@example.com")
	is.NoError(err)

	key.PassphraseFetcher = testPassphraseFetch

	pk, err := key.bestPrivateKey()
	is.NoError(err)
	is.NotNil(pk)
}

func TestKey_NoKeyFound(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	_, err = k.Key("test1111@example.com")
	is.Error(err)
}

func TestKey_NoPassphrase(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	key, err := k.Key("test2@example.com")
	is.NoError(err)

	// First, test without a fetcher
	pk, err := key.bestPrivateKey()
	is.NoError(err)
	is.NotNil(pk)

	// Set a fetcher, and make sure it doesn't force a call.
	key.PassphraseFetcher = func(name string) ([]byte, error) {
		return []byte("this should fail if there is a password on the key"), nil
	}

	pk, err = key.bestPrivateKey()
	is.NoError(err)
	is.NotNil(pk)
}

func testPassphraseFetch(name string) ([]byte, error) {
	return []byte(test1pw), nil
}
