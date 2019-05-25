// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package zip

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Unzip a zip file into its dstination
func Unzip(src string, dst string) ([]string, error) {
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
		fpath := filepath.Join(dst, f.Name)

		prefix := filepath.Clean(dst) + string(os.PathSeparator)

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
