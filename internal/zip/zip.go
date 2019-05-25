// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package zip

import (
	"archive/zip"
	"io"
	"os"
)

// Zip zips a list of files into a zip file located at dst.
func Zip(files []string, dst string) error {
	file, err := os.Open(dst)
	if err != nil {
		return err
	}

	err = RawZip(files, file)
	file.Close()
	return err
}

//RawZip zips a list of files and writes the zip to a io.Writer.
func RawZip(files []string, dst io.Writer) error {
	w := zip.NewWriter(dst)
	for _, v := range files {
		zipEntry, err := w.Create(v)
		if err != nil {
			return err
		}
		file, err := os.Open(v)
		if err != nil {
			return err
		}

		_, err = io.Copy(zipEntry, file)
		if err != nil {
			return err
		}
		file.Close()
	}

	err := w.Close()
	if err != nil {
		return err
	}

	return nil
}
