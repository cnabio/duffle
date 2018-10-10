package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	test1pw       = "password"
	keyringFile   = "testdata/keyring.gpg"
	fullKeyID     = "Test One (Signer) <test1@example.com>"
	keyEmail      = "test1@example.com"
	key2Email     = "test2@example.com"
	fullExtraID   = "Extra Key (Signer) <extra1@example.com>"
	publicKeyFile = "testdata/public.gpg"
)

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

func TestCreateKey(t *testing.T) {
	is := assert.New(t)
	u := UserID{
		Name:    "User Name",
		Comment: "Comment",
		Email:   "email@example.com",
	}
	k, err := CreateKey(u)
	is.NoError(err)
	is.NotNil(k.entity.PrimaryKey)
	is.NotNil(k.entity.PrivateKey)
	kk, err := k.bestPrivateKey()
	is.NoError(err)
	is.NotNil(kk)
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

func TestKey_UserID(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	key, err := k.Key("test2@example.com")
	is.NoError(err)

	u, err := key.UserID()
	is.NoError(err)
	is.Equal(u.String(), "test2@example.com <test2@example.com>")

	key, err = k.Key(fullKeyID)
	is.NoError(err)

	u, err = key.UserID()
	is.NoError(err)
	is.Equal(u.String(), fullKeyID)
}

func TestKey_Fingerprint(t *testing.T) {
	expect := "5D76 712C E625 988A 272A 7E28 9B79 91DD 4037 8340"

	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	key, err := k.Key("test2@example.com")
	is.NoError(err)

	is.Equal(key.Fingerprint(), expect)
}

func testPassphraseFetch(name string) ([]byte, error) {
	return []byte(test1pw), nil
}
