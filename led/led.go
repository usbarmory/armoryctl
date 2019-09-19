// armoryctl | https://github.com/inversepath/armoryctl
//
// USB armory Mk II - hardware control tool
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//

package led

import (
	"strings"

	"github.com/inversepath/armoryctl/internal"
)

// Turn on/off LED by name.
func Set(name string, on bool) (err error) {
	name = "LED_" + strings.ToUpper(name)

	if on {
		return armoryctl.LEDOn(name)
	} else {
		return armoryctl.LEDOff(name)
	}
}
