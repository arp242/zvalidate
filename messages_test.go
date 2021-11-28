package zvalidate

import (
	"fmt"
	"testing"
)

func TestTranslate(t *testing.T) {
	v := New()
	v.Messages(Messages{Required: func() string { return "X" }})
	v.Required("empty", "")
	v.Required("empty-c", "", "msg")

	have := fmt.Sprintf("%v", v.Errors)
	want := `map[empty:[X] empty-c:[msg]]`
	if have != want {
		t.Errorf("\nhave: %s\nwant: %s", have, want)
	}
}
