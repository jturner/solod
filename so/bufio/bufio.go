// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bufio implements buffered I/O. It wraps an io.Reader or io.Writer
// object, creating another object (Reader or Writer) that also implements
// the interface but provides buffering and some help for textual I/O.
//
// Based on the [bufio] package.
//
// [bufio]: https://github.com/golang/go/blob/go1.26.2/src/bufio
package bufio

import "solod.dev/so/errors"

// DefaultBufSize is the default buffer size used by [NewReader] and [NewWriter].
const DefaultBufSize = 4096

var (
	ErrInvalidUnreadByte = errors.New("bufio: invalid use of UnreadByte")
	ErrInvalidUnreadRune = errors.New("bufio: invalid use of UnreadRune")
	ErrBufferFull        = errors.New("bufio: buffer full")
	ErrNegativeCount     = errors.New("bufio: negative count")
)
