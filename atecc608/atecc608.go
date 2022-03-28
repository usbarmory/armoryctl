// armoryctl | https://github.com/usbarmory/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// Links:
//   http://ww1.microchip.com/downloads/en/DeviceDoc/ATECC608A-CryptoAuthentication-Device-Summary-Data-Sheet-DS40001977B.pdf
//   https://github.com/usbarmory/usbarmory/wiki/I%C2%B2C-(Mk-II)

// Package atecc608 supports communication with Microchip ATECC608A and
// ATECC608B secure elements.
package atecc608

import (
	"bytes"
	"fmt"
	"time"

	"github.com/usbarmory/armoryctl/internal"
)

var (
	I2CBus     = 0
	I2CAddress = 0x60
)

const (
	CmdAddress        = 0x03
	CRC16Poly  uint16 = 0x8005
)

// Max command execution time in ms considering a Clock-Divider set to the
// default/recommended value of 0x00,
// (p66, Table 10-5, ATECC608A Full Datasheet).
const CmdMaxExecutionTime = (200 + 50) * time.Millisecond

// CmdExecutionTime sets the wait time between command execution and result
// retrieval.
var CmdExecutionTime = CmdMaxExecutionTime

// Minimum required cmd fields:
//   count (1) + op (1) + param1 (1) + param2 (2) + crc16 (2).
const cmdMinLen = 7

// Minimum required response fields:
//   count (1) + data (1) + crc16 (2).
const responseMinLen = 4

// Cmd represents the list of supported command codes,
// (p65, 10.4.1. Command Summary, ATECC608A Full Datasheet).
var Cmd = map[string]byte{
	"AES":         0x51,
	"CheckMac":    0x28,
	"Counter":     0x24,
	"DeriveKey":   0x1C,
	"ECDH":        0x43,
	"GenDig":      0x15,
	"GenKey":      0x40,
	"Info":        0x30,
	"KDF":         0x56,
	"Lock":        0x17,
	"MAC":         0x08,
	"Nonce":       0x16,
	"PrivWrite":   0x46,
	"Random":      0x1B,
	"Read":        0x02,
	"SecureBoot":  0x80,
	"SelfTest":    0x77,
	"Sign":        0x41,
	"SHA":         0x47,
	"UpdateExtra": 0x20,
	"Verify":      0x45,
	"Write":       0x12,
}

// Status represents the device status/error codes,
// (p64-65, Tab 10-3, ATECC608A Full Datasheet).
var Status = map[byte]string{
	0x00: "successful command execution",
	0x01: "checkmac or verify miscompare",
	0x03: "parse error",
	0x05: "ECC fault",
	0x07: "self test error",
	0x08: "health test error",
	0x0f: "execution error",
	0x11: "after wake, prior to first command",
	0xee: "watchdog about to expire",
	0xff: "CRC or other communications error",
}

// Supported tests and result bit mask,
// (p100, Table 11-43, ATECC608A Full Datasheet).
var testMask = map[string]byte{
	"RNG":   0x20,
	"AES":   0x10,
	"ECDH":  0x08,
	"ECDSA": 0x02,
	"DRBG":  0x01,
}

func crc16(data []byte) []byte {
	var crc uint16

	for i := 0; i < len(data); i++ {
		for shift := uint8(0x01); shift > 0x00; shift <<= 1 {
			// data and crc bits
			var d uint8
			var c uint8

			if uint8(data[i])&uint8(shift) != 0 {
				d = 1
			}

			c = uint8(crc >> 15)
			crc <<= 1

			if d != c {
				crc ^= CRC16Poly
			}
		}
	}

	return []byte{byte(crc & 0xff), byte(crc >> 8)}
}

func verifyResponse(res []byte) (data []byte, err error) {
	// ATECC response packet format:
	//   count [1] | status/error/response data[variable] | crc16 [2]
	//
	// (p63, Table 10-1, ATECC608A Full Datasheet)
	if len(res) < responseMinLen {
		err = fmt.Errorf("invalid response, got less than %d bytes", responseMinLen)
		return
	}

	size := len(res) - 2

	count := res[0]
	payload := res[:size]
	data = res[1:size]
	crc := res[size:]

	if !bytes.Equal(crc16(payload), crc) {
		err = fmt.Errorf("checksum verification failure")
		return
	}

	// A response with 4 bytes must contain a valid status/error code,
	// otherwise data is being transferred.
	if count > responseMinLen {
		return
	}

	status := data[0]

	if Status[status] == "" {
		err = fmt.Errorf("invalid status/error code: %x", status)
	} else if status != 0x00 && (status <= 0x0f || status == 0xff) {
		err = fmt.Errorf("%s", Status[status])
	}

	return
}

// Wake issues a device wake-up which is always needed before starting a
// new command session.
func Wake() (err error) {
	// Any error at the very first I2CWrite() is silently ignored as
	// the device always returns a "Write Error" here.
	//
	// Writing 0x00 triggers the chip wake-up
	// (p47, 7.1 I/O Conditions, ATECC608A Full Datasheet).
	_ = armoryctl.I2CWrite(I2CBus, I2CAddress, 0x00, []byte{0x00})

	// Wait tWHI
	// (p56, 9.3 AC Parameters: All I/O Interfaces, ATECC608A Full Datasheet).
	time.Sleep(1500 * time.Microsecond)

	// It is necessary to read 4 bytes of data to verify that the chip
	// wake-up has been successful.
	res, err := armoryctl.I2CRead(I2CBus, I2CAddress, 0x00, 4)

	if err != nil {
		return
	}

	data, err := verifyResponse(res)

	if err != nil && data[0] != 0x11 {
		err = fmt.Errorf("wake-up failed")
	}

	return
}

// Idle puts the device in idle mode,
// (p50, Table 7-2, ATECC608A Full Datasheet).
func Idle() {
	_ = armoryctl.I2CWrite(I2CBus, I2CAddress, 0x02, nil)
}

// Sleep puts the device in sleep mode,
// (p50, Table 7-2, ATECC608A Full Datasheet).
func Sleep() {
	_ = armoryctl.I2CWrite(I2CBus, I2CAddress, 0x01, nil)
}

// ExecuteCmd issues an ATECC command conforming to:
//   * p55, Table  9-1, ATECC508A Full Datasheet
//   * p63, Table 10-1, ATECC608A Full Datasheet
//
// The wake flag results in the executed command to be issued individually
// within a Wake() and Idle() cycle, when the flag is false the caller must
// take care of waking/idling/sleeping according to its desired command
// sequence.
func ExecuteCmd(opcode byte, param1 [1]byte, param2 [2]byte, data []byte, wake bool) (res []byte, err error) {
	if wake {
		if err = Wake(); err != nil {
			return
		}

		defer Idle()
	}

	// ATECC cmd packet format:
	//   count [1] | cmd fields [variable] | crc16 [2]
	//
	// ATECC cmd format:
	//   opcode [1] | param1 [1] | param2 [2] | data [variable]
	//
	// (p63, Table 10-1, ATECC608A Full Datasheet)
	var pkt []byte

	count := []byte{byte(cmdMinLen + len(data))}

	pkt = append(pkt, count...)
	pkt = append(pkt, opcode)
	pkt = append(pkt, param1[:]...)
	pkt = append(pkt, param2[:]...)
	pkt = append(pkt, data...)
	pkt = append(pkt, crc16(pkt)...)

	if err = armoryctl.I2CWrite(I2CBus, I2CAddress, CmdAddress, pkt); err != nil {
		return
	}

	time.Sleep(CmdExecutionTime)

	// The output FIFO is shared among status, error, and command results.
	// The first read command is needed to read how many bytes are present
	// in the output buffer.
	//
	// (p64, 10.3 Status/Error Codes, ATECC608A Full Datasheet)
	resCount, err := armoryctl.I2CRead(I2CBus, I2CAddress, CmdAddress, 1)

	if err != nil {
		return
	}

	// The second read command gets the rest of the response from the
	// output buffer.
	res, err = armoryctl.I2CRead(I2CBus, I2CAddress, CmdAddress, uint(resCount[0]))

	if err != nil {
		return
	}

	return verifyResponse(res)
}

// SelfTest executes the self test command and returns its results.
func SelfTest() (res string, err error) {
	// param1 0x3b: performs all available tests.
	data, err := ExecuteCmd(Cmd["SelfTest"], [1]byte{0x3b}, [2]byte{0x00, 0x00}, nil, true)

	if err != nil {
		return
	}

	for k, v := range testMask {
		if data[0]&v != 0x00 {
			res += k + ":FAIL "
		} else {
			res += k + ":PASS "
		}
	}

	return
}

// Info executes the info command and returns the device serial number and
// software revision.
func Info() (res string, err error) {
	// param1 0x80: reads 32 bytes configuration region
	// param2 0x0000: represents the start address
	data, err := ExecuteCmd(Cmd["Read"], [1]byte{0x80}, [2]byte{0x00, 0x00}, nil, true)

	if err != nil {
		return
	}

	// The first 32 bytes in the configuration region will contain:
	// 72 bits - unique serial number: bytes <0:3> and <8:12>
	// 4 bytes - device revision number: bytes <4:7>.
	serial := []byte{}
	serial = append(serial, data[0:4]...)
	serial = append(serial, data[8:13]...)
	revision := data[4:8]

	return fmt.Sprintf("serial:0x%x revision:0x%x", serial, revision), nil
}
