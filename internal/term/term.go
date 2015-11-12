// Based on ssh/terminal:
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux,!appengine darwin freebsd openbsd

package term

type fder interface {
	Fd() uintptr
}
