// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) The armoryctl authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build linux

package armoryctl

import (
	"fmt"
	"log"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

func findGPIO(name string) (pin gpio.PinIO, err error) {
	_, err = host.Init()

	if err != nil {
		return
	}

	pin = gpioreg.ByName(name)
	if pin == nil {
		err = fmt.Errorf("Failed to find gpio %s" + name)
	}

	return pin, err
}

// Configure a GPIO pin as output high or low.
func GPIOSetOutput(name string, high bool) (err error) {
	p, err := findGPIO(name)

	if err != nil {
		return
	}

	if Logger != nil {
		log.Printf("GPIO %s high:%v\n", name, high)
	}

	if high {
		err = p.Out(gpio.High)
	} else {
		err = p.Out(gpio.Low)
	}

	return
}
