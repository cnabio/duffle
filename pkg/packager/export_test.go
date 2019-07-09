package packager

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/deislabs/duffle/pkg/imagestore"
	"github.com/deislabs/duffle/pkg/imagestore/imagestoremocks"

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

	imagesAdded := []string{}

	is := &imagestoremocks.MockStore{
		AddStub: func(im string) (string, error) {
			imagesAdded = append(imagesAdded, im)
			return "", nil
		},
	}

	ex := Exporter{
		source: source,
		imageStoreBuilder: &imagestoremocks.MockBuilder{
			ArchiveDirStub: func(archiveDir string) {
				const expectedArchiveDirSuffix = "examplebun-0.1.0-export"
				if !strings.HasSuffix(archiveDir, expectedArchiveDirSuffix) {
					t.Fatalf("expected archive ending in %s, got %s", expectedArchiveDirSuffix, archiveDir)
				}
			},
			LogsStub: func(io.Writer) {
			},
			BuildStub: func() (imagestore.Store, error) {
				return is, nil
			},
		},
		logs:   filepath.Join(tempDir, "export-logs"),
		loader: loader.NewLoader(),
	}

	if err := ex.Export(); err != nil {
		t.Errorf("Expected no error, got error: %v", err)
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

	imagesAdded := []string{}

	is := &imagestoremocks.MockStore{
		AddStub: func(im string) (string, error) {
			imagesAdded = append(imagesAdded, im)
			return "", nil
		},
	}

	ex := Exporter{
		source:      "testdata/examplebun/bundle.json",
		destination: filepath.Join(tempDir, "random-directory", "examplebun-whatev.tgz"),
		imageStoreBuilder: &imagestoremocks.MockBuilder{
			ArchiveDirStub: func(archiveDir string) {
				const expectedArchiveDirSuffix = "examplebun-0.1.0-export"
				if !strings.HasSuffix(archiveDir, expectedArchiveDirSuffix) {
					t.Fatalf("expected archive ending in %s, got %s", expectedArchiveDirSuffix, archiveDir)
				}
			},
			LogsStub: func(io.Writer) {
			},
			BuildStub: func() (imagestore.Store, error) {
				return is, nil
			},
		},
		logs:   filepath.Join(tempDir, "export-logs"),
		loader: loader.NewLoader(),
	}

	if err := ex.Export(); err == nil {
		t.Error("Expected path does not exist error, got no error")
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

	is := &imagestoremocks.MockStore{
		AddStub: func(im string) (string, error) {
			// return the same digest for all images, but only one of them has a digest in the bundle manifest so just
			// that one will fail verification
			return "sha256:222222228fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540", nil
		},
	}

	ex := Exporter{
		source: source,
		imageStoreBuilder: &imagestoremocks.MockBuilder{
			ArchiveDirStub: func(archiveDir string) {
			},
			LogsStub: func(io.Writer) {
			},
			BuildStub: func() (imagestore.Store, error) {
				return is, nil
			},
		},
		logs:   filepath.Join(tempDir, "export-logs"),
		loader: loader.NewLoader(),
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
