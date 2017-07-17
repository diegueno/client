// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build !windows

package libkb

// #include<resolv.h>
import "C"

func resInit() {
	C.res_init()
}
