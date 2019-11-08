[![Build Status](https://travis-ci.org/zgoat/zvalidate.svg?branch=master)](https://travis-ci.org/zgoat/zvalidate)
[![codecov](https://codecov.io/gh/zgoat/zvalidate/branch/master/graph/badge.svg?token=n0k8YjbQOL)](https://codecov.io/gh/zgoat/zvalidate)
[![GoDoc](https://godoc.org/github.com/zgoat/zvalidate?status.svg)](https://godoc.org/github.com/zgoat/zvalidate)

Validation for Go.

Basic usage example:

	v := zvalidate.New()
	v.Required("firstName", customer.FirstName)
	if v.HasErrors() {
		fmt.Println("Had the following validation errors:")
		for key, errors := range v.Errors {
			fmt.Printf("    %s: %s", key, strings.Join(errors))
		}
	}

When possible validators parse the value and return the result; e.g. `Email()`
returns `mail.Address`.

See godoc for more info.
