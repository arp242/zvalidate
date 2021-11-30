Simple validation for Go. Some things that make it different from the (many)
other libraries:

- No struct tags – which I don't find a good tool for this – just functions.
- Validations return parsed values.
- Easy to display validation errors in UI.
- Doesn't use reflection (other than type assertions); mostly typed.
- No external dependencies.
- Easy to add nested validations.
- Not tied to HTTP (useful for validating CLI flags, for example).
- Supports translating error messages.

I originally wrote this at my previous employer
([github.com/teamwork/validate][tw]), this is an improved (and incompatible)
version.

API docs: https://godocs.io/zgo.at/zvalidate

[tw]: https://github.com/teamwork/validate

Basic example:

```go
email := "martin@arp42.net"

v := zvalidate.New()
v.Required("email", email)
m := v.Email("email", email)

if v.HasErrors() {
    fmt.Printf("Had the following validation errors:\n%s", v)
}

fmt.Printf("parsed email: %s\n", m.Address)
```

All validators are just method calls on the `Validator` struct, and follow the
same patterns:

- The input's zero type (empty string, 0, nil, etc.) is valid. Use the
  `Required()` validator if you want to make a parameter required.

- `key string, value [..]` are the first two arguments, where `key` is the
  parameter name (to display in the error or next to a form) and `value` is what
  we want validated (type of `value` depends on validation).

- Optionally accept a custom message as the last parameter.

The error text only includes a simple human description such as *"must be set"*
or *"must be a valid email"*. When adding new validations, make sure that they
can be displayed properly when joined with commas. A text such as *"Error: this
field must be higher than 42"* would look weird:

    must be set, Error: this field must be higher than 42

Validations
-----------

List of validations with abbreviated function signature (`key string, value
[..]` omitted):

| Function                         | Description                                |
| --------                         | -----------                                |
| Required()                       | Value must not be the type's zero value    |
| Exclude([]string) string         | Value is not in the exclude list           |
| Include([]string) string         | Value must be in the include list          |
| Range(min, max int)              | Minimum and maximum int value              |
| Len(min, max int) int            | Character length of string                 |
| Integer() int64                  | Integer value                              |
| Boolean() bool                   | Boolean value                              |
| Domain() []string                | Domain name; returns list of domain labels |
| Hostname() []string              | Any hostname                               |
| URL() \*url.URL                  | Valid URL                                  |
| Email() mail.Address             | Email address                              |
| IPv4() net.IP                    | IPv4 address                               |
| IP() net.IP                      | IPv4 or IPv6 address                       |
| HexColor() (uint8, uint8, uint8) | Colour as hex triplet (#123456 or #123)    |
| Date(layout string)              | Parse according to the given layout        |
| Phone() string                   | Looks like a phone number                  |
| UTF8()                           | String is valid UTF-8                      |
| Contains([]\*unicode.RangeTable) | Only allow the given character ranges      |

You can set your own errors with `v.Append()`:

```go
if !some_complex_condition {
    v.Append("foo", "must be a valid foo")
}
```

Nested validations
------------------

`Sub()` allows adding nested subvalidations; this is useful if a form creates
more than one object. For example:

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
the key name without subkey, as if you called `v.Append("key", "msg")`. This is
mostly to support cases like:

```go
func (Customer c) Validate() {
    v := validate.New()
    v.Email("email", c.Email)
    if v.HasErrors() {
        return v
    }

    // Check if this email already exists in the DB, may return various errors.
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

Displaying errors
-----------------

The `Validator` type satisfies the `error` interface, so you can return them as
errors; usually you want to return `ErrorOrNil()`.

The general idea is that validation errors should usually be displayed along the
input element, instead of a list in a flash message (but you can do either).

- To display a **flash message** or **CLI** just call `String()` or `HTML()`.

- For **Go templates** there is a `TemplateError()` helper which can be added to
  the `template.FuncMap`. See the godoc for that function for details and an
  example.

- For **JavaScript** `Errors` is represented as `map[string][]string`, and
  marshals well to JSON; in your frontend you just have to find the input
  belonging to the map key. A simple example might be:

  ```javascript
  var display_errors = function(errors) {
      var hidden = '';
      for (var k in errors) {
          if (!errors.hasOwnProperty(k))
              continue;

          var elem = document.querySelector('*[name=' + k + ']')
          if (!elem) {
              hidden += k + ': ' + errors[k].join(', ');
              continue;
          }

          var err = document.createElement('span');
          err.className = 'err';
          err.innerHTML = 'Error: ' + errors[k].join(', ') + '.';
          elem.insertAdjacentElement('afterend', err);
      }

      if (hidden !== '')
          alert(hidden);
  };

  display_errors({
      'xxx':    ['oh noes', 'asd'],
      'hidden': ['asd'],
  });
  ```


**caveat**: if there is an error without a corresponding form element then that
error won't be displayed. This is why the above examples `Pop()` all the errors
they want to display, and then display anything that's left at the end. This
prevents "hidden" errors.

i18n
----
You can change the messages with `Validator.Messages`:

```go
v := zvalidate.New()
v = v.Locale(zvalidate.Messages{
    Required: func() string { return myTranslateFunction("must be set") },
    Exclude:  func() string { return myTranslateFunction("cannot be ‘%s’") },
    // ...
})
```

You can also use this to change the default message if you don't like one of
them. `zvalidate.DefaultMessages` is used by default. If you don't specify one
of the struct fields then it will fall back to the values in that struct.
