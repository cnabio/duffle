package signature

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadKeyRing(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	is.Len(k.entities, 2)
	is.Equal(k.entities[0].Identities[fullKeyID].UserId.Email, keyEmail)
	is.NotNil(k.entities[0].PrivateKey)
}

func TestLoadKeyRings(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRings(keyringFile, "testdata/public.gpg")
	is.NoError(err)
	is.Len(k.entities, 2)
	is.Equal(k.entities[0].Identities[fullKeyID].UserId.Email, keyEmail)
	is.NotNil(k.entities[0].PrivateKey)
}

func TestKeyRing_Len(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)
	is.Equal(k.Len(), 2)
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

	// Test that we can add the same keys again, and have it silently skip
	// duplicates.
	l := kr.Len()
	is.True(l > 0)
	extras, err = os.Open("testdata/extra.gpg")
	is.NoError(err)

	// Re-add extras
	is.NoError(kr.Add(extras))
	k2, err := kr.Key("extra1@example.com")
	is.NoError(err)
	is.Equal(k2.entity.Identities[fullExtraID].Name, fullExtraID)
	is.Equal(l, kr.Len())
}

func TestKeyRing_AddKey(t *testing.T) {
	is := assert.New(t)

	kr, err := LoadKeyRing(keyringFile)
	is.NoError(err)

	k, err := CreateKey(UserID{Name: "a", Comment: "b", Email: "c@e"})
	is.NoError(err)
	is.NotNil(k)

	kr.AddKey(k)
	k2, err := kr.Key("c@e")
	is.NoError(err)
	pk, err := k2.bestPrivateKey()
	is.NoError(err)
	is.NotNil(pk)

	// Test that if we re-add the same key it will be ignored.
	l := kr.Len()
	kr.AddKey(k)
	is.Equal(l, kr.Len())
}

func TestCreateKeyRing(t *testing.T) {
	is := assert.New(t)
	extras, err := os.Open("testdata/extra.gpg")
	is.NoError(err)

	kr := CreateKeyRing(testPassphraseFetch)
	is.NoError(kr.Add(extras))

	k, err := kr.Key("extra1@example.com")
	is.NoError(err)
	is.Equal(k.entity.Identities[fullExtraID].Name, fullExtraID)
}

func TestKeyRing_SavePrivate(t *testing.T) {
	is := assert.New(t)
	kr, err := LoadKeyRingFetcher(keyringFile, testPassphraseFetch)
	is.NoError(err)

	is.Error(kr.SavePrivate("testdata/noclobber.empty", false))

	dirname, err := ioutil.TempDir("", "signature-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		is.NoError(os.RemoveAll(dirname))
	}()

	newfile := filepath.Join(dirname, "save.gpg")
	// We do this to verify that the clobber flag is working.
	is.NoError(ioutil.WriteFile(newfile, []byte(" "), 0755))
	is.NoError(kr.SavePrivate(newfile, true))

	// Finally, we test loading the newly saved keyring
	kr2, err := LoadKeyRing(newfile)
	is.NoError(err)
	is.Len(kr2.entities, len(kr.entities))

	// Test that a known key exists.
	kk, err := kr2.Key("123A4002462DC23B")
	is.NoError(err)
	is.Equal(kk.entity.Identities[key2Email].Name, key2Email)
}

func TestKeyRing_SavePublic(t *testing.T) {
	is := assert.New(t)
	kr, err := LoadKeyRingFetcher(keyringFile, testPassphraseFetch)
	is.NoError(err)

	is.Error(kr.SavePublic("testdata/noclobber.empty", false))

	dirname, err := ioutil.TempDir("", "signature-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		is.NoError(os.RemoveAll(dirname))
	}()

	newfile := filepath.Join(dirname, "save.gpg")
	// We do this to verify that the clobber flag is working.
	is.NoError(ioutil.WriteFile(newfile, []byte(" "), 0755))
	is.NoError(kr.SavePublic(newfile, true))

	// Finally, we test loading the newly saved keyring
	kr2, err := LoadKeyRing(newfile)
	is.NoError(err)
	is.Len(kr2.entities, len(kr.entities))

	// Test that a known key exists.
	kk, err := kr2.Key("123A4002462DC23B")
	is.NoError(err)
	is.Equal(kk.entity.Identities[key2Email].Name, key2Email)

	// Test that the key does NOT have a private component
	_, err = kk.bestPrivateKey()
	is.Error(err)
}

func TestKeyRing_PrivateKeys(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(keyringFile)
	is.NoError(err)

	keys := k.PrivateKeys()
	is.Len(keys, 2)

	// Make sure we are not loading public keys
	k, err = LoadKeyRing(publicKeyFile)
	is.NoError(err)

	keys = k.PrivateKeys()
	is.Len(keys, 0)
}

func TestKeyRing_Keys(t *testing.T) {
	is := assert.New(t)
	k, err := LoadKeyRing(publicKeyFile)
	is.NoError(err)

	keys := k.Keys()
	is.Len(keys, 2)
}
