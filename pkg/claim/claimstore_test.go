package claim

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/deis/duffle/pkg/utils/crud"
)

func TestCanSaveReadAndDelete(t *testing.T) {
	claim := New("foo")
	claim.Bundle = "foobundle"

	tempDir, err := ioutil.TempDir("", "duffletest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	storeDir := filepath.Join(tempDir, "claimstore")
	store := NewClaimStore(crud.NewFileSystemStore(storeDir, "json"))

	err = store.Store(*claim)
	if err != nil {
		t.Errorf("Failed to store: %s", err)
	}

	c, err := store.Read("foo")
	if err != nil {
		t.Errorf("Failed to read: %s", err)
	}

	if c.Bundle != claim.Bundle {
		t.Errorf("Expected to read back bundle %s, got %s", claim.Bundle, c.Bundle)
	}

	err = store.Delete("foo")
	if err != nil {
		t.Errorf("Failed to delete: %s", err)
	}

	_, err = store.Read("foo")
	if err == nil {
		t.Errorf("Should have had error reading after deletion but did not")
	}
}

func TestCanUpdate(t *testing.T) {
	claim := New("foo")
	claim.Bundle = "foobundle"
	rev := claim.Revision

	tempDir, err := ioutil.TempDir("", "duffletest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	storeDir := filepath.Join(tempDir, "claimstore")
	store := NewClaimStore(crud.NewFileSystemStore(storeDir, "json"))

	err = store.Store(*claim)

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
