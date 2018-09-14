package signature

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/openpgp"
)

// Key represents an individual signing key
//
// A key can be used to verify messages. If it also contains
// a private key, it can sign messages as well.
type Key struct {
	entity *openpgp.Entity
}

// KeyRing represents a collection of keys as specified by OpenPGP
type KeyRing struct {
	entities openpgp.EntityList
}

// Key returns the key with the given ID.
//
// ID is a hex ID or (conventionally) and email address.
//
// If no such key exists, this will return an error.
func (r *KeyRing) Key(id string) (*Key, error) {
	hexID, err := strconv.ParseInt(id, 16, 64)
	if err == nil {
		k := r.entities.KeysById(uint64(hexID))
		l := len(k)
		if l > 1 {
			return nil, fmt.Errorf("required one key, got %d", l)
		}
		if l == 1 {
			return &Key{entity: k[0].Entity}, nil
		}
		// Else fallthrough and try a string-based lookup
	}

	// If we get here, there was no key found when looking by hex ID.
	// So we try again by string name in the email field.
	for _, e := range r.entities {
		for _, ident := range e.Identities {
			if ident.UserId.Email == id {
				return &Key{entity: e}, nil
			}
		}
	}
	return nil, errors.New("key not found")
}

// LoadKeyring loads a keyring from a path.
func LoadKeyring(path string) (*KeyRing, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	entities, err := openpgp.ReadKeyRing(f)
	if err != nil {
		return nil, err
	}
	return &KeyRing{
		entities: entities,
	}, nil
}
