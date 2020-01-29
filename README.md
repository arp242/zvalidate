[![Build Status](https://travis-ci.org/zgoat/zvalidate.svg?branch=master)](https://travis-ci.org/zgoat/zvalidate)
[![codecov](https://codecov.io/gh/zgoat/zvalidate/branch/master/graph/badge.svg?token=n0k8YjbQOL)](https://codecov.io/gh/zgoat/zvalidate)
[![GoDoc](https://godoc.org/zgo.at/zvalidate?status.svg)](https://pkg.go.dev/zgo.at/zvalidate)

Simple validation for Go. Some things that make it different from the (many)
other libraries:

- Validations return parsed values.
- No struct tags, which I don't find a good tool for this kind of thing.
- Easy to display validation errors in UI.
- Doesn't use reflection (other than type assertions).
- Has no external dependencies.
- Easy to add nested validations.

I originally wrote this at my previous employer
([github.com/teamwork/validate][tw]), this is an improved (and incompatible)
version.

[tw]: https://github.com/teamwork/validate

Basic usage example:

```go
name := "Martin"
email := "martin@arp42.net"

v := zvalidate.New()
v.Required("email", name)
m := v.Email("email", email)

if v.HasErrors() {
    fmt.Printf("Had the following validation errors:\n%s", v)
}

fmt.Printf("parsed email: %s\n", m.Address)
```

All validators are just method calls on the `Validator` struct, and follow the
same patterns:

- treat the input's zero type (empty string, 0, nil, etc.) as valid (use the
  `Required()` validator if you want to make a parameter required);

- have `key string, value [..]` as the first two arguments, where `key` is the
  parameter name (to display in the error or next to a form) and `value` is what
  we want validated (type of `value` depends on validation);

- optionally accept a custom message as the last parameter.

The error text only includes a simple human description such as *"must be set"*
or *"must be a valid email"*. When adding new validations, make sure that they
can be displayed properly when joined with commas. A text such as *"Error: this
field must be higher than 42"* would look weird:

    must be set, Error: this field must be higher than 42

Validations
-----------

List of validations with abbreviated function signature (`key string, value
[..]` omitted):

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

You can set your own errors with v.Append():

```go
if !some_complex_condition {
    v.Append("foo", "must be a valid foo")
}
```

Nested validations
------------------

`Sub()` allows adding nested subvalidations; this is useful if a form creates
more than one object.

Nothing is nested in the `Errors` data structure (which is just a
`map[string][]string`, and is mostly just a bit of code to create consistent
keys.

For example:

```go
v := zvalidate.New()
v.Sub("settings", -1, customer.Settings.Validate())
```

This will merge the `Validator` object in to `v` and prefix all the keys with
`settings.`, so you'll have `settings.timezone` (instead of just `timezone`).

You can also add arrays:

```go
for i, a := range customer.Addresses {
    a.Sub("addresses", i, c.Validate())
}
```

This will be added as `addresses[0].city`, `addresses[1].city`, etc.

If the error is not a `Validator` then the `Error()` text will be added as just
the key name without subkey (same as v.Append("key", "msg"); this is mostly to
support cases like:

```go
func (Customer c) Validate() {
    v := validate.New()

    ok, err := c.isUniqueEmail(c.Email)
    if err != nil {
        return err
    }
    if !ok {
        v.Append("email", "must be unique")
    }

    return v.ErrorOrNil()
}
```

### Displaying errors

The `Validator` type satisfies the `error` interface, so you can return them as
errors; usually you want to return `ErrorOrNil()`.

The general idea is that validation errors should usually be displayed along the
input element, instead of a list in a flash message (but you can do either).

To display a flash message just call `String()` or `HTML()`.

`Errors` is represented as `map[string][]string`, and marshals well to JSON; in
your frontend you just have to find the input belonging to the map key:

```
// TODO
```

In Go templates:

```go
template.FuncMap["validate"] = func(k string, v map[string][]string) template.HTML {
    if v == nil {
        return template.HTML("")
    }
    e, ok := v[k]
    if !ok {
        return template.HTML("")
    }
    return template.HTML(fmt.Sprintf(`<span class="err">Error: %s</span>`,
        template.HTMLEscapeString(strings.Join(e, ", "))))
}
```

And then use it:

    <label for="data_retention">Data retention in days</label>
    <input type="number" name="settings.data_retention" id="limits_page" value="{{.Site.Settings.DataRetention}}">
    {{validate "site.settings.data_retention" .Validate}}

**caveat**: if there is an error with a corresponding form element then that
won't be displayed. This is why the above examples `Pop()` all the errors they
want to display, and then display anything that's left. This prevents "hidden"
errors.

### i18n

There is no direct support for i18n, but the messages are all exported as
`Message*` and can be replaced by your i18n system of choice.
