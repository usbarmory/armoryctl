// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) The armoryctl authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build tamago,arm

package armoryctl

import (
	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
)

func LED(name string, on bool) (err error) {
	return usbarmory.LED(name, on)
}
