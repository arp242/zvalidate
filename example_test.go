package zvalidate_test

import (
	"fmt"

	"zgo.at/zvalidate"
)

func ExampleTest() {
	name := "Martin"
	email := "martin@arp42.net"

	v := zvalidate.New()
	v.Required("email", name)
	m := v.Email("email", email)

	if v.HasErrors() {
		fmt.Printf("Had the following validation errors:\n%s", v)
	}

	fmt.Printf("parsed email: %s\n", m.Address)

	// Output: parsed email: martin@arp42.net
}
