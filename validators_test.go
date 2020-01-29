package zvalidate

import (
	"fmt"
	"net/mail"
	"reflect"
	"testing"
	"time"
)

func TestRequiredInt(t *testing.T) {
	tests := []struct {
		a    interface{}
		want bool
	}{
		{0, true},
		{int64(0), true},
		{uint(0), true},
		{uint64(0), true},
		{1, false},
		{int64(1), false},
		{uint(1), false},
		{uint64(1), false},
	}

	for i, tt := range tests {
		name := fmt.Sprintf("%v", i)
		t.Run(name, func(t *testing.T) {
			v := New()
			v.Required(name, tt.a)
			if got := v.HasErrors(); got != tt.want {
				t.Errorf("\ngot:  %#v\nwant: %#v\n", got, tt.want)
			}
		})
	}
}

func TestValidators(t *testing.T) {
	tests := []struct {
		val        func(Validator)
		wantErrors map[string][]string
	}{
		// Required
		{
			func(v Validator) {
				v.Required("firstName", "not empty")
				v.Required("age", 32)
			},
			make(map[string][]string),
		},
		{
			func(v Validator) {
				v.Required("lastName", "", "foo")
				v.Required("age", 0, "bar")
			},
			map[string][]string{"lastName": {"foo"}, "age": {"bar"}},
		},
		{
			func(v Validator) {
				v.Required("lastName", "")
				v.Required("age", 0)
			},
			map[string][]string{"lastName": {"must be set"}, "age": {"must be set"}},
		},
		{
			func(v Validator) {
				v.Required("email", "")
				v.Email("email", "")

				v.Required("email2", "asd")
				v.Email("email2", "asd")

				v.Required("email3", "asd@example.com")
				v.Email("email3", "asd@example.com")
			},
			map[string][]string{
				"email":  {"must be set"},
				"email2": {"must be a valid email address"},
			},
		},
		{
			func(v Validator) { v.Required("k", true) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Required("k", false) },
			map[string][]string{"k": {"must be set"}},
		},
		{
			func(v Validator) { v.Required("k", []string{}) },
			map[string][]string{"k": {"must be set"}},
		},
		{
			func(v Validator) { v.Required("k", *new([]string)) },
			map[string][]string{"k": {"must be set"}},
		},
		{
			func(v Validator) { v.Required("k", []string{""}) },
			map[string][]string{"k": {"must be set"}},
		},
		{
			func(v Validator) { v.Required("k", []string{"", "", ""}) },
			map[string][]string{"k": {"must be set"}},
		},
		{
			func(v Validator) { v.Required("k", []string{" "}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Required("k", []string{"", "", " "}) },
			make(map[string][]string),
		},

		// Required mailaddress
		{
			func(v Validator) {
				v.Required("k1", mail.Address{})
				//v.Required("k2", mailaddress.List{})
			},
			//map[string][]string{"k1": {"must be set"}, "k2": {"must be set"}},
			map[string][]string{"k1": {"must be set"}},
		},
		{
			func(v Validator) {
				v.Required("k1", mail.Address{Address: "foo@example.com"})
				//v.Required("k2", mailaddress.List{mailaddress.New("", "asd")})
			},
			make(map[string][]string),
		},

		// Len
		{
			func(v Validator) { v.Len("v", "w00t", 2, 5) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Len("v", "w00t", 4, 0) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Len("v", "w00t", 0, 4) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Len("v", "w00t", 1, 2) },
			map[string][]string{"v": {"must be shorter than 2 characters"}},
		},
		{
			func(v Validator) { v.Len("v", "w00t", 1, 2, "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.Len("v", "w00t", 16, 32) },
			map[string][]string{"v": {"must be longer than 16 characters"}},
		},
		{
			func(v Validator) { v.Len("v", "ราคาเหนือจอง", 12, 12) },
			make(map[string][]string),
		},
		// Exclude
		{
			func(v Validator) { v.Exclude("key", "val", []string{}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Exclude("key", "val", nil) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Exclude("key", "val", []string{"valx"}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Exclude("key", "val", []string{"VAL"}) },
			map[string][]string{"key": {`cannot be ‘VAL’`}},
		},
		{
			func(v Validator) { v.Exclude("key", "val", []string{"VAL"}, "foo") },
			map[string][]string{"key": {`foo`}},
		},
		{
			func(v Validator) { v.Exclude("key", "val", []string{"hello", "val"}) },
			map[string][]string{"key": {`cannot be ‘val’`}},
		},

		// Include
		{
			func(v Validator) { v.Include("key", "val", []string{}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Include("key", "val", nil) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Include("key", "val", []string{"valx"}) },
			map[string][]string{"key": {`must be one of ‘valx’`}},
		},
		{
			func(v Validator) { v.Include("key", "val", []string{"valx"}, "foo") },
			map[string][]string{"key": {`foo`}},
		},
		{
			func(v Validator) { v.Include("key", "val", []string{"VAL"}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Include("key", "val", []string{"hello", "val"}) },
			make(map[string][]string),
		},

		// Domain
		{
			func(v Validator) { v.Domain("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "example.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "example.com.test.asd") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "example-test.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "ﻢﻔﺗﻮﺣ.ﺬﺑﺎﺑﺓ") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "xn--pgbg2dpr.xn--mgbbbe5a") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.Domain("v", "one-label") },
			map[string][]string{"v": {"must be a valid domain"}},
		},
		{
			func(v Validator) { v.Domain("v", "one-label", "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.Domain("v", "example.com:-)") },
			map[string][]string{"v": {"must be a valid domain"}},
		},
		{
			func(v Validator) { v.Domain("v", "ex ample.com") },
			map[string][]string{"v": {"must be a valid domain"}},
		},

		// URL
		{
			func(v Validator) { v.URL("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "example.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "example.com.test.asd/testing.html") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "example-test.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "ﻢﻔﺗﻮﺣ.ﺬﺑﺎﺑﺓ") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "xn--pgbg2dpr.xn--mgbbbe5a") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "http://x.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "unknownschema://x.com?q=v&x=2%3Aa#frag") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "complex://x.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "http://sunbeam.teamwork.localhost:9000/bucket/1/avatar-1.jpeg") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.URL("v", "one-label") },
			map[string][]string{"v": {"must be a valid url"}},
		},
		{
			func(v Validator) { v.URL("v", "http://x") },
			map[string][]string{"v": {"must be a valid url"}},
		},
		{
			func(v Validator) { v.URL("v", "one-label", "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.URL("v", "example.com:-)") },
			map[string][]string{"v": {"must be a valid url"}},
		},
		{
			func(v Validator) { v.URL("v", "ex ample.com") },
			map[string][]string{"v": {"must be a valid url: parse http://ex%20ample.com: invalid URL escape \"%20\""}},
		},
		{
			func(v Validator) { v.URL("v", "unknown_schema://x.com") },
			map[string][]string{"v": {"must be a valid url: parse unknown_schema://x.com: " +
				"first path segment in URL cannot contain colon"}},
		},

		// HexColor
		{
			func(v Validator) { v.HexColor("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.HexColor("v", "#36a") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.HexColor("v", "#3a6ea5") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.HexColor("v", "fff") },
			map[string][]string{"v": {"must be a valid color code"}},
		},
		{
			func(v Validator) { v.HexColor("v", "#ff") },
			map[string][]string{"v": {"must be a valid color code"}},
		},
		{
			func(v Validator) { v.HexColor("v", "#fffffff") },
			map[string][]string{"v": {"must be a valid color code"}},
		},

		// Date
		{
			func(v Validator) { v.Date("k", "2017-11-14T13:37:00Z", time.RFC3339) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Date("k", "2017-11-14", time.RFC3339) },
			map[string][]string{"k": {"must be a date as ‘2006-01-02T15:04:05Z07:00’"}},
		},
		{
			func(v Validator) { v.Date("k", "2017-11-14", time.RFC3339, "not valid") },
			map[string][]string{"k": {"not valid"}},
		},

		// Email
		// Don't need to extensively validate emails, we have tests for that in
		// the mailaddress package already.
		{
			func(v Validator) { v.Email("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Email("v", "martin@example.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Email("v", "martin") },
			map[string][]string{"v": {"must be a valid email address"}},
		},
		{
			func(v Validator) { v.Email("v", "martin", "foo") },
			map[string][]string{"v": {"foo"}},
		},

		// IPv4
		{
			func(v Validator) { v.IPv4("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.IPv4("v", "127.0.0.1") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.IPv4("v", "127.0.0.4/8") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "127.0.0.4/8", "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.IPv4("v", "127.1") }, // Technically correct but Go doesn't seem to like it.
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "127.0.0.506") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "127.") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "asdf") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "::1") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},

		// IP
		{
			func(v Validator) { v.IP("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.IP("v", "127.0.0.1") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.IP("v", "::1") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.IP("v", "127.0.0.4/8") },
			map[string][]string{"v": {"must be a valid IPv4 or IPv6 address"}},
		},
		{
			func(v Validator) { v.IP("v", "127.0.0.4/8", "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.IP("v", "127.1") }, // Technically correct but Go doesn't seem to like it.
			map[string][]string{"v": {"must be a valid IPv4 or IPv6 address"}},
		},
		{
			func(v Validator) { v.IP("v", "127.0.0.506") },
			map[string][]string{"v": {"must be a valid IPv4 or IPv6 address"}},
		},
		{
			func(v Validator) { v.IP("v", "127.") },
			map[string][]string{"v": {"must be a valid IPv4 or IPv6 address"}},
		},
		{
			func(v Validator) { v.IP("v", "asdf") },
			map[string][]string{"v": {"must be a valid IPv4 or IPv6 address"}},
		},

		// Phone
		{
			func(v Validator) { v.Phone("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Phone("v", "12345123") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Phone("v", "(+31)-12345123") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Phone("v", "[+31]-12345123") },
			map[string][]string{"v": {"must be a valid phone number"}},
		},

		// Range
		{
			func(v Validator) { v.Range("v", 4, 2, 5) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Range("v", 4, 4, 0) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Range("v", 4, 0, 4) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Range("v", 4, 1, 2) },
			map[string][]string{"v": {"must be lower than 2"}},
		},
		{
			func(v Validator) { v.Range("v", 4, 1, 2, "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.Range("v", 4, 16, 32) },
			map[string][]string{"v": {"must be higher than 16"}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			tt.val(v)

			if !reflect.DeepEqual(v.Errors, tt.wantErrors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", v.Errors, tt.wantErrors)
			}
		})
	}
}

func TestInteger(t *testing.T) {
	tests := []struct {
		val        func(Validator) int64
		want       int64
		wantErrors map[string][]string
	}{
		{
			func(v Validator) int64 { return v.Integer("k", "") },
			0,
			make(map[string][]string),
		},
		{
			func(v Validator) int64 { return v.Integer("k", "6") },
			6,
			make(map[string][]string),
		},
		{
			func(v Validator) int64 { return v.Integer("k", " 6 ") },
			6,
			make(map[string][]string),
		},
		{
			func(v Validator) int64 { return v.Integer("k", "0") },
			0,
			make(map[string][]string),
		},
		{
			func(v Validator) int64 { return v.Integer("k", "-1") },
			-1,
			make(map[string][]string),
		},
		{
			func(v Validator) int64 { return v.Integer("k", "1.2") },
			0,
			map[string][]string{"k": {"must be a whole number"}},
		},
		{
			func(v Validator) int64 { return v.Integer("k", "asd") },
			0,
			map[string][]string{"k": {"must be a whole number"}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			i := tt.val(v)

			if !reflect.DeepEqual(v.Errors, tt.wantErrors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", v.Errors, tt.wantErrors)
			}

			if i != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", i, tt.want)
			}
		})
	}
}

func TestBoolean(t *testing.T) {
	tests := []struct {
		val        func(Validator) bool
		want       bool
		wantErrors map[string][]string
	}{
		{
			func(v Validator) bool { return v.Boolean("k", "on") },
			false,
			map[string][]string{"k": {"must be a boolean"}},
		},
		{
			func(v Validator) bool { return v.Boolean("k", "") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "0") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "n") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "no") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "f") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "false") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "FALSE") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "1") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "y") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "yes") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "t") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "true") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "TRUE") },
			true,
			make(map[string][]string),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			i := tt.val(v)

			if !reflect.DeepEqual(v.Errors, tt.wantErrors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", v.Errors, tt.wantErrors)
			}

			if i != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", i, tt.want)
			}
		})
	}
}

func TestDomain(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"xn--bcher-kva.example", []string{"bücher", "example"}},
		{"www.example.com", []string{"www", "example", "com"}},
		{"www.example.com.", []string{"www", "example", "com"}},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			out := v.Domain("", tt.in)

			if v.HasErrors() {
				t.Fatal(v.Error())
			}

			if !reflect.DeepEqual(out, tt.want) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}
