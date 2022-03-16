// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build tamago,arm

package armoryctl

import (
	"fmt"

	"github.com/usbarmory/tamago/soc/imx6"
)

const I2CBus = 0

func init() {
	imx6.I2C1.Init()
}

func I2CRead(bus int, addr int, reg uint8, size uint) (val []byte, err error) {
	if bus != I2CBus {
		return nil, fmt.Errorf("I2C bus must be set to %d", I2CBus)
	}

	return imx6.I2C1.Read(uint8(addr), uint32(reg), 1, int(size))
}

func I2CWrite(bus int, addr int, reg uint8, val []byte) (err error) {
	if bus != I2CBus {
		return fmt.Errorf("I2C bus must be set to %d", I2CBus)
	}

	return imx6.I2C1.Write(val, uint8(addr), uint32(reg), 1)
}
