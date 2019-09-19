// armoryctl | https://github.com/inversepath/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package armoryctl

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
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
