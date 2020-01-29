package zvalidate_test

import (
	"fmt"
	"html/template"
	"os"

	"zgo.at/zvalidate"
)

func Example() {
	email := "martin@arp42.net"

	v := zvalidate.New()
	v.Required("email", email)
	m := v.Email("email", email)

	if v.HasErrors() {
		fmt.Printf("Had the following validation errors:\n%s", v)
	}

	fmt.Printf("parsed email: %s\n", m.Address)

	// Output: parsed email: martin@arp42.net
}

func ExampleTemplateError() {
	funcs := template.FuncMap{
		"validate":   zvalidate.TemplateError,
		"has_errors": zvalidate.TemplateHasErrors,
	}

	t := template.Must(template.New("").Funcs(funcs).Parse(`
<input name="xxx">
{{validate "xxx" .Validate}}

{{if has_errors .Validate}}Hidden: {{.Validate.HTML}}{{end}}
	`))

	v := zvalidate.New()
	v.Append("xxx", "oh noes")
	v.Append("hidden", "sneaky")

	t.Execute(os.Stdout, map[string]interface{}{
		"Validate": &v,
	})

	// Output:
	// <input name="xxx">
	// <span class="err">Error: oh noes</span>
	//
	// Hidden: <ul class='zvalidate'>
	// <li><strong>hidden</strong>: sneaky.</li>
	// </ul>
}
