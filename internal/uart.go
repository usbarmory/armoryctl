// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build linux

package armoryctl

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/albenik/go-serial/v2"
)

func checkUART(path string) (err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = fmt.Errorf("%s missing", path)
	}

	return
}

func UARTWrite(path string, speed int, cmd string) (res string, err error) {
	err = checkUART(path)

	if err != nil {
		return
	}

	port, err := serial.Open(path, serial.WithBaudrate(speed))

	if err != nil {
		return
	}
	defer func() { _ = port.Close() }() // make errcheck happy

	if Logger != nil {
		log.Printf(">> %s\n", cmd)
	}

	_, err = port.Write([]byte(cmd))

	if err != nil {
		return
	}

	time.Sleep(500 * time.Millisecond)

	n := 0
	r := make([]byte, 1024)

	for {
		n, err = port.Read(r)

		if err != nil {
			return
		}

		if strings.Contains(string(r), "OK") {
			break
		}

		if strings.Contains(string(r), "ERROR") {
			err = errors.New("response error")
			return
		}

		if n == 0 {
			break
		}
	}

	res = string(r[len(cmd):])

	if Logger != nil {
		log.Printf("<< %s\n", res)
	}

	return
}
