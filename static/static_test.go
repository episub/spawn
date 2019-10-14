package static

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

const (
	TestDefaultPath = "default_test"
	TestVariantPath = "variant_test"
)

var localFileTestCases = []struct {
	Path            string
	ExpectedName    string
	ExpectedPath    string
	ExpectedContent []byte
}{
	{
		Path:            "default",
		ExpectedName:    "default",
		ExpectedPath:    "default_test/default",
		ExpectedContent: []byte("default"),
	},
	{
		Path:            "variant",
		ExpectedName:    "variant",
		ExpectedPath:    "variant_test/variant",
		ExpectedContent: []byte("variant"),
	},
	{
		Path:            "/absolute/folder",
		ExpectedName:    "",
		ExpectedPath:    "/absolute/folder",
		ExpectedContent: []byte{},
	},
}

func TestCanonicalPath(t *testing.T) {
	os.Setenv(EnvStaticDefaultFolder, TestDefaultPath)
	os.Setenv(EnvStaticVariantFolder, TestVariantPath)

	for _, tc := range localFileTestCases {
		t.Run(fmt.Sprintf("Path %s", tc.Path), func(t *testing.T) {
			returned := CanonicalPath(tc.Path)

			if returned != tc.ExpectedPath {
				t.Errorf("Expected path %s, but was given %s", tc.ExpectedPath, returned)
			}

			// Try and open the file, seeing if contents match, if expected contents
			// has been provided
			if len(tc.ExpectedContent) > 0 {
				b, err := ioutil.ReadFile(returned)
				if err != nil {
					t.Error(err)
					return
				}

				if bytes.Compare(
					bytes.TrimSpace(tc.ExpectedContent),
					bytes.TrimSpace(b),
				) != 0 {
					t.Errorf("Expected content %s, but had %s", tc.ExpectedContent, b)
				}
			}
		})
	}
}

func TestLocalFile(t *testing.T) {
	os.Setenv(EnvStaticDefaultFolder, TestDefaultPath)
	os.Setenv(EnvStaticVariantFolder, TestVariantPath)

	for _, tc := range localFileTestCases {
		t.Run(fmt.Sprintf("Path %s", tc.Path), func(t *testing.T) {
			if len(tc.ExpectedContent) == 0 {
				// Skip this
				return
			}

			file := NewLocalFile(tc.Path)
			b, err := file.Bytes()

			if err != nil {
				t.Error(err)
				return
			}

			if bytes.Compare(
				bytes.TrimSpace(tc.ExpectedContent),
				bytes.TrimSpace(b),
			) != 0 {
				t.Errorf("Expected content %s, but had %s", tc.ExpectedContent, b)
			}

			// Check the file name is what we expect
			name, err := file.Name()
			if err != nil {
				t.Error(err)
			}

			if name != tc.ExpectedName {
				t.Errorf("Expected name %s, but had %s", tc.ExpectedName, name)
			}
		})
	}
}
