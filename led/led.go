// armoryctl | https://github.com/f-secure-foundry/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package led provides control for the USB armory Mk II LEDs.
package led

import (
	"github.com/f-secure-foundry/armoryctl/internal"
)

// Turn on/off LED by name.
func Set(name string, on bool) (err error) {
	return armoryctl.LED(name, on)
}
