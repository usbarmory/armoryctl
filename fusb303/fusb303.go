// armoryctl | https://github.com/inversepath/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// Links:
//   https://www.onsemi.com/pub/Collateral/FUSB303-D.PDF
//   https://github.com/inversepath/usbarmory/wiki/I%C2%B2C-(Mk-II)

package fusb303

import (
	"github.com/inversepath/armoryctl/internal"
)

// I2C bus number
var I2CBus = 0

// I2C address
var I2CAddress = 0x31

// Current mode values and meaning
// (FUSB303/D, Table 6).
var CurrentMode = map[byte]string{
	0x00: "0.0 A",
	0x01: "0.5 A",
	0x02: "1.5 A",
	0x03: "3.0 A",
}

// Get device identifier, reading I2C data address 0x01
// (DEVICE ID, (FUSB303/D, Table 13).
func GetDeviceID() (id []byte, err error) {
	return armoryctl.I2CRead(I2CBus, I2CAddress, 0x01, 1)
}

// Get detected current advertisement, reading I2C data address 0x11 (STATUS,
// (FUSB303/D, Table 22) and extracting value BC_LVL[1:0].
func GetCurrentMode() (mode byte, err error) {
	val, err := armoryctl.I2CRead(I2CBus, I2CAddress, 0x11, 1)

	if err != nil {
		return
	}

	mode = (val[0] >> 1) & 3

	return
}

// Force enable, writing I2C data address 0x05
// (CONTROL_1, FUSB303/D - Table 17).
func Enable() (err error) {
	return armoryctl.I2CWrite(I2CBus, I2CAddress, 0x05, []byte{0xbb})
}

// Force disable, writing I2C data address 0x05
// (CONTROL_1, FUSB303/D - Table 17).
func Disable() (err error) {
	return armoryctl.I2CWrite(I2CBus, I2CAddress, 0x05, []byte{0xb3})
}
