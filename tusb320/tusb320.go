// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) The armoryctl authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// Links:
//   http://www.ti.com/lit/ds/symlink/tusb320.pdf
//   https://github.com/usbarmory/usbarmory/wiki/I%C2%B2C-(Mk-II)

// Package tusb320 supports communication with the Texas Instruments TUSB320
// USB Type-C port controller.
package tusb320

import (
	"github.com/usbarmory/armoryctl/internal"
)

var (
	I2CBus     = 0
	I2CAddress = 0x61
)

// Current mode values and meaning (SSLSEN9E, Table 7).
var CurrentMode = map[byte]string{
	0x00: "0.5 A",
	0x01: "1.5 A",
	0x02: "0.5 A",
	0x03: "3.0 A",
}

func reverse(val []byte) []byte {
	for i := len(val)/2 - 1; i >= 0; i-- {
		rev := len(val) - 1 - i
		val[i], val[rev] = val[rev], val[i]
	}

	return val
}

// Get device identifier, reading I2C data address 0x00 - 0x07
// (SLLSEN9E, Table 7).
func GetDeviceID() (id []byte, err error) {
	id, err = armoryctl.I2CRead(I2CBus, I2CAddress, 0x00, 8)
	return reverse(id), err
}

// Get detected current advertisement, reading I2C data address 0x08 (CSR,
// (SLLSEN9E, Table 7) and extracting value CURRENT_MODE_ADVERTISE.
func GetCurrentMode() (mode byte, err error) {
	val, err := armoryctl.I2CRead(I2CBus, I2CAddress, 0x08, 1)

	if err != nil {
		return
	}

	mode = (val[0] >> 4) & 3

	return
}
