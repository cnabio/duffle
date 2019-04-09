package claim

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/stretchr/testify/assert"

	"github.com/deislabs/duffle/pkg/utils/crud"
)

func TestCanSaveReadAndDelete(t *testing.T) {
	is := assert.New(t)
	claim, err := New("foo")
	is.NoError(err)
	claim.Bundle = &bundle.Bundle{Name: "foobundle", Version: "0.1.2"}

	tempDir, err := ioutil.TempDir("", "duffletest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	storeDir := filepath.Join(tempDir, "claimstore")
	store := NewClaimStore(crud.NewFileSystemStore(storeDir, "json"))

	is.NoError(store.Store(*claim), "Failed to store: %s", err)

	c, err := store.Read("foo")
	is.NoError(err, "Failed to read: %s", err)
	is.Equal(c.Bundle, claim.Bundle, "Expected to read back bundle %s, got %s", claim.Bundle.Name, c.Bundle.Name)

	claims, err := store.List()
	is.NoError(err, "Failed to list: %s", err)
	is.Len(claims, 1)
	is.Equal(claims[0], claim.Name)

	is.NoError(store.Delete("foo"))

	_, err = store.Read("foo")
	is.Error(err, "Should have had error reading after deletion but did not")
}

func TestCanUpdate(t *testing.T) {
	claim, err := New("foo")
	assert.NoError(t, err)
	claim.Bundle = &bundle.Bundle{Name: "foobundle", Version: "0.1.2"}
	rev := claim.Revision

	tempDir, err := ioutil.TempDir("", "duffletest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	storeDir := filepath.Join(tempDir, "claimstore")
	store := NewClaimStore(crud.NewFileSystemStore(storeDir, "json"))

	err = store.Store(*claim)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Millisecond)
	claim.Update(ActionInstall, StatusSuccess)

	err = store.Store(*claim)
	if err != nil {
		t.Errorf("Failed to update: %s", err)
	}

	c, err := store.Read("foo")
	if err != nil {
		t.Errorf("Failed to read: %s", err)
	}

	if c.Result.Action != ActionInstall {
		t.Errorf("Expected to read back action %s, got %s", ActionInstall, c.Result.Action)
	}
	if c.Revision == rev {
		t.Errorf("Expected to read back new revision, got old revision %s", rev)
	}
}

func TestReadAll(t *testing.T) {
	is := assert.New(t)

	tempDir, err := ioutil.TempDir("", "duffletest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	storeDir := filepath.Join(tempDir, "claimstore")
	store := NewClaimStore(crud.NewFileSystemStore(storeDir, "json"))

	claim, err := New("foo")
	is.NoError(err)
	claim.Bundle = &bundle.Bundle{Name: "foobundle", Version: "0.1.0"}

	is.NoError(store.Store(*claim), "Failed to store: %s", err)

	claim2, err := New("bar")
	is.NoError(err)
	claim2.Bundle = &bundle.Bundle{Name: "barbundle", Version: "0.1.0"}

	is.NoError(store.Store(*claim2), "Failed to store: %s", err)

	claim3, err := New("baz")
	is.NoError(err)
	claim3.Bundle = &bundle.Bundle{Name: "bazbundle", Version: "0.1.0"}

	is.NoError(store.Store(*claim3), "Failed to store: %s", err)

	claims, err := store.ReadAll()
	is.NoError(err, "Failed to read claims: %s", err)

	is.Len(claims, 3)
	is.Equal("foo", claim.Name)
	is.Equal("bar", claim2.Name)
	is.Equal("baz", claim3.Name)
}
