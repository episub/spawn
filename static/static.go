package static

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"log"
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
	Name() (string, error)
	ETag() (string, error)
	io.ReadSeeker
}

// LocalFile References a file on the local filesystem
type LocalFile struct {
	Path     string
	osFile   *os.File
	fileInfo os.FileInfo
}

// UnionFile Reference to a file on the local filesystem
type UnionFile struct {
	osFile   *os.File
	Path     string
	fileInfo os.FileInfo
	bytes    *[]byte
}

// NewLocalFile Returns a new LocalFile struct.  Does not check if the file
// exists unless the file is called -- e.g., through read or seek
func NewLocalFile(path string) (LocalFile, error) {
	return LocalFile{Path: path}, nil
}

// Name Returns the file name
func (f *LocalFile) Name() (string, error) {
	if f.fileInfo == nil {
		info, err := os.Stat(f.Path)
		if err != nil {
			return "", err
		}
		f.fileInfo = info
	}

	return f.fileInfo.Name(), nil
}

// Read Reader
func (f *LocalFile) Read(b []byte) (n int, err error) {
	if f.osFile == nil {
		// Not yet open, so try opening:
		file, err := os.Open(f.Path)
		if err != nil {
			return 0, err
		}
		f.osFile = file
	}

	return f.osFile.Read(b)
}

// Seek Seeker
func (f *LocalFile) Seek(offset int64, whence int) (ret int64, err error) {
	if f.osFile == nil {
		// Not yet open, so try opening:
		file, err := os.Open(f.Path)
		if err != nil {
			return 0, err
		}
		f.osFile = file
	}

	return f.osFile.Seek(offset, whence)
}

// ETag Returns an etag for the file
func (f *LocalFile) ETag() (string, error) {
	return modifiedEtag(f.Path)
}

// modifiedEtag returns an etag based on the file full location and last
// modified value
func modifiedEtag(filePath string) (string, error) {
	file, err := os.Stat(filePath)

	if err != nil {
		return "", err
	}

	modifiedTime := file.ModTime()

	etagRaw := md5.Sum([]byte(fmt.Sprintf("%s:%s", filePath, modifiedTime)))
	etag := base64.StdEncoding.EncodeToString(etagRaw[:])
	return etag, nil
}

// NewUnionFile Returns a new UnionFile struct.  Does not check if the file
// exists unless the file is called -- e.g., through read or seek
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
	log.Printf("CANONICAL PATH: %s", path)

	if f.fileInfo == nil {
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		f.fileInfo = info
	}

	return f.fileInfo.Name(), nil
}

// Read Reader
func (f *UnionFile) Read(b []byte) (n int, err error) {
	if f.osFile == nil {
		// Not yet open, so try opening:
		file, err := os.Open(f.RealPath())
		if err != nil {
			return 0, err
		}
		f.osFile = file
	}

	return f.osFile.Read(b)
}

// Seek Seeker
func (f *UnionFile) Seek(offset int64, whence int) (ret int64, err error) {
	if f.osFile == nil {
		// Not yet open, so try opening:
		file, err := os.Open(f.RealPath())
		if err != nil {
			return 0, err
		}
		f.osFile = file
	}

	return f.osFile.Seek(offset, whence)
}

// ETag Returns an etag for the file
func (f *UnionFile) ETag() (string, error) {
	return modifiedEtag(f.RealPath())
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
