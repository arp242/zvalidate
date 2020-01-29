package zvalidate

import (
	"fmt"
	"html/template"
	"testing"
)

func TestTemplateError(t *testing.T) {
	tests := []struct {
		key  string
		in   *Validator
		want template.HTML
	}{
		{"", nil, ""},
		{"xxx", nil, ""},
		{"", &Validator{Errors: map[string][]string{"k": {"xx"}}}, ""},
		{"k", &Validator{Errors: map[string][]string{"k": {"xx"}}}, `<span class="err">Error: xx</span>`},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			out := TemplateError(tt.key, tt.in)
			if out != tt.want {
				t.Errorf("\nout:  %q\nwant: %q", out, tt.want)
			}
		})
	}
}

func TestTemplateHasErrors(t *testing.T) {
	tests := []struct {
		in   *Validator
		want bool
	}{
		{nil, false},
		{&Validator{}, false},
		{&Validator{Errors: map[string][]string{}}, false},
		{&Validator{Errors: map[string][]string{"k": {"xx"}}}, true},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			out := TemplateHasErrors(tt.in)
			if out != tt.want {
				t.Errorf("\nout:  %t\nwant: %t", out, tt.want)
			}
		})
	}
}
