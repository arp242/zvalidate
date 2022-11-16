// Package zvalidate provides simple validation for Go.
//
// See the README.markdown for an introduction.
package zvalidate

import (
	"encoding/json"
	"errors"
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
	msg    Messages
}

// New initializes a new Validator.
func New() Validator {
	return Validator{Errors: make(map[string][]string), msg: DefaultMessages}
}

// Messages sets the messages to use for validation errors.
func (v *Validator) Messages(m Messages) {
	if m.Required == nil {
		m.Required = DefaultMessages.Required
	}
	if m.Domain == nil {
		m.Domain = DefaultMessages.Domain
	}
	if m.Hostname == nil {
		m.Hostname = DefaultMessages.Hostname
	}
	if m.URL == nil {
		m.URL = DefaultMessages.URL
	}
	if m.Email == nil {
		m.Email = DefaultMessages.Email
	}
	if m.IPv4 == nil {
		m.IPv4 = DefaultMessages.IPv4
	}
	if m.IP == nil {
		m.IP = DefaultMessages.IP
	}
	if m.HexColor == nil {
		m.HexColor = DefaultMessages.HexColor
	}
	if m.LenLonger == nil {
		m.LenLonger = DefaultMessages.LenLonger
	}
	if m.LenShorter == nil {
		m.LenShorter = DefaultMessages.LenShorter
	}
	if m.Exclude == nil {
		m.Exclude = DefaultMessages.Exclude
	}
	if m.Include == nil {
		m.Include = DefaultMessages.Include
	}
	if m.Integer == nil {
		m.Integer = DefaultMessages.Integer
	}
	if m.Bool == nil {
		m.Bool = DefaultMessages.Bool
	}
	if m.Date == nil {
		m.Date = DefaultMessages.Date
	}
	if m.Phone == nil {
		m.Phone = DefaultMessages.Phone
	}
	if m.RangeHigher == nil {
		m.RangeHigher = DefaultMessages.RangeHigher
	}
	if m.RangeLower == nil {
		m.RangeLower = DefaultMessages.RangeLower
	}
	if m.UTF8 == nil {
		m.UTF8 = DefaultMessages.UTF8
	}
	if m.Contains == nil {
		m.Contains = DefaultMessages.Contains
	}
	v.msg = m
}

// As tries to convert this error to a Validator, returning nil if it's not.
func As(err error) *Validator {
	v := new(Validator)
	if errors.As(err, v) || errors.As(err, &v) {
		return v
	}
	return nil
}

// Error interface.
func (v Validator) Error() string { return v.String() }

// Code returns the HTTP status code for the error. Satisfies the guru.coder
// interface in zgo.at/guru.
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
//	if v.HasErrors() {
//	    return v
//	}
//	return nil
//
// Can now be:
//
//	return v.ErrorOrNil()
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
//	func (c Customer) validateSettings() error {
//	    v := zvalidate.New()
//	    v.Required("domain", c.Domain)
//	    v.Required("email", c.Email)
//	    return v.ErrorOrNil()
//	}
//
//	v := zvalidate.New()
//	v.Required("name", customer.Name)
//
//	// Keys will be added as "settings.domain" and "settings.email".
//	v.Sub("settings", "", customer.validateSettings())
//
//	// List as array; keys will be added as "addresses[0].city" etc.
//	for i, addr := range customer.Addresses {
//	    v.Sub("addresses", i, addr.Validate())
//	}
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
		if k != "" {
			b.WriteString(k)
			b.WriteString(": ")
		}
		b.WriteString(strings.Join(v.Errors[k], ", "))
		b.WriteString(".\n")
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
		b.WriteString("<li>")
		if k != "" {
			b.WriteString(fmt.Sprintf("<strong>%s</strong>: ", template.HTMLEscapeString(k)))
		}
		b.WriteString(fmt.Sprintf("%s.</li>\n", template.HTMLEscapeString(strings.Join(v.Errors[k], ", "))))
	}

	b.WriteString("</ul>\n")
	return template.HTML(b.String())
}
