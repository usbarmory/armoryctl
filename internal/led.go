// armoryctl | https://github.com/f-secure-foundry/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package armoryctl

import (
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/sysfs"
)

func findLED(name string) (l *sysfs.LED, err error) {
	_, err = host.Init()

	if err != nil {
		return
	}

	return sysfs.LEDByName(name)
}

func LEDOn(name string) (err error) {
	l, err := findLED(name)

	if err != nil {
		return
	}

	return l.Out(gpio.High)
}

func LEDOff(name string) (err error) {
	l, err := findLED(name)

	if err != nil {
		return
	}

	return l.Out(gpio.Low)
}
