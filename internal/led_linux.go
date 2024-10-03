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
	"strings"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/sysfs"
)

func LED(name string, on bool) (err error) {
	_, err = host.Init()

	if err != nil {
		return
	}

	led, err := sysfs.LEDByName("LED_" + strings.ToUpper(name))

	if err != nil {
		return
	}

	if on {
		err = led.Out(gpio.High)
	} else {
		err = led.Out(gpio.Low)
	}

	return
}
