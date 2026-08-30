// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xfile

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/multierr"
)

// GetCurrentDirectory ...
func GetCurrentDirectory() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("err", err)
	}

	return strings.Replace(dir, "\\", "/", -1)
}

// GetCurrentPackage ...
func GetCurrentPackage() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("err", err)
	}

	return strings.Replace(dir, "\\", "/", -1)
}

// MakeDirectory ...
func MakeDirectory(dirs ...string) error {
	var errs error
	for _, dir := range dirs {
		if !Exists(dir) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				errs = multierr.Append(errs, err)
			}
		}
	}
	return errs
}

// ReadFile reads the entire file contents.
func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// ReadFileString reads the entire file contents as a string.
func ReadFileString(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile writes data to a file.
func WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

// WriteFileString writes a string to a file.
func WriteFileString(path string, data string) error {
	return ioutil.WriteFile(path, []byte(data), 0644)
}

// AppendFile appends data to a file.
func AppendFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// AppendFileString appends a string to a file.
func AppendFileString(path string, data string) error {
	return AppendFile(path, []byte(data))
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// MoveFile moves a file from src to dst.
func MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}

// DeleteFile deletes a file.
func DeleteFile(path string) error {
	return os.Remove(path)
}

// DeleteDir deletes a directory and all its contents.
func DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// GetFileSize returns the size of a file in bytes.
func GetFileSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// GetFileMode returns the file mode.
func GetFileMode(path string) (os.FileMode, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Mode(), nil
}

// IsFile checks if the path is a file.
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

// GetTempDir returns the temporary directory.
func GetTempDir() string {
	return os.TempDir()
}

// CreateTempFile creates a temporary file.
func CreateTempFile(pattern string) (*os.File, error) {
	return ioutil.TempFile("", pattern)
}

// CreateTempDir creates a temporary directory.
func CreateTempDir(pattern string) (string, error) {
	return ioutil.TempDir("", pattern)
}

// GetExt returns the file extension.
func GetExt(path string) string {
	return filepath.Ext(path)
}

// GetBase returns the base name of the file.
func GetBase(path string) string {
	return filepath.Base(path)
}

// GetDir returns the directory part of the path.
func GetDir(path string) string {
	return filepath.Dir(path)
}

// Abs returns an absolute representation of path.
func Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Rel returns a relative pathname from the base directory to the target path.
func Rel(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// Join joins path elements into a single path.
func Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Split splits path into directory and file name components.
func Split(path string) (string, string) {
	return filepath.Split(path)
}

// Walk walks the file tree rooted at root, calling fn for each file or directory.
func Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

// Glob returns the names of all files matching pattern.
func Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// Chmod changes the mode of the named file.
func Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// Chown changes the numeric uid and gid of the named file.
func Chown(path string, uid, gid int) error {
	return os.Chown(path, uid, gid)
}

// GetHomeDir returns the user's home directory.
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// GetWorkDir returns the current working directory.
func GetWorkDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// GetExecDir returns the directory of the executable.
func GetExecDir() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(execPath)
}
