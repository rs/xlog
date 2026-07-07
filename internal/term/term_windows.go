// Based on ssh/terminal:
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows
// +build windows

package term

import (
	"io"
)

// IsTerminal returns true if w writes to a terminal.
func IsTerminal(w io.Writer) bool {
	return false
}
