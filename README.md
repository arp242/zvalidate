[![Build Status](https://travis-ci.org/zgoat/zvalidate.svg?branch=master)](https://travis-ci.org/zgoat/zvalidate)
[![codecov](https://codecov.io/gh/zgoat/zvalidate/branch/master/graph/badge.svg?token=n0k8YjbQOL)](https://codecov.io/gh/zgoat/zvalidate)
[![GoDoc](https://godoc.org/zgo.at/zvalidate?status.svg)](https://pkg.go.dev/zgo.at/zvalidate)

Validation for Go.

Basic usage example:

    v := zvalidate.New()
    v.Required("email", customer.Email)
    m := v.Email("email", customer.Email)

    if v.HasErrors() {
        fmt.Println("Had the following validation errors:")
        for key, errors := range v.Errors {
            fmt.Printf("    %s: %s", key, strings.Join(errors))
        }
    }

    fmt.Printf("parsed email: %q <%s>\n", m.Name, m.Address)

When possible validators parse the value and return the result; e.g. `Email()`
returns `mail.Address`.

All validators treat the input's zero type (empty string, 0, nil, etc.) as
valid. Use the `Required()` validator if you want to make a parameter required. 

All validators optionally accept a custom message as the last parameter:

    v.Required("key", value, "you really need to set this")

The error text only includes a simple human description such as "must be set"
or "must be a valid email". When adding new validations, make sure that they
can be displayed properly when joined with commas. A text such as "Error:
this field must be higher than 42" would look weird:

    must be set, Error: this field must be higher than 42

You can set your own errors with v.Append():

    if !condition {
        v.Append("key", "must be a valid foo")
    }

---

List of validations with abbreviated function signature:

| Function                         | Description                                 |
| --------                         | -----------                                 |
| Required()                       | Value must not be the type's zero value     |
| Exclude([]string)                | Value is not in the exclude list            |
| Include([]string)                | Value is in the include list                |
| Range(min, max int)              | Minimum and maximum int value.              |
| Len(min, max int) int            | Character length of string                  |
| Integer() int64                  | Integer value                               |
| Boolean() bool                   | Boolean value                               |
| Domain() []string                | Domain name; returns list of domain labels. |
| URL() \*url.URL                  | Valid URL                                   |
| Email() mail.Address             | Email address                               |
| IPv4() net.IP                    | IPv4 address                                |
| IP() net.IP                      | IPv4 or IPv6 address                        |
| HexColor() (uint8, uint8, uint8) | Colour as hex triplet (#123456 or #123)     |
| Date(layout string)              | Parse according to the given layout         |
| Phone() string                   | Looks like a phone number                   |
