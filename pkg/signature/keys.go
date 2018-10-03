package signature

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

// PassphraseFetcher receives a keyname, and is responsible for returning the associated passphrase
type PassphraseFetcher func(name string) ([]byte, error)

// Key represents an individual signing key
//
// A key can be used to verify messages. If it also contains
// a private key, it can sign messages as well.
type Key struct {
	PassphraseFetcher PassphraseFetcher
	entity            *openpgp.Entity
	// selectedPrivateKey is reserved for use in cases where we want to
	// set a specific private key instead of looking through the entity
	// to load the key. This is necessary when choosing subkeys.
	selectedPrivateKey *packet.PrivateKey
}

// bestPrivateKey will find a private key and decrypt it if necessary.
//
// If a specific key is pinned on selectedPrivateKey, that key will be used.
// Otherwise, it will use the strategy in findPrivateKey.
func (k *Key) bestPrivateKey() (*packet.PrivateKey, error) {

	pk, err := k.findPrivateKey()
	if err != nil {
		return pk, err
	}

	// If key is not encrypted, return now.
	if !pk.Encrypted {
		return pk, nil
	}

	return pk, decryptPassphrase(k.entity.PrimaryKey.KeyIdShortString(), pk, k.PassphraseFetcher)
}

// findPrivateKey finds an acceptable private key for signing.
//
// If selectedPrivateKey is set this will use that key. Otherwise, it
// will start with the subkeys and seek for a signer, defaulting back to
// the top-level key.
//
// If no keys have the CanSign flag set, this will return an error.
//
// Finally, if no selectedPrivateKey is set, this will set the found
// key so that once it is unlocked we can avoid re-decrypting it.
func (k *Key) findPrivateKey() (*packet.PrivateKey, error) {
	// If a private key has already been set, use that.
	if k.selectedPrivateKey != nil {
		return k.selectedPrivateKey, nil
	}
	e := k.entity

	// It may be the case that a master key cannot be used for signing. It is not
	// clear how to test for that case. (in subkeys, you can do sk.Sig.FlagSign)
	if e.PrivateKey != nil && e.PrivateKey.CanSign() {
		k.selectedPrivateKey = e.PrivateKey
		return e.PrivateKey, nil
	}
	for _, sk := range e.Subkeys {
		// FlagSign checks if it is allowed to sign, while CanSign
		// verifies that the algorithm is capable of signing.
		if sk.Sig.FlagSign && sk.PrivateKey.CanSign() {
			k.selectedPrivateKey = sk.PrivateKey
			return sk.PrivateKey, nil
		}
	}

	return nil, errors.New("no signing key found")
}

func canSign(k *openpgp.Key) bool {
	return k.SelfSignature.FlagSign
}

// KeyRing represents a collection of keys as specified by OpenPGP
type KeyRing struct {
	entities          openpgp.EntityList
	PassphraseFetcher PassphraseFetcher
}

// Add adds new keys to the keyring.
func (r *KeyRing) Add(armoredKeys io.Reader) error {
	entities, err := openpgp.ReadArmoredKeyRing(armoredKeys)
	if err != nil {
		return err
	}
	r.entities = append(r.entities, entities...)
	return nil
}

// Key returns the key with the given ID.
//
// ID is a hex ID or (conventionally) an email address.
//
// If no such key exists, this will return an error.
func (r *KeyRing) Key(id string) (*Key, error) {
	// NB: GnuPG allows any of the following to be used:
	// - Hex ID (we support)
	// - Email (we support)
	// - Substring match on OpenPGP User Name (we support if first two fail)
	// - Fingerprint
	// - OpenPGP User Name ("Name (Comment) <email>")
	// - Partial email
	// - Subject DN (x509)
	// - Issuer DN (x509)
	// - Keygrip (40 hex digits)

	hexID, err := strconv.ParseInt(id, 16, 64)
	println("looking for", hexID)
	if err == nil {
		k := r.entities.KeysById(uint64(hexID))
		l := len(k)
		if l > 1 {
			return nil, fmt.Errorf("required one key, got %d", l)
		}
		if l == 1 {
			return &Key{entity: k[0].Entity, PassphraseFetcher: r.PassphraseFetcher}, nil
		}
		// Else fallthrough and try a string-based lookup
	}

	// If we get here, there was no key found when looking by hex ID.
	// So we try again by string name in the email field. We also do weak matching
	// at the same time.
	weak := map[[20]byte]*openpgp.Entity{}
	for _, e := range r.entities {
		for _, ident := range e.Identities {
			// XXX Leave this commented section
			// It is not clear whether we should skip identities that were not self-signed
			// with the Sign flag on. Since the entity is at a higher level than the identity,
			// it seems like we are more interested in the entity's capability than the
			// identity the user requested, and we can always walk the subkeys to see if
			// any of those are allowed to sign. So I am leaving this commented.
			//if !ident.SelfSignature.FlagSign {
			//	continue
			//}
			if ident.UserId.Email == id {
				return &Key{entity: e, PassphraseFetcher: r.PassphraseFetcher}, nil
			}
			if strings.Contains(ident.Name, id) {
				weak[e.PrimaryKey.Fingerprint] = e
			}
		}
	}

	switch len(weak) {
	case 0:
		return nil, errors.New("key not found")
	case 1:
		for _, first := range weak {
			return &Key{entity: first, PassphraseFetcher: r.PassphraseFetcher}, nil
		}
	}
	return nil, errors.New("multiple matching keys found")
}

// Save writes a keyring to disk as a binary entity list.
//
// This is the standard format described by the OpenPGP specification. The file will thus be
// importable to any OpenPGP compliant app that can read entity lists (that is, a list of
// OpenPGP packets).
//
// Note that if the keyring contains encrypted keys, the saving process will need to
// decrypt every single key. Make sure the *KeyRing has a PassphraseFetcher before calling
// Save.
func (r *KeyRing) Save(filepath string, clobber bool) error {
	if !clobber {
		if _, err := os.Stat(filepath); err == nil {
			return errors.New("keyring file exists")
		}
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write to a buffer so we don't nuke a keychain.
	temp := bytes.NewBuffer(nil)
	for _, e := range r.entities {

		// The serializer has no decryption, so we have to do this manually before saving.
		// Yes, this is a major pain. But apparently encrypted keys cannot be serialized.
		if e.PrivateKey.Encrypted {
			if err := decryptPassphrase(e.PrimaryKey.KeyIdShortString(), e.PrivateKey, r.PassphraseFetcher); err != nil {
				return err
			}
		}

		for _, sk := range e.Subkeys {
			if sk.PrivateKey.Encrypted {
				if err := decryptPassphrase(e.PrimaryKey.KeyIdShortString()+" subkey", sk.PrivateKey, r.PassphraseFetcher); err != nil {
					return err
				}
			}
		}

		// According to the godocs, when we call this, we lose "signatures from other entities", but preserve public and private keys.
		if err := e.SerializePrivate(temp, nil); err != nil {
			return err
		}
	}
	_, err = io.Copy(f, temp)
	return err
}

func decryptPassphrase(msg string, pk *packet.PrivateKey, fetcher PassphraseFetcher) error {
	if fetcher == nil {
		return errors.New("unable to decrypt key")
	}
	pass, err := fetcher(msg)
	if err != nil {
		return err
	}

	return pk.Decrypt(pass)
}

// LoadKeyRing loads a keyring from a path.
func LoadKeyRing(path string) (*KeyRing, error) {
	// TODO: Should we create a default passphrase fetcher?
	return LoadKeyRingFetcher(path, nil)
}

// LoadKeyRingFetcher loads a keyring from a path.
//
// If PassphraseFetcher is non-nil, it will be called whenever an encrypted key needs to be decrypted.
// If left nil, this will cause the keyring to emit an error whenever an encrypted key needs to be decrypted.
func LoadKeyRingFetcher(path string, fetcher PassphraseFetcher) (*KeyRing, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	entities, err := openpgp.ReadKeyRing(f)
	if err != nil {
		return nil, err
	}
	return &KeyRing{
		entities:          entities,
		PassphraseFetcher: fetcher,
	}, nil
}
