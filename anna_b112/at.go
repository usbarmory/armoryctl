// armoryctl | https://github.com/f-secure-foundry/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// Links:
//   https://www.u-blox.com/sites/default/files/ANNA-B112_DataSheet_%28UBX-18011707%29.pdf
//   https://github.com/f-secure-foundry/usbarmory/wiki/Bluetooth

package anna_b112

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/f-secure-foundry/armoryctl/internal"
)

// Serial device path
var UARTPath = "/dev/ttymxc0"

// Serial device speed
var UARTSpeed = 115200

var responseStringPattern = regexp.MustCompile(`"[^"]+"|OK`)

func sendATCmd(cmd string) (response string, err error) {
	response, err = armoryctl.UARTWrite(UARTPath, UARTSpeed, "AT"+cmd+"\r")

	if err != nil {
		return
	}

	m := responseStringPattern.FindStringSubmatch(response)

	if len(m) == 1 {
		response = m[0]
	} else {
		err = errors.New("response error")
	}

	return response, err
}

// Get device manufacturer (AT+CGMI).
func GetDeviceManufacturer() (manufacturer string, err error) {
	return sendATCmd("+CGMI")
}

// Get device model (AT+CGMM).
func GetDeviceModel() (model string, err error) {
	return sendATCmd("+CGMM")
}

// Get product serial number (AT+CGSN).
func GetDeviceSerial() (model string, err error) {
	return sendATCmd("+CGSN")
}

// Get software version (AT+CGMR).
func GetSoftwareVersion() (version string, err error) {
	return sendATCmd("+CGMR")
}

// Get device name (AT+UBTLN?).
func GetDeviceName() (name string, err error) {
	return sendATCmd("+UBTLN?")
}

// Set Bluetooth device name (AT+UBTLN="device name") and parmanently store
// the current configuration (AT+&W, AT+CPWROFF).
func SetDeviceName(name string) (err error) {
	cmds := [3]string{"+UBTLN=\""+name+"\"", "&W", "+CPWROFF"}

	for _, cmd := range cmds {
		_, err = sendATCmd(cmd)

		if err != nil {
			return
		}
	}

	return
}

// Assemble the device identification string from device manufacturer, model,
// product serial, software version and Bluetooth device name.
func Info() (id string, err error) {
	manufacturer, err := GetDeviceManufacturer()

	if err != nil {
		manufacturer = fmt.Sprintf("error:(%s)", err)
	}

	model, err := GetDeviceModel()

	if err != nil {
		model = fmt.Sprintf("error:(%s)", err)
	}

	serial, err := GetDeviceSerial()

	if err != nil {
		serial = fmt.Sprintf("error:(%s)", err)
	}

	version, err := GetSoftwareVersion()

	if err != nil {
		version = fmt.Sprintf("error:(%s)", err)
	}

	name, err := GetDeviceName()

	if err != nil {
		name = fmt.Sprintf("error:(%s)", err)
	}

	id = "manufacturer:" + manufacturer + " model:" + model + " serial:" + serial + " sw:" + version + " device_name:" + name

	return
}

// Reset the BLE module by toggling the RESET_N pin (GPIO 9).
func Reset() (err error) {
	err = armoryctl.GPIOSetOutput("GPIO9", false)

	if err != nil {
		return
	}

	// grace time to ensure reset triggering
	time.Sleep(1 * time.Second)

	return armoryctl.GPIOSetOutput("GPIO9", true)
}

// Toggle BLE visibility to non discoverable (AT+UBTDM=1), non pairable
// (AT+UBTPM=1), non connectable (AT+UBTCM=1) and disable any BLE role
// (AT+UBTLE=0), finally permanently store current configuration (AT&W, AT+CPWROFF).
func Disable() (err error) {
	cmds := [6]string{"+UBTDM=1", "+UBTPM=1", "+UBTCM=1", "+UBTLE=0", "&W", "+CPWROFF"}

	for _, cmd := range cmds {
		_, err = sendATCmd(cmd)

		if err != nil {
			return
		}
	}

	return
}

// Toggle BLE visibility to always discoverable (AT+UBTDM=3), pairable
// (AT+UBTPM=2), connectable (AT+UBTCM=2) and set BLE role to peripheral
// (AT+UBTLE=2), finally permanently store current configuration (AT&W, AT+CPWROFF).
func Enable() (err error) {
	cmds := [6]string{"+UBTDM=3", "+UBTPM=2", "+UBTCM=2", "+UBTLE=2", "&W", "+CPWROFF"}

	for _, cmd := range cmds {
		_, err = sendATCmd(cmd)

		if err != nil {
			return
		}
	}

	return
}

// Enter bootloader mode by driving low SWITCH_1 (GPIO 27) and
// SWITCH_2 (GPIO 26) during a module reset cycle.
func EnterBootloaderMode() (err error) {
	err = armoryctl.GPIOSetOutput("GPIO26", false)

	if err != nil {
		return
	}

	err = armoryctl.GPIOSetOutput("GPIO27", false)

	if err != nil {
		return
	}

	return Reset()
}

// Enter normal mode by driving high SWITCH_1 (GPIO 27) and
// SWITCH_2 (GPIO 26) during a module reset cycle.
func EnterNormalMode() (err error) {
	err = armoryctl.GPIOSetOutput("GPIO26", true)

	if err != nil {
		return
	}

	err = armoryctl.GPIOSetOutput("GPIO27", true)

	if err != nil {
		return
	}

	return Reset()
}

// Set the low frequency clock source to the internal RC with default
// parameters recommended by Nordic SDK, using the AT+UPRODLFCLK command.
// (see nRF5_SDK_15.3.0_59ac345/components/softdevice/s132/headers/nrf_sdm.h).
func ATSetInternalRCLFCK() (err error) {
	_, err = sendATCmd("+UPROD=1")

	if err != nil {
		return
	}

	_, err = sendATCmd("+UPRODLFCLK=0,16,2")

	return
}
