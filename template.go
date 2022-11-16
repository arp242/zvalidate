package zvalidate

import (
	"fmt"
	"html/template"
	"strings"
)

// TemplateError displays validation errors for the given key.
//
// This will Pop() errors and modify the Validator in-place, so we can see if
// there are any "hidden" errors later on.
func TemplateError(k string, v *Validator) template.HTML {
	if v == nil || !v.HasErrors() {
		return template.HTML("")
	}

	errs := v.Pop(k)
	if errs == nil {
		return template.HTML("")
	}

	return template.HTML(fmt.Sprintf(`<span class="err">Error: %s</span>`,
		template.HTMLEscapeString(strings.Join(errs, ", "))))
}

// TemplateHasErrors reports if there are any validation errors.
//
// This is useful because "and" evaluates all arguments, and this will error
// out:
//
//	{{if and .Validate .Validate.HasErrors}}
func TemplateHasErrors(v *Validator) bool {
	if v == nil {
		return false
	}
	return v.HasErrors()
}
