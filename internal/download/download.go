// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package download

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

// ErrFileNotExist means a file does not exist
var ErrFileNotExist = errors.New("file does not exist")

// File a file using http
func File(path string, url string, executable bool) error {
	url, valid := sanitizeURL(url)
	if !valid {
		return errors.New("an invalid URL is provided")
	}

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

	if executable {
		out.Chmod(0744)
	}

	return nil
}

// JSON file from the internet to a data object
func JSON(url string, data interface{}) error {
	url, valid := sanitizeURL(url)
	if !valid {
		return errors.New("an invalid URL is provided")
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

// Git clones a git repo
func Git(path string, url string, version string) error {
	err := os.MkdirAll(path, 0744)
	if err != nil {
		return err
	}

	cloneCMD := exec.Command("git", "clone", url)
	cloneCMD.Dir = path
	err = cloneCMD.Run()
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

func sanitizeURL(rawurl string) (string, bool) {
	u, err := url.Parse(rawurl)
	if err != nil || u.Host == "" {
		return "", false

	}

	if u.Scheme != "" {
		u.Scheme = "https"
	}
	return u.String(), true
}
