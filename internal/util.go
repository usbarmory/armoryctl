// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) The armoryctl authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build linux

package armoryctl

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var Logger *log.Logger

const traversalPattern = `../`

func UnzipFile(src string, dst string) (err error) {
	if Logger != nil {
		log.Printf("uncompressing %s in %s\n", src, dst)
	}

	reader, err := zip.OpenReader(src)

	if err != nil {
		return
	}
	defer func() { _ = reader.Close() }() // make errcheck happy

	err = os.MkdirAll(dst, 0700)

	if err != nil {
		return
	}

	for _, f := range reader.Reader.File {
		if strings.Contains(f.Name, traversalPattern) {
			return fmt.Errorf("path traversal detected")
		}

		dstPath := filepath.Join(dst, f.Name)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(dstPath, f.Mode())

			if err != nil {
				return
			}
		} else {
			err = os.MkdirAll(path.Dir(dstPath), 0700)

			if err != nil {
				return
			}

			output, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, f.Mode())

			if err != nil {
				return err
			}
			defer func() { _ = output.Close() }() // make errcheck happy

			input, err := f.Open()

			if err != nil {
				return err
			}
			defer func() { _ = input.Close() }() // make errcheck happy

			_, err = io.Copy(output, input)

			if err != nil {
				return err
			}

			_ = output.Close()
			_ = os.Chtimes(dstPath, f.ModTime(), f.ModTime())
		}
	}

	return
}

func ExecCommand(cmd string, args []string, root bool, input string) (output string, err error) {
	var c *exec.Cmd

	if root {
		c = exec.Command("/usr/bin/sudo", append([]string{cmd}, args...)...)
	} else {
		c = exec.Command(cmd, args...)
	}

	if Logger != nil {
		log.Printf("executing: %s %s\n", cmd, args)
	}

	var stdin bytes.Buffer
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if input != "" {
		_, err = stdin.WriteString(input)
		c.Stdin = &stdin

		if err != nil {
			err = fmt.Errorf("error writing to stdin")
			return
		}
	}

	c.Stdout = &stdout
	c.Stderr = &stderr

	err = c.Run()

	if err != nil {
		err = errors.New(stderr.String())
	}

	return stdout.String(), err
}
