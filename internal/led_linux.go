// armoryctl | https://github.com/f-secure-foundry/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build linux

package armoryctl

import (
	"strings"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/sysfs"
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
