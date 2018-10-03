package manifest

import (
	"encoding/json"
	"os"
)

// Load opens the named file for reading. If successful, the manifest is returned.
func Load(name string) (*Manifest, error) {
	mfst := New()
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&mfst); err != nil {
		return nil, err
	}
	return mfst, nil
}
