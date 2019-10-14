package static

import (
	"io/ioutil"
	"os"
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
}

// LocalFile Reference to a file on the local filesystem
type LocalFile struct {
	Path     string
	fileInfo os.FileInfo
	bytes    *[]byte
}

// NewLocalFile Returns a new LocalFile struct
func NewLocalFile(path string) LocalFile {
	return LocalFile{Path: path}
}

// Name Returns the file name
func (f *LocalFile) Name() (string, error) {
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
func (f *LocalFile) Bytes() ([]byte, error) {
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

// CanonicalPath Returns the path for a file.  It checks if the path is an
// absolute one, and then returns that path if file exists, otherwise an
// error.  If path is relative, then checks if file exists in variant folder
// first, and if not, in default folder.  If file doesn't exist in either,
// returns error, otherwise returns the path
func CanonicalPath(file string) string {
	defaultPath := os.Getenv(EnvStaticDefaultFolder)
	variantPath := os.Getenv(EnvStaticVariantFolder)

	// Check if path is absolute, returning root file if so
	if []byte(file)[0] == []byte("/")[0] {
		return file
	}

	// Variant folder specified, so check there:
	if len(variantPath) > 0 {
		path := variantPath + "/" + file
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Default folder, if specified
	if len(defaultPath) > 0 {
		path := defaultPath + "/" + file
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Nothing found, so just return the provided file
	return file
}
