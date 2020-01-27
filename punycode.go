package zvalidate

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// Adapted from:
// https://github.com/golang/net/blob/c0dbc17a35534bf2e581d7a942408dc936316da4/idna/punycode.go#L34
//
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements the Punycode algorithm from RFC 3492.

// These parameter values are specified in section 5.
//
// All computation is done with int32s, so that overflow behavior is identical
// regardless of whether int is 32-bit or 64-bit.
const (
	base        int32 = 36
	damp        int32 = 700
	initialBias int32 = 72
	initialN    int32 = 128
	skew        int32 = 38
	tmax        int32 = 26
	tmin        int32 = 1
)

func punyError(s string) error { return &labelError{s, "A3"} }

type labelError struct{ label, code_ string }

func (e labelError) code() string { return e.code_ }
func (e labelError) Error() string {
	return fmt.Sprintf("idna: invalid label %q", e.label)
}

// decode decodes a string as specified in section 6.2.
func punyDecode(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	pos := 1 + strings.LastIndex(encoded, "-")
	if pos == 1 {
		return "", punyError(encoded)
	}
	if pos == len(encoded) {
		return encoded[:len(encoded)-1], nil
	}
	output := make([]rune, 0, len(encoded))
	if pos != 0 {
		for _, r := range encoded[:pos-1] {
			output = append(output, r)
		}
	}

	i, n, bias := int32(0), initialN, initialBias
	for pos < len(encoded) {
		oldI, w := i, int32(1)
		for k := base; ; k += base {
			if pos == len(encoded) {
				return "", punyError(encoded)
			}
			digit, ok := punyDecodeDigit(encoded[pos])
			if !ok {
				return "", punyError(encoded)
			}
			pos++
			i += digit * w
			if i < 0 {
				return "", punyError(encoded)
			}
			t := k - bias
			if t < tmin {
				t = tmin
			} else if t > tmax {
				t = tmax
			}
			if digit < t {
				break
			}
			w *= base - t
			if w >= math.MaxInt32/base {
				return "", punyError(encoded)
			}
		}
		x := int32(len(output) + 1)
		bias = punyAdapt(i-oldI, x, oldI == 0)
		n += i / x
		i %= x
		if n > utf8.MaxRune || len(output) >= 1024 {
			return "", punyError(encoded)
		}
		output = append(output, 0)
		copy(output[i+1:], output[i:])
		output[i] = n
		i++
	}
	return string(output), nil
}

func punyDecodeDigit(x byte) (digit int32, ok bool) {
	switch {
	case '0' <= x && x <= '9':
		return int32(x - ('0' - 26)), true
	case 'A' <= x && x <= 'Z':
		return int32(x - 'A'), true
	case 'a' <= x && x <= 'z':
		return int32(x - 'a'), true
	}
	return 0, false
}

// adapt is the bias adaptation function specified in section 6.1.
func punyAdapt(delta, numPoints int32, firstTime bool) int32 {
	if firstTime {
		delta /= damp
	} else {
		delta /= 2
	}
	delta += delta / numPoints
	k := int32(0)
	for delta > ((base-tmin)*tmax)/2 {
		delta /= base - tmin
		k += base
	}
	return k + (base-tmin+1)*delta/(delta+skew)
}
