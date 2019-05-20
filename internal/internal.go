// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package internal

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Unzip a zip file
func Unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		prefix := filepath.Clean(dest) + string(os.PathSeparator)

		// Check for CVE-2018-8008. AKA ZipSlip
		if !(len(fpath) >= len(prefix) && fpath[0:len(prefix)] == prefix) { // strings.HasPrefix(fpath, prefix) removed in order to remove the dependency
			return filenames, errors.New(fpath + ": illegal file path")
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}

// ErrFileNotExist means a file does not exist
var ErrFileNotExist = errors.New("file does not exist")

// DownloadFile a file using http
func DownloadFile(path string, url string, executable bool) error {
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		os.Remove(path)
		return ErrFileNotExist
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	if executable == true {
		out.Chmod(0777)
	}

	return nil
}

// GenCopy copies a file or a dir depending on the type
func GenCopy(src, dest string, info os.FileInfo) error {
	if info.IsDir() {
		return DirCopy(src, dest, info)
	}
	return FileCopy(src, dest, info)
}

// FileCopy copies a file
func FileCopy(src, dest string, info os.FileInfo) error {

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

// DirCopy copies a directory
func DirCopy(srcdir, destdir string, info os.FileInfo) error {
	if err := os.MkdirAll(destdir, info.Mode()); err != nil {
		return err
	}

	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := GenCopy(cs, cd, content); err != nil {
			return err
		}
	}
	return nil
}

// DownloadJSON file from the internet to a data object
func DownloadJSON(url string, data interface{}) error {
	if strings.HasPrefix(url, "http://") {
		return errors.New("only https json downloads are supported")
	}

	if !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}

	return nil
}

// DownloadGit clones a git repo
func DownloadGit(path string, url string, version string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	clone := exec.Command("git", "clone", url)
	clone.Dir = path
	err = clone.Run()
	if err != nil {
		return err
	}

	checkout := exec.Command("git", "checkout", version)
	checkout.Dir = path
	err = checkout.Run()
	if err != nil {
		return err
	}

	return nil
}

// CompileFromSource compiles a go program
func CompileFromSource(path string, outFile string) error {
	cmd := exec.Command("go", "build", "-o", outFile, filepath.Join(path, "main.go"))
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
