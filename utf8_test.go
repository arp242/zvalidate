// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zvalidate

import (
	"testing"
)

type ValidTest struct {
	in  string
	out bool
}

var validTests = []ValidTest{
	{"", true},
	{"a", true},
	{"abc", true},
	{"Ж", true},
	{"ЖЖ", true},
	{"брэд-ЛГТМ", true},
	{"☺☻☹", true},
	{"aa\xe2", false},
	{string([]byte{66, 250}), false},
	{string([]byte{66, 250, 67}), false},
	{"a\uFFFDb", true},
	{string("\xF4\x8F\xBF\xBF"), true},      // U+10FFFF
	{string("\xF4\x90\x80\x80"), false},     // U+10FFFF+1; out of range
	{string("\xF7\xBF\xBF\xBF"), false},     // 0x1FFFFF; out of range
	{string("\xFB\xBF\xBF\xBF\xBF"), false}, // 0x3FFFFFF; out of range
	{string("\xc0\x80"), false},             // U+0000 encoded in two bytes: incorrect
	{string("\xed\xa0\x80"), false},         // U+D800 high surrogate (sic)
	{string("\xed\xbf\xbf"), false},         // U+DFFF low surrogate (sic)
}

func TestValid(t *testing.T) {
	for _, tt := range validTests {
		if validString(tt.in) != tt.out {
			t.Errorf("ValidString(%q) = %v; want %v", tt.in, !tt.out, tt.out)
		}
	}
}
