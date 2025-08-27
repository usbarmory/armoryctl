USB armory Mk II - hardware control tool
========================================

armoryctl | https://github.com/usbarmory/armoryctl  
Copyright (c) The armoryctl authors. All Rights Reserved.

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

Daniele Bianco  
daniele@inversepath.com  

Introduction
============

The `armoryctl` tool provides user space support for communicating with the
[USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki) internal
peripherals.

The functions leveraged by the tool are all exported to allow use of the
package as a library.

The package simplifies communication with the following on-board USB armory Mk
II components:

* GPIOs
  - white LED
  - blue LED

* I²C slaves
  - Type-C plug port controller (TUSB320)
  - Type-C receptacle port controller (FUSB303)
  - Secure Element (ATECC608A/ATECC608B)
  - Power Management Integrated Circuit (PF1510)

* UART
  - Bluetooth module (ANNA-B112)

Warning
=======

This tool is only meant to be used on USB armory Mk II hardware, execution on
any other hardware is unsupported and can lead to irreversible damage.

The use of this tool is therefore **at your own risk**.

Requirements
============

The `ble rc_lfck flash` and `ble update` commands require the `openocd` tool
compiled with the `-enable-imx_gpio` flag ([compilation instructions](https://github.com/usbarmory/usbarmory/wiki/Bluetooth#using-openocd-for-jtag-access)).

Operation
=========

```
Usage: armoryctl [options] [command]
  -c string
    	ANNA-B112 firmware cache path (default "~/.armoryctl")
  -d	debug
  -f	skip hardware check and force execution
  -i int
    	ATECC608 I2C bus number
  -l int
    	ATECC608 I2C address (default 96)
  -m int
    	FUSB303 I2C bus number
  -n int
    	FUSB303 I2C address (default 49)
  -o int
    	TUSB320 I2C bus number
  -p int
    	TUSB320 I2C address (default 97)
  -q int
    	PF1510 I2C bus number
  -r int
    	PF1510 I2C address (default 8)
  -s int
    	ANNA-B112 UART speed (default 115200)
  -u string
    	ANNA-B112 UART path (default "/dev/ttymxc0")
  -x string
    	OpenOCD lookpath (default "openocd")

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
```

Installing
==========

You can automatically download, compile and install the package, under your
GOPATH, as follows:

```
go install github.com/usbarmory/armoryctl@latest
```

Alternatively you can manually compile it from source:

```
git clone https://github.com/usbarmory/armoryctl
cd armoryctl && make
```

The tool can be cross compiled for an ARM target as follows:

```
make armoryctl GOARCH=arm
```

The default compilation target automatically runs all available unit tests.

License
=======

armoryctl | https://github.com/usbarmory/armoryctl  
Copyright (c) The armoryctl authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/armoryctl/blob/master/LICENSE) file.
