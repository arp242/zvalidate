// Package zvalidate provides simple validation for Go.
//
// See the README.markdown for an introduction.
package zvalidate

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

// New initializes a new Validator.
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

// Append a new error.
func (v *Validator) Append(key, value string, format ...interface{}) {
	v.Errors[key] = append(v.Errors[key], fmt.Sprintf(value, format...))
}

// Pop an error, removing all errors for this key.
//
// This is mostly useful when displaying errors next to forms: Pop() all the
// errors you want to display, and then display anything that's left with a
// flash message or the like. This prevents "hidden" errors.
//
// Returns nil if there are no errors for this key.
func (v *Validator) Pop(key string) []string {
	if len(v.Errors[key]) == 0 {
		return nil
	}

	errs := v.Errors[key]
	delete(v.Errors, key)
	return errs
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

// Sub adds sub-validations.
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
//   // key: "settings.domain"
//   v.Sub("settings", -1, customer.Settings.Validate())
//
//   // key: "addresses[1].city"
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

// Strings representation of all errors, or a blank string if there are none.
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

// HTML representation of all errors, or a blank string if there are none.
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
