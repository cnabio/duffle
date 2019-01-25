package packager

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/duffle/pkg/loader"
	"github.com/deislabs/duffle/pkg/signature"
)

func TestExportSigned(t *testing.T) {
	testFixtures := filepath.Join("..", "signature", "testdata")
	testPublicRing := filepath.Join(testFixtures, "public.gpg")
	signedBun := filepath.Join(testFixtures, "signed.json.asc")

	tempDir, err := setupTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	if err := setupSignedBundle(tempDir, signedBun); err != nil {
		t.Fatal(err)
	}
	kr, err := signature.LoadKeyRing(testPublicRing)
	if err != nil {
		t.Fatal(err)
	}

	tempPWD, pwd, err := setupPWD()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Chdir(pwd)
		os.RemoveAll(tempPWD)
	}()

	destination := "example-cool-bundle-0.1.0.tgz"
	ex := Exporter{
		Source:      filepath.Join(tempDir, "bundle.cnab"),
		Destination: destination,
		Thin:        true,
		Logs:        filepath.Join(tempDir, "export-logs"),
		Loader:      loader.NewSecureLoader(kr),
		Unsigned:    false,
	}

	if err := ex.Export(); err != nil {
		t.Errorf("Expected no error, got error: %v", err)
	}

	_, err = os.Stat(destination)
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but was not created", destination)
	} else if err != nil {
		t.Errorf("Error with compressed bundle file: %v", err)
	}
}

func TestExport(t *testing.T) {
	source, err := filepath.Abs(filepath.Join("testdata", "examplebun", "bundle.json"))
	if err != nil {
		t.Fatal(err)
	}
	tempDir, tempPWD, pwd, err := setupExportTestEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll(tempDir)
		os.Chdir(pwd)
		os.RemoveAll(tempPWD)
	}()

	ex := Exporter{
		Source:   source,
		Thin:     true,
		Logs:     filepath.Join(tempDir, "export-logs"),
		Loader:   loader.NewDetectingLoader(),
		Unsigned: true,
	}

	if err := ex.Export(); err != nil {
		t.Errorf("Expected no error, got error: %v", err)
	}

	expectedFile := "examplebun-0.1.0.tgz"
	_, err = os.Stat(expectedFile)
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but was not created", expectedFile)
	} else if err != nil {
		t.Errorf("Error with compressed bundle file: %v", err)
	}
}

func TestExportCreatesFileProperly(t *testing.T) {
	tempDir, err := setupTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	ex := Exporter{
		Source:      "testdata/examplebun/bundle.json",
		Destination: filepath.Join(tempDir, "random-directory", "examplebun-whatev.tgz"),
		Thin:        true,
		Logs:        filepath.Join(tempDir, "export-logs"),
		Unsigned:    true,
		Loader:      loader.NewDetectingLoader(),
	}

	if err := ex.Export(); err == nil {
		t.Error("Expected path does not exist error, got no error")
	}

	if err := os.MkdirAll(filepath.Join(tempDir, "random-directory"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := ex.Export(); err != nil {
		t.Errorf("Expected no error, got error: %s", err)
	}

	expectedFile := filepath.Join(tempDir, "random-directory", "examplebun-whatev.tgz")
	_, err = os.Stat(expectedFile)
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but was not created", expectedFile)
	} else if err != nil {
		t.Errorf("Error with compressed bundle archive: %v", err)
	}
}

func setupSignedBundle(tempDir, signedBundle string) error {
	from, err := os.Open(signedBundle)
	if err != nil {
		return err
	}
	defer from.Close()
	to, err := os.OpenFile(filepath.Join(tempDir, "bundle.cnab"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	return err
}

func setupTempDir() (string, error) {
	tempDir, err := ioutil.TempDir("", "duffle-export-test")
	if err != nil {
		return "", err
	}
	return tempDir, nil
}

func setupPWD() (string, string, error) {
	tempPWD, err := ioutil.TempDir("", "duffle-export-test")
	if err != nil {
		return "", "", err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	if err := os.Chdir(tempPWD); err != nil {
		return "", "", err
	}

	return tempPWD, pwd, nil
}

func setupExportTestEnvironment() (string, string, string, error) {
	tempDir, err := setupTempDir()
	if err != nil {
		return "", "", "", err
	}

	tempPWD, pwd, err := setupPWD()
	if err != nil {
		return "", "", "", err
	}

	return tempDir, tempPWD, pwd, nil
}
