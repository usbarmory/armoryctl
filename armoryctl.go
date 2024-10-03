// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build linux

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/usbarmory/armoryctl/anna_b112"
	"github.com/usbarmory/armoryctl/atecc608"
	"github.com/usbarmory/armoryctl/fusb303"
	"github.com/usbarmory/armoryctl/internal"
	"github.com/usbarmory/armoryctl/led"
	"github.com/usbarmory/armoryctl/pf1510"
	"github.com/usbarmory/armoryctl/tusb320"
)

type Config struct {
	debug  bool
	force  bool
	logger *log.Logger
}

var conf *Config

const warning = `
████████████████████████████████████████████████████████████████████████████████
                                **  WARNING  **

This tool is only meant to be used on USB armory Mk II hardware, execution on
any other hardware is unsupported and can lead to irreversible damage.

The use of this tool is therefore **at your own risk**.
████████████████████████████████████████████████████████████████████████████████
`

const commandUsage = `
LED control
  led (white|blue) (on|off)

Type-C plug port controller (TUSB320)
  tusb id			# read controller identifier
  tusb current_mode		# read advertised current

Type-C receptacle port controller (FUSB303)
  fusb id			# read controller identifier
  fusb current_mode		# read advertised current
  fusb enable			# enable the controller
  fusb disable			# disable the controller

Bluetooth module (ANNA-B112)
  ble info			# read device information
  ble enable			# set visible peripheral BLE role
  ble disable			# disable BLE communication
  ble reset			# reset the module
  ble bootloader_mode		# switch to bootloader mode
  ble normal_mode		# switch to normal operation
  ble rc_lfck (flash|at)	# set LF clock source to internal RC oscillator
  ble update <firmware path>	# module firmware update
  ble name <device name>	# set device name

Secure Element (ATECC608A/ATECC608B)
  atecc info			# read device information
  atecc self_test		# execute self test procedure

Power Management Integrated Circuit (PF1510)
  pmic info			# read device information
`

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	conf = &Config{
		logger: log.New(os.Stdout, log.Prefix(), log.Flags()),
	}

	cachePath := ""
	usr, err := user.Current()

	if err == nil {
		cachePath = filepath.Join(usr.HomeDir, ".armoryctl")
	}

	flag.Usage = func() {
		tags := ""

		if armoryctl.Revision != "" && armoryctl.Build != "" {
			tags = fmt.Sprintf("%s (%s)", armoryctl.Revision, armoryctl.Build)
		}

		log.Printf("USB armory Mk II hardware control tool\n%s\n", tags)
		log.Print(warning)
		log.Printf("Usage: armoryctl [options] [command]\n")
		flag.PrintDefaults()
		log.Print(commandUsage)
	}

	flag.BoolVar(&conf.debug, "d", false, "debug")
	flag.BoolVar(&conf.force, "f", false, "skip hardware check and force execution")

	flag.StringVar(&anna_b112.CachePath, "c", cachePath, "ANNA-B112 firmware cache path")
	flag.StringVar(&anna_b112.OpenOCDPath, "x", anna_b112.OpenOCDPath, "OpenOCD lookpath")
	flag.StringVar(&anna_b112.UARTPath, "u", anna_b112.UARTPath, "ANNA-B112 UART path")
	flag.IntVar(&anna_b112.UARTSpeed, "s", anna_b112.UARTSpeed, "ANNA-B112 UART speed")

	flag.IntVar(&atecc608.I2CBus, "i", atecc608.I2CBus, "ATECC608 I2C bus number")
	flag.IntVar(&atecc608.I2CAddress, "l", atecc608.I2CAddress, "ATECC608 I2C address")

	flag.IntVar(&fusb303.I2CBus, "m", fusb303.I2CBus, "FUSB303 I2C bus number")
	flag.IntVar(&fusb303.I2CAddress, "n", fusb303.I2CAddress, "FUSB303 I2C address")

	flag.IntVar(&tusb320.I2CBus, "o", tusb320.I2CBus, "TUSB320 I2C bus number")
	flag.IntVar(&tusb320.I2CAddress, "p", tusb320.I2CAddress, "TUSB320 I2C address")

	flag.IntVar(&pf1510.I2CBus, "q", pf1510.I2CBus, "PF1510 I2C bus number")
	flag.IntVar(&pf1510.I2CAddress, "r", pf1510.I2CAddress, "PF1510 I2C address")
}

func checkModel() (match bool) {
	model, _ := os.ReadFile("/proc/device-tree/model")
	match, _ = regexp.Match("F-Secure USB armory Mk II", model)

	return
}

func invalid() {
	flag.Usage()
	log.Fatalf("error: invalid command given")
}

func main() {
	var err error
	var res string

	defer func() {
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if len(res) != 0 {
			log.Printf("%s", res)
		}
	}()

	flag.Parse()

	if len(flag.Args()) < 2 {
		flag.Usage()
		err = errors.New("no valid command given")
		return
	}

	if !conf.force && !checkModel() {
		err = errors.New("this tool is only meant to be used on USB armory Mk II hardware")
		return
	}

	if conf.debug {
		armoryctl.Logger = conf.logger
	}

	device := flag.Arg(0)
	command := flag.Arg(1)

	op := fmt.Sprintf("%s %s", device, command)

	switch op {
	case "led white", "led blue":
		if len(flag.Args()) < 3 {
			invalid()
		}

		switch flag.Arg(2) {
		case "on":
			err = led.Set(command, true)
		case "off":
			err = led.Set(command, false)
		default:
			invalid()
		}
	case "tusb id":
		var id []byte
		id, err = tusb320.GetDeviceID()

		if err == nil {
			res = string(id)
		}
	case "fusb id":
		var id []byte
		id, err = fusb303.GetDeviceID()

		if err == nil {
			res = fmt.Sprintf("0x%x", id)
		}
	case "tusb current_mode":
		var mode byte
		mode, err = tusb320.GetCurrentMode()

		if err == nil {
			res = tusb320.CurrentMode[mode]
		}
	case "fusb current_mode":
		var mode byte
		mode, err = fusb303.GetCurrentMode()

		if err == nil {
			res = fusb303.CurrentMode[mode]
		}
	case "fusb enable":
		err = fusb303.Enable()
	case "fusb disable":
		err = fusb303.Disable()
	case "ble info":
		res, err = anna_b112.Info()
	case "ble enable":
		err = anna_b112.Enable()
	case "ble disable":
		err = anna_b112.Disable()
	case "ble reset":
		err = anna_b112.Reset()
	case "ble bootloader_mode":
		err = anna_b112.EnterBootloaderMode()
	case "ble normal_mode":
		err = anna_b112.EnterNormalMode()
	case "ble rc_lfck":
		if len(flag.Args()) < 3 {
			invalid()
		}

		switch flag.Arg(2) {
		case "at":
			err = anna_b112.ATSetInternalRCLFCK()
		case "flash":
			err = anna_b112.FlashSetInternalRCLFCK()
		default:
			invalid()
		}
	case "ble update":
		if len(flag.Args()) < 3 {
			invalid()
		}

		err = anna_b112.Update(flag.Arg(2))
	case "ble name":
		if len(flag.Args()) < 3 {
			invalid()
		}

		err = anna_b112.SetDeviceName(flag.Arg(2))
	case "atecc info":
		res, err = atecc608.Info()
	case "atecc self_test":
		res, err = atecc608.SelfTest()
	case "pmic info":
		res, err = pf1510.Info()
	default:
		invalid()
	}
}
