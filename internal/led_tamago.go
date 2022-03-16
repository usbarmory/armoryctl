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
	"github.com/usbarmory/tamago/board/f-secure/usbarmory/mark-two"
)

func LED(name string, on bool) (err error) {
	return usbarmory.LED(name, on)
}
