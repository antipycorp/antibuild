package internal

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

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

var ErrFileNotExist = errors.New("file does not exist")

func DownloadFile(filepath string, url string, executable bool) error {
	// Create the file
	out, err := os.Create(filepath)
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
		os.Remove(filepath)
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

func GenCopy(src, dest string, info os.FileInfo) error {
	if info.IsDir() {
		return DirCopy(src, dest, info)
	}
	return FileCopy(src, dest, info)
}

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
