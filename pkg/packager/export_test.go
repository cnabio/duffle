package packager

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/deislabs/duffle/pkg/loader"
)

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

	configArchiveDir := ""
	imagesAdded := []string{}

	is := &mockImageStore{
		configureStub: func(archiveDir string, logs io.Writer) error {
			configArchiveDir = archiveDir
			return nil
		},
		addStub: func(im string) (string, error) {
			imagesAdded = append(imagesAdded, im)
			return "", nil
		},
	}

	ex := Exporter{
		Source:     source,
		ImageStore: is,
		Logs:       filepath.Join(tempDir, "export-logs"),
		Loader:     loader.NewLoader(),
	}

	if err := ex.Export(); err != nil {
		t.Errorf("Expected no error, got error: %v", err)
	}

	expectedConfigArchiveDirBase := "examplebun-0.1.0-export"
	configArchiveDirBase := filepath.Base(configArchiveDir)
	if configArchiveDirBase != expectedConfigArchiveDirBase {
		t.Errorf("ImageStore.configure was passed an archive directory ending in %s; expected %s", configArchiveDirBase, expectedConfigArchiveDirBase)
	}

	expectedImagesAdded := []string{"mock/examplebun:0.1.0", "mock/image-a:58326809e0p19b79054015bdd4e93e84b71ae1ta", "mock/image-b:88426103e0p19b38554015bd34e93e84b71de2fc"}
	sort.Strings(expectedImagesAdded)
	sort.Strings(imagesAdded)
	if !reflect.DeepEqual(imagesAdded, expectedImagesAdded) {
		t.Errorf("ImageStore.add was called with %v; expected %v", imagesAdded, expectedImagesAdded)
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

	configArchiveDir := ""
	imagesAdded := []string{}

	is := &mockImageStore{
		configureStub: func(archiveDir string, logs io.Writer) error {
			configArchiveDir = archiveDir
			return nil
		},
		addStub: func(im string) (string, error) {
			imagesAdded = append(imagesAdded, im)
			return "", nil
		},
	}

	ex := Exporter{
		Source:      "testdata/examplebun/bundle.json",
		Destination: filepath.Join(tempDir, "random-directory", "examplebun-whatev.tgz"),
		ImageStore:  is,
		Logs:        filepath.Join(tempDir, "export-logs"),
		Loader:      loader.NewLoader(),
	}

	if err := ex.Export(); err == nil {
		t.Error("Expected path does not exist error, got no error")
	}

	expectedConfigArchiveDirBase := "examplebun-0.1.0-export"
	configArchiveDirBase := filepath.Base(configArchiveDir)
	if configArchiveDirBase != expectedConfigArchiveDirBase {
		t.Errorf("ImageStore.configure was passed an archive directory ending in %s; expected %s", configArchiveDirBase, expectedConfigArchiveDirBase)
	}

	expectedImagesAdded := []string{"mock/examplebun:0.1.0", "mock/image-a:58326809e0p19b79054015bdd4e93e84b71ae1ta", "mock/image-b:88426103e0p19b38554015bd34e93e84b71de2fc"}
	sort.Strings(expectedImagesAdded)
	sort.Strings(imagesAdded)
	if !reflect.DeepEqual(imagesAdded, expectedImagesAdded) {
		t.Errorf("ImageStore.add was called with %v; expected %v", imagesAdded, expectedImagesAdded)
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

func TestExportDigestMismatch(t *testing.T) {
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

	is := &mockImageStore{
		configureStub: func(archiveDir string, logs io.Writer) error {
			return nil
		},
		addStub: func(im string) (string, error) {
			// return the same digest for all images, but only one of them has a digest in the bundle manifest so just
			// that one will fail verification
			return "sha256:222222228fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540", nil
		},
	}

	ex := Exporter{
		Source:     source,
		ImageStore: is,
		Logs:       filepath.Join(tempDir, "export-logs"),
		Loader:     loader.NewLoader(),
	}

	if err := ex.Export(); err.Error() != "Error preparing artifacts: content digest mismatch: image mock/image-a:"+
		"58326809e0p19b79054015bdd4e93e84b71ae1ta has digest "+
		"sha256:222222228fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540 but the digest should be "+
		"sha256:111111118fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540 according to the bundle manifest" {
		t.Errorf("Unexpected error: %v", err)
	}
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

type mockImageStore struct {
	configureStub func(archiveDir string, logs io.Writer) error
	addStub       func(im string) (string, error)
}

func (i *mockImageStore) configure(archiveDir string, logs io.Writer) error {
	return i.configureStub(archiveDir, logs)
}

func (i *mockImageStore) add(im string) (string, error) {
	return i.addStub(im)
}
