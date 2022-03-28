// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build linux

package armoryctl

import (
	"fmt"
	"log"
	"os"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

func checkI2C(bus int) (err error) {
	dev := fmt.Sprintf("/dev/i2c-%d", bus)

	if _, err = os.Stat(dev); os.IsNotExist(err) {
		err = fmt.Errorf("%s missing, ensure that i2c-dev kernel module is loaded", dev)
	}

	return
}

func I2CRead(bus int, addr int, reg uint8, size uint) (val []byte, err error) {
	err = checkI2C(bus)

	if err != nil {
		return
	}

	_, err = host.Init()

	if err != nil {
		return
	}

	b, err := i2creg.Open(fmt.Sprintf("%d", bus))

	if err != nil {
		return
	}
	defer func() { _ = b.Close() }() // make errcheck happy

	w := []byte{byte(reg)}
	r := make([]byte, size)

	if Logger != nil {
		log.Printf("I2C read addr:%#x reg:%#x\n", addr, reg)
	}

	err = b.Tx(uint16(addr), w, r)

	if err != nil {
		return
	}

	if Logger != nil {
		log.Printf("I2C read: %#x\n", r)
	}

	return r, nil
}

func I2CWrite(bus int, addr int, reg uint8, val []byte) (err error) {
	err = checkI2C(bus)

	if err != nil {
		return
	}

	_, err = host.Init()

	if err != nil {
		return
	}

	b, err := i2creg.Open(fmt.Sprintf("%d", bus))

	if err != nil {
		return
	}
	defer func() { _ = b.Close() }() // make errcheck happy

	var w []byte

	w = append(w, byte(reg))
	w = append(w, val...)

	if Logger != nil {
		log.Printf("I2C write addr:%#x reg:%#x val:%#x\n", addr, reg, w)
	}

	err = b.Tx(uint16(addr), w, nil)

	return
}
