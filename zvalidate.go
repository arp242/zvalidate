// Package zvalidate provides simple validation for Go.
//
// Basic usage example:
//
//    v := zvalidate.New()
//    v.Required("email", customer.Email)
//    m := v.Email("email", customer.Email)
//
//    if v.HasErrors() {
//        fmt.Println("Had the following validation errors:")
//        for key, errors := range v.Errors {
//            fmt.Printf("    %s: %s", key, strings.Join(errors))
//        }
//    }
//
//    fmt.Printf("parsed email: %q <%s>\n", m.Name, m.Address)
//
// All validators treat the input's zero type (empty string, 0, nil, etc.) as
// valid. Use the Required() validator if you want to make a parameter required.
//
// All validators optionally accept a custom message as the last parameter:
//
//   v.Required("key", value, "you really need to set this")
//
// The error text only includes a simple human description such as "must be set"
// or "must be a valid email". When adding new validations, make sure that they
// can be displayed properly when joined with commas. A text such as "Error:
// this field must be higher than 42" would look weird:
//
//   must be set, Error: this field must be higher than 42
//
// You can set your own errors with v.Append():
//
//   if !condition {
//       v.Append("key", "must be a valid foo")
//   }
package zvalidate // import "zgo.at/zvalidate"

import (
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strings"
)

// Validator hold the validation errors.
//
// Typically you shouldn't create this directly but use the New() function.
type Validator struct {
	Errors map[string][]string `json:"errors"`
}

// New makes a new Validator and ensures that it is properly initialized.
func New() Validator {
	return Validator{Errors: make(map[string][]string)}
}

// Error interface.
func (v Validator) Error() string { return v.String() }

// Code returns the HTTP status code for the error. Satisfies the guru.coder
// interface in github.com/teamwork/guru.
func (v Validator) Code() int { return 400 }

// ErrorJSON for reporting errors as JSON.
func (v Validator) ErrorJSON() ([]byte, error) { return json.Marshal(v) }

// Append a new error to the error list for this key.
func (v *Validator) Append(key, value string, format ...interface{}) {
	v.Errors[key] = append(v.Errors[key], fmt.Sprintf(value, format...))
}

// HasErrors reports if this validation has any errors.
func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}

// ErrorOrNil returns nil if there are no errors, or the Validator object if
// there are.
//
// This makes it a bit more elegant to return from a function:
//
//   if v.HasErrors() {
//       return v
//   }
//   return nil
//
// Can now be:
//
//   return v.ErrorOrNil()
func (v *Validator) ErrorOrNil() error {
	if v.HasErrors() {
		return v
	}
	return nil
}

// Sub allows adding sub-validations.
//
// Errors from the subvalidation are merged with the top-level one, the keys are
// added as "top.sub" or "top[n].sub".
//
// If the error is not a Validator the text will be added as just the key name
// without subkey (i.e. the same as v.Append("key", "msg")).
//
// For example:
//
//   v := zvalidate.New()
//   v.Required("name", customer.Name)
//
//   // e.g. "settings.domain"
//   v.Sub("settings", -1, customer.Settings.Validate())
//
//   // e.g. "addresses[1].city"
//   for i, a := range customer.Addresses {
//       a.Sub("addresses", i, c.Validate())
//   }
func (v *Validator) Sub(key, subKey string, err error) {
	if err == nil {
		return
	}

	if subKey != "" {
		key = fmt.Sprintf("%s[%s]", key, subKey)
	}

	sub, ok := err.(*Validator)
	if !ok {
		ss, ok := err.(Validator)
		if !ok {
			v.Append(key, err.Error())
			return
		}
		sub = &ss
	}
	if !sub.HasErrors() {
		return
	}

	for k, val := range sub.Errors {
		mk := fmt.Sprintf("%s.%s", key, k)
		v.Errors[mk] = append(v.Errors[mk], val...)
	}
}

// Merge errors from another validator in to this one.
func (v *Validator) Merge(other Validator) {
	for k, val := range other.Errors {
		v.Errors[k] = append(v.Errors[k], val...)
	}
}

// Strings representation of all errors, or a blank string if there are no
// errors.
func (v *Validator) String() string {
	if !v.HasErrors() {
		return ""
	}

	// Make sure the order is always the same.
	keys := make([]string, len(v.Errors))
	i := 0
	for k := range v.Errors {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("%s: %s.\n", k, strings.Join(v.Errors[k], ", ")))
	}
	return b.String()
}

// HTML representation of all errors, or a blank string if there are no errors.
func (v *Validator) HTML() template.HTML {
	if !v.HasErrors() {
		return ""
	}

	// Make sure the order is always the same.
	keys := make([]string, len(v.Errors))
	i := 0
	for k := range v.Errors {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("<ul class='zvalidate'>\n")
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("<li><strong>%s</strong>: %s.</li>\n",
			template.HTMLEscapeString(k),
			template.HTMLEscapeString(strings.Join(v.Errors[k], ", "))))
	}

	b.WriteString("</ul>\n")
	return template.HTML(b.String())
}
