[![Build Status](https://travis-ci.org/zgoat/validate.svg?branch=master)](https://travis-ci.org/zgoat/validate)
[![codecov](https://codecov.io/gh/zgoat/validate/branch/master/graph/badge.svg?token=n0k8YjbQOL)](https://codecov.io/gh/zgoat/validate)
[![GoDoc](https://godoc.org/github.com/zgoat/validate?status.svg)](https://godoc.org/github.com/zgoat/validate)

HTTP request parameter validation for Go.

Basic usage example:

	v := validate.New()
	v.Required("firstName", customer.FirstName)
	if v.HasErrors() {
		fmt.Println("Had the following validation errors:")
		for key, errors := range v.Errors {
			fmt.Printf("    %s: %s", key, strings.Join(errors))
		}
	}

See godoc for more info.
