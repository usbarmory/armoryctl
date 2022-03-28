// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// Links:
//   https://www.nxp.com/docs/en/data-sheet/PF1510.pdf
//   https://github.com/usbarmory/usbarmory/wiki/I%C2%B2C-(Mk-II)

// Package pf1510 supports communication with the NXP PF1510 PMIC.
package pf1510

import (
	"fmt"

	"github.com/usbarmory/armoryctl/internal"
)

var (
	I2CBus     = 0
	I2CAddress = 0x08
)

// DEVICE_ID (p53, Table 52, PF1510 Datasheet).
var DeviceID = map[byte]string{
	0x4: "PF1510",
}

// FAMILY (p53, Table 52, PF1510 Datasheet).
var Family = map[byte]string{
	15: "15",
}

// Get device identifier and chip family reading I2C data address
// 0x00: device_id <0:2>, family <3:7>
func Info() (res string, err error) {
	// Register DEVICE_ID - ADDR 0x00
	// (p53, Table 52, PF1510 Datasheet).
	val, err := armoryctl.I2CRead(I2CBus, I2CAddress, 0x00, 1)

	if err != nil {
		return
	}

	id := val[0] & 0x07
	family := (val[0] & 0xf8) >> 3

	// Register OTP_FLAVOR - ADDR 0x01
	// (p53, Table 52, PF1510 Datasheet).
	otp, err := armoryctl.I2CRead(I2CBus, I2CAddress, 0x01, 1)

	if err != nil {
		return
	}

	// Register SILICON_REV - ADDR 0x02
	// (p53, Table 54, PF1510 Datasheet).
	rev, err := armoryctl.I2CRead(I2CBus, I2CAddress, 0x02, 1)

	if err != nil {
		return
	}

	res = fmt.Sprintf(`id:%#x("%s") family:%#x("%s") otp:"A%d" rev:%#x`, id, DeviceID[id], family, Family[family], otp[0], rev[0])

	return
}
