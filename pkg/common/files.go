// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// MakeName creates a node name that follows the ethereum convention
// for such names. It adds the operation system name and Go runtime version
// the name.
func MakeName(name, version string) string {
	return fmt.Sprintf("%s/v%s/%s/%s", name, version, runtime.GOOS, runtime.Version())
}

// FileExist checks if a file exists at filePath.
func FileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

// AbsolutePath returns datadir + filename, or filename if it is absolute.
func AbsolutePath(datadir string, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(datadir, filename)
}


func WriteFile(file string, content []byte) error {
	// Create the wallet directory with appropriate permissions
	// in case it is not present yet.

	if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
		const dirPerm = 0700
		if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
			return err
		}
	}

	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	return os.Rename(f.Name(), file)
}

func GetFileList(root, expZipFile string, isExtName bool) ([]string, error) {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	fileList := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// file name: Pxxxx.pa
		if isExtName {
			fileTemp := strings.Split(file.Name(), ".")
			if fileTemp[len(fileTemp)-1] == expZipFile {
				fileList = append(fileList, file.Name())
			}
		} else {
			fileTemp := strings.Split(file.Name(), ".")
			if fileTemp[len(fileTemp)-1] == expZipFile {
				fileList = append(fileList, fileTemp[0])
			}
		}
	}
	return fileList, nil
}

func ReadFile(file string) ([]byte, error) {
	jBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return jBytes, nil
}

func GetRandFilePath() string {
	dir, err := ioutil.TempDir("", "temp")
	if err != nil {
		panic(err)
	}
	return dir
}


