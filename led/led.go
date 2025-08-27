// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) The armoryctl authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package led provides control for the USB armory Mk II LEDs.
package led

import (
	"github.com/usbarmory/armoryctl/internal"
)

// Turn on/off LED by name.
func Set(name string, on bool) (err error) {
	return armoryctl.LED(name, on)
}
