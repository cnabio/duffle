package packager

import (
	"archive/tar"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExport(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "duffle-export")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	ex := Exporter{
		Source:      "testdata/examplebun",
		Destination: tempDir,
		Full:        false,
	}

	if err := ex.Export(); err != nil {
		t.Errorf("Expected no error, got error: %v", err)
	}
	expectedFile := filepath.Join(tempDir, "examplebun-0.1.0.tar.gz")
	_, err = os.Stat(expectedFile)
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but was not created", expectedFile)
	} else if err != nil {
		t.Errorf("Error with compressed bundle file: %v", err)
	}
}

func TestArchive(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "duffle-export")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	ex := Exporter{
		Source:      "testdata/examplebun",
		Destination: tempDir,
		Full:        false,
	}

	tarFile, err := ex.Archive("examplebun-0.1.0")
	if err != nil {
		t.Errorf("Expected no error, got error: %v", err)
	}

	expectedTarPath := filepath.Join(ex.Destination, "examplebun-0.1.0.tar")
	if tarFile != expectedTarPath {
		t.Errorf("Expected tarfile %s, got %s", expectedTarPath, tarFile)
	}

	_, err = os.Stat(expectedTarPath)
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but was not created", expectedTarPath)
	} else if err != nil {
		t.Errorf("Error with compressed bundle file: %v", err)
	}
}

func TestCompress(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "duffle-export")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tarHome, err := ioutil.TempDir("", "duffle-export-testtar")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tarHome)

	tarPath := filepath.Join(tarHome, "examplebun-0.1.0.tar")
	if err := mockTarFile(tarPath); err != nil {
		t.Fatal(err)
	}

	ex := Exporter{
		Source:      "testdata/examplebun",
		Destination: tempDir,
		Full:        false,
	}

	if err := ex.Compress(tarPath, gzipCompression); err != nil {
		t.Errorf("Expected no error, got error: %v", err)
	}
	expectedFile := "examplebun-0.1.0.tar.gz"
	expectedPath := filepath.Join(tempDir, expectedFile)
	if _, err := os.Stat(expectedPath); err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but does not", expectedPath)
	} else if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(tarPath); !os.IsNotExist(err) {
		t.Errorf("Expected tar file %s to be cleaned up but was not", tarPath)
	}

}

func mockTarFile(path string) error {
	tarfile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer tarfile.Close()

	tw := tar.NewWriter(tarfile)
	defer tw.Close()
	words := "some words"
	hdr := &tar.Header{
		Name: strings.TrimSuffix(filepath.Base(path), ".tar"),
		Mode: 0600,
		Size: int64(len(words)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(words)); err != nil {
		return err
	}

	return nil
}
