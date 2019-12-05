package static

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const (
	// EnvStaticDefaultFolder Default folder for files
	EnvStaticDefaultFolder = "STATIC_DEFAULT_FOLDER"
	// EnvStaticVariantFolder Variant folder containing override files
	EnvStaticVariantFolder = "STATIC_VARIANT_FOLDER"
)

// File A generic interface for a file, which could be pulled from local
// filesystem, or a remote URL
type File interface {
	Bytes() ([]byte, error)
	Name() (string, error)
	RealPath() string
}

// UnionFile Reference to a file on the local filesystem
type UnionFile struct {
	Path     string
	fileInfo os.FileInfo
	bytes    *[]byte
}

// NewUnionFile Returns a new UnionFile struct
func NewUnionFile(path string) (UnionFile, error) {
	defaultPath := os.Getenv(EnvStaticDefaultFolder)
	if len(defaultPath) == 0 {
		return UnionFile{}, fmt.Errorf(EnvStaticDefaultFolder + " environment variable cannot be null")
	}

	return UnionFile{Path: path}, nil
}

// Name Returns the file name
func (f *UnionFile) Name() (string, error) {
	path := CanonicalPath(f.Path)

	if f.fileInfo == nil {
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		f.fileInfo = info
	}

	return f.fileInfo.Name(), nil
}

// Bytes Returns the bytes for this file if it exists on local system
func (f *UnionFile) Bytes() ([]byte, error) {
	// We cache the value of the bytes so we don't re-read
	if f.bytes == nil {
		path := CanonicalPath(f.Path)

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return b, err
		}

		f.bytes = &b
	}

	return *f.bytes, nil
}

// RealPath Returns the path to this file
func (f *UnionFile) RealPath() string {
	return CanonicalPath(f.Path)
}

// CanonicalPath Returns the path for a file.  It checks first within the
// variant folder, and then within the default.  It does some cleaning of
// the url and checking to ensure we haven't escaped the sandbox of our static
// or default folders
func CanonicalPath(file string) string {
	var dir string
	defaultPath := os.Getenv(EnvStaticDefaultFolder)
	variantPath := os.Getenv(EnvStaticVariantFolder)

	// Variant folder specified, so check there and return if exists
	if len(variantPath) > 0 {
		dir = sandboxedFile(variantPath, file)
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}

	return sandboxedFile(defaultPath, file)
}

// sandboxedFile Returns the folder in question, ensuring that someone can't
// escape our sandboxed folder and access files outside
func sandboxedFile(dir string, file string) string {
	return filepath.Join(dir, filepath.FromSlash(path.Clean("/"+file)))
}
