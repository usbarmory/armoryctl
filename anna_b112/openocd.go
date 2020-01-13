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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/f-secure-foundry/armoryctl/internal"
)

// Directory to optionally save backed up and written firmware images.
var CachePath = ""

// OpenOCD binary path, if the path contains a slash it is tried directly,
// otherwise it is searched in the directories named by the PATH environment
// variable.
var OpenOCDPath = "openocd"

const updateConfig = "ANNA-B112-CF-*.json"
const updateManifest = "manifest.json"

// These are fixed offsets that are not extracted from the update
// configuration.
const bootloaderOffset = 0x78000
const softDeviceCRCOffset = 0x7e000
const flashSize = 512 * 1024

const readFlashTemplate = `
init;
flash read_bank nrf52.flash _FLASH_FILE_;
flash read_bank nrf52.uicr _UICR_FILE_;
exit
`

const writeFlashTemplate = `
init;
reset halt;
nrf51 mass_erase;
flash write_bank nrf52.uicr _UICR_FILE_;
flash write_bank nrf52.flash _FLASH_FILE_;
reset;
exit
`

const interfaceTemplate = `
interface imx_gpio
imx_gpio_peripheral_base 0x0209c000
imx_gpio_swd_nums 4 6
`

const transportTemplate = `
transport select swd
source [find target/nrf52.cfg]
`

// JSON tag for bootloader entry in firmware metadata
const bootloaderTag = "Bootloader"

// JSON tag for connectivity software entry in firmware metadata
const connectivitySoftwareTag = "ConnectivitySoftware"

// JSON tag for Nordic application software entry in firmware metadata
const softDeviceTag = "SoftDevice"

type configurationEntry struct {
	Label       string
	Description string
	File        string
	Version     string
	Address     string
	Size        string
	CRC32       string
}

type manifestEntry struct {
	Manifest struct {
		Bootloader struct {
			BinFile        string `json:"bin_file"`
			DatFile        string `json:"dat_file"`
			InitPacketData struct {
				ApplicationVersion int64 `json:"application_version"`
				DeviceRevision     int   `json:"device_revision"`
				DeviceType         int   `json:"device_type"`
				FirmwareCRC16      int   `json:"firmware_crc16"`
				SoftdeviceReq      []int `json:"softdevice_req"`
			} `json:"init_packet_data"`
		} `json:"bootloader"`
		DFUVersion float64 `json:"dfu_version"`
	} `json:"manifest"`
}

func getBootloader(path string, archive string) (bootloader []byte, err error) {
	err = armoryctl.UnzipFile(filepath.Join(path, archive), path)

	if err != nil {
		return
	}

	j, err := ioutil.ReadFile(filepath.Join(path, updateManifest))

	if err != nil {
		return
	}

	var m manifestEntry

	if json.Valid([]byte(j)) {
		err = json.Unmarshal(j, &m)
	} else {
		err = fmt.Errorf("invalid JSON in bootloader manifest file")
	}

	if err != nil {
		return
	}

	return ioutil.ReadFile(filepath.Join(path, m.Manifest.Bootloader.BinFile))
}

func getConfig(path string) (config []byte, err error) {
	configFile, err := filepath.Glob(filepath.Join(path, updateConfig))

	if err != nil {
		return
	}

	if len(configFile) == 0 {
		err = fmt.Errorf("JSON configuration file not found")
		return
	}

	return ioutil.ReadFile(configFile[0])
}

func prepareImage(path string, output string) (err error) {
	path = filepath.Join(path, "uart")

	var c []configurationEntry
	var flash [flashSize]byte

	for i := range flash {
		flash[i] = 0xff
	}

	j, err := getConfig(path)

	if err != nil {
		return
	}

	// The update configuration file in u-connect v2.0.0-065 boundle
	// includes invalid JSON, we detect and patch it.
	invalidJSONPattern := regexp.MustCompile(`("Version": ".+"),(\r\s+})`)
	j = []byte(invalidJSONPattern.ReplaceAllString(string(j), "$1$2"))

	if json.Valid(j) {
		err = json.Unmarshal(j, &c)
	} else {
		err = fmt.Errorf("invalid JSON in configuration file")
	}

	if err != nil {
		return
	}

	config := make(map[string]configurationEntry)

	for _, tag := range c {
		config[tag.Label] = tag
	}

	requiredTags := [...]string{bootloaderTag, connectivitySoftwareTag, softDeviceTag}

	for _, tag := range requiredTags {
		if tag, ok := config[tag]; !ok {
			return fmt.Errorf("invalid JSON, %s tag not found in update configuration", tag)
		}
	}

	tag := config[bootloaderTag]
	bootloader, err := getBootloader(path, tag.File)

	if err != nil {
		return
	}

	copy(flash[bootloaderOffset:], bootloader)

	tag = config[connectivitySoftwareTag]
	connectivitySoftware, err := ioutil.ReadFile(filepath.Join(path, tag.File))

	if err != nil {
		return
	}

	addr, err := strconv.ParseInt(tag.Address, 0, 32)

	if err != nil {
		return
	}

	copy(flash[addr:], connectivitySoftware)

	tag = config[softDeviceTag]
	softDevice, err := ioutil.ReadFile(filepath.Join(path, tag.File))

	if err != nil {
		return
	}

	addr, err = strconv.ParseInt(tag.Address, 0, 32)

	if err != nil {
		return
	}

	copy(flash[addr:], softDevice)

	crc, err := strconv.ParseInt(tag.CRC32, 0, 32)

	if err != nil {
		return
	}

	crc32 := make([]byte, 4)
	binary.LittleEndian.PutUint32(crc32, uint32(crc))

	copy(flash[softDeviceCRCOffset:], crc32)

	size, err := strconv.ParseInt(tag.Size, 0, 32)

	if err != nil {
		return
	}

	s := make([]byte, 4)
	binary.LittleEndian.PutUint32(s, uint32(size))

	copy(flash[softDeviceCRCOffset+4:], s)

	return ioutil.WriteFile(output, flash[:], 0644)
}

func initCache() (cachePath string, err error) {
	cachePath = CachePath

	if cachePath == "" {
		cachePath, err = ioutil.TempDir("", "openocd-")
	} else {
		err = os.MkdirAll(cachePath, os.ModePerm)
	}

	return
}

func initOpenOCD() (openocd string, interfacePath string, transportPath string, tmpDir string, err error) {
	openocd, err = exec.LookPath(OpenOCDPath)

	if err != nil {
		return
	}

	tmpDir, err = ioutil.TempDir("", "openocd-")

	if err != nil {
		return
	}

	interfacePath = filepath.Join(tmpDir, "interface.cfg")
	transportPath = filepath.Join(tmpDir, "transport.cfg")

	err = ioutil.WriteFile(interfacePath, []byte(interfaceTemplate), 0644)

	if err != nil {
		return
	}

	err = ioutil.WriteFile(transportPath, []byte(transportTemplate), 0644)

	return
}

// Overwrites CUSTOMER[31] register to achieve the effects of
// `AT+UPRODLFCLK=0,16,2`.
func patchUICR(inputPath string) (outputPath string, err error) {
	uicr, err := ioutil.ReadFile(inputPath)

	if err != nil {
		return
	}

	outputPath = inputPath + "-patched"

	copy(uicr[len(uicr)-4:], []byte{0x80, 0x04, 0x00, 0x80})
	err = ioutil.WriteFile(outputPath, uicr[:], 0644)

	return
}

func execOpenOCD(flashPath string, uicrPath string, template string, retry bool) (err error) {
	openocd, interfacePath, transportPath, tmpDir, err := initOpenOCD()

	if err != nil {
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }() // make errcheck happy

	cmd := strings.ReplaceAll(template, "_FLASH_FILE_", flashPath)
	cmd = strings.ReplaceAll(cmd, "_UICR_FILE_", uicrPath)

	args := []string{"-f", interfacePath, "-f", transportPath, "-c", cmd}

	_, err = armoryctl.ExecCommand(openocd, args, true, "")

	if retry && err != nil {
		// The first time the flash command is executed after a mass_erase it
		// can fail, we ignore any error at first attempt and re-execute the
		// same command a second time.
		_, err = armoryctl.ExecCommand(openocd, args, true, "")
	}

	return
}

// Backup the nrf52.flash and nrf52.uicr regions.
func Backup() (flashPath string, uicrPath string, err error) {
	cachePath, err := initCache()

	if err != nil {
		return
	}

	time := time.Now().Unix()
	flashPath = filepath.Join(cachePath, fmt.Sprintf("flash-%d.bin", time))
	uicrPath = filepath.Join(cachePath, fmt.Sprintf("UICR-%d.bin", time))

	err = execOpenOCD(flashPath, uicrPath, readFlashTemplate, false)

	return
}

// Write the nrf52.flash and nrf52.uicr regions. *IMPORTANT*: the `mass_erase`
// command is issued before starting the procedure, all data contained in the
// flash and uicr regions (including module MAC address and/or any other
// configuration) will be lost.
func Flash(flashPath string, uicrPath string) (err error) {
	return execOpenOCD(flashPath, uicrPath, writeFlashTemplate, true)
}

// Update the ANNA-B112 firmware, a backup of the current flash and UICR region
// is created before overwriting them. The update also performs the operation
// described in FlashSetInternalRCLFCK.
func Update(updateFile string) (err error) {
	tmpDir, err := ioutil.TempDir("", "update-")

	if err != nil {
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }() // make errcheck happy

	cachePath, err := initCache()

	if err != nil {
		return
	}

	err = armoryctl.UnzipFile(updateFile, tmpDir)

	if err != nil {
		return
	}

	flash := filepath.Join(cachePath, "flash.bin")

	err = prepareImage(tmpDir, flash)

	if err != nil {
		return
	}

	_, uicr, err := Backup()

	if err != nil {
		return
	}

	uicr, err = patchUICR(uicr)

	if err != nil {
		return
	}

	return Flash(flash, uicr)
}

// Set the low frequency clock source to the internal RC with default
// parameters recommended by Nordic SDK, by modifying the relevant area in the
// module flash.
// (see nRF5_SDK_15.3.0_59ac345/components/softdevice/s132/headers/nrf_sdm.h).
func FlashSetInternalRCLFCK() (err error) {
	flash, uicr, err := Backup()

	if err != nil {
		return
	}

	uicr, err = patchUICR(uicr)

	if err != nil {
		return
	}

	return Flash(flash, uicr)
}
