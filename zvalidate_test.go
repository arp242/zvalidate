package zvalidate

import (
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		a, b, want map[string][]string
	}{
		{
			map[string][]string{},
			map[string][]string{},
			map[string][]string{},
		},
		{
			map[string][]string{"a": {"b"}},
			map[string][]string{},
			map[string][]string{"a": {"b"}},
		},
		{
			map[string][]string{},
			map[string][]string{"a": {"b"}},
			map[string][]string{"a": {"b"}},
		},
		{
			map[string][]string{"a": {"b"}},
			map[string][]string{"a": {"c"}},
			map[string][]string{"a": {"b", "c"}},
		},
		{
			map[string][]string{"a": {"b"}},
			map[string][]string{"q": {"c"}},
			map[string][]string{"a": {"b"}, "q": {"c"}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			in := New()
			in.Errors = tt.a
			other := New()
			other.Errors = tt.b

			in.Merge(other)

			if !reflect.DeepEqual(tt.want, in.Errors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", in.Errors, tt.want)
			}
		})
	}
}

func TestSub(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		v := New()
		v.Required("name", "")
		v.HexColor("color", "not a color")

		// Easy case
		s := New()
		s.Required("domain", "")
		s.Email("contactEmail", "not an email")
		v.Sub("setting", "", s.ErrorOrNil())

		// List
		addr1 := New()
		addr1.Required("city", "Bristol")
		v.Sub("addresses", "home", addr1)
		addr2 := New()
		addr2.Required("city", "")
		v.Sub("addresses", "office", addr2)

		// Non-Validator.
		v.Sub("other", "", errors.New("oh noes"))
		v.Sub("emails", "home", nil)
		v.Sub("emails", "office", errors.New("not an email"))

		// Sub with Sub.
		s1 := New()
		s2 := New()
		s2.Append("err", "very sub")
		s1.Sub("sub2", "", s2)
		v.Sub("sub1", "", s1)

		ls1 := New()
		ls2 := New()
		ls2.Append("err", "very sub")
		ls1.Sub("lsub2", "holiday", ls2)
		v.Sub("lsub1", "", ls1)

		want := map[string][]string{
			"lsub1.lsub2[holiday].err": []string{"very sub"},
			"sub1.sub2.err":            []string{"very sub"},
			"name":                     []string{"must be set"},
			"color":                    []string{"must be a valid color code"},
			"setting.domain":           []string{"must be set"},
			"setting.contactEmail":     []string{"must be a valid email address"},
			"addresses[office].city":   []string{"must be set"},
			"other":                    []string{"oh noes"},
			"emails[office]":           []string{"not an email"},
		}

		if d := cmp.Diff(v.Errors, want); d != "" {
			t.Errorf("(-got +want)\n:%s", d)
		}
	})
}

func TestString(t *testing.T) {
	tests := []struct {
		in   Validator
		want string
	}{
		{Validator{}, ""},
		{Validator{map[string][]string{}}, ""},

		{Validator{map[string][]string{
			"k": {"oh no"},
		}}, "k: oh no.\n"},
		{Validator{map[string][]string{
			"k": {"oh no", "more"},
		}}, "k: oh no, more.\n"},
		{Validator{map[string][]string{
			"k": {"oh no", "more", "even more"},
		}}, "k: oh no, more, even more.\n"},
		{Validator{map[string][]string{
			"k":  {"oh no", "more", "even more"},
			"k2": {"asd"},
		}}, "k: oh no, more, even more.\nk2: asd.\n"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := tt.in.String()
			if out != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}

func BenchmarkString(b *testing.B) {
	v := New()
	noOfErrors := 256
	const err = "Oh no!"
	for i := 0; i < noOfErrors; i++ {
		key := fmt.Sprintf("err%d", i)
		v.Append(key, err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.String()
	}
}

func TestHTML(t *testing.T) {
	tests := []struct {
		in   Validator
		want template.HTML
	}{
		{Validator{}, ""},
		{Validator{map[string][]string{}}, ""},

		{Validator{map[string][]string{
			"k": {"oh no"},
		}}, "<ul class='zvalidate'>\n<li><strong>k</strong>: oh no.</li>\n</ul>\n"},
		{Validator{map[string][]string{
			"k": {"oh no", "more"},
		}}, "<ul class='zvalidate'>\n<li><strong>k</strong>: oh no, more.</li>\n</ul>\n"},
		{Validator{map[string][]string{
			"k": {"oh no", "more", "even more"},
		}}, "<ul class='zvalidate'>\n<li><strong>k</strong>: oh no, more, even more.</li>\n</ul>\n"},
		{Validator{map[string][]string{
			"k":  {"oh no", "more", "even more"},
			"k2": {"asd"},
		}}, "<ul class='zvalidate'>\n<li><strong>k</strong>: oh no, more, even more.</li>\n<li><strong>k2</strong>: asd.</li>\n</ul>\n"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := tt.in.HTML()
			if out != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}

func TestErrorOrNil(t *testing.T) {
	tests := []struct {
		in   *Validator
		want error
	}{
		{&Validator{}, nil},
		{&Validator{Errors: map[string][]string{}}, nil},
		{
			&Validator{Errors: map[string][]string{"x": []string{"X"}}},
			&Validator{Errors: map[string][]string{"x": []string{"X"}}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			got := tt.in.ErrorOrNil()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", got, tt.want)
			}
		})
	}
}

func TestPop(t *testing.T) {
	v := New()
	v.Append("a", "err")
	v.Append("a", "err2")
	v.Append("b", "err3")

	{ // Non-existing key.
		out := v.Pop("nonexistent")
		var want []string
		if !reflect.DeepEqual(out, want) {
			t.Errorf("wrong return\nout:  %#v\nwant: %#v", out, want)
		}

		wantErr := map[string][]string{"a": {"err", "err2"}, "b": {"err3"}}
		if d := cmp.Diff(v.Errors, wantErr); d != "" {
			t.Errorf("(-got +want)\n:%s", d)
		}
	}

	{ // pop "a"
		out := v.Pop("a")
		want := []string{"err", "err2"}
		if !reflect.DeepEqual(out, want) {
			t.Errorf("wrong return\nout:  %#v\nwant: <nil>", out)
		}

		wantErr := map[string][]string{"b": {"err3"}}
		if d := cmp.Diff(v.Errors, wantErr); d != "" {
			t.Errorf("(-got +want)\n:%s", d)
		}
	}

	{ // pop "a" again, nothing should happen.
		out := v.Pop("a")
		var want []string
		if !reflect.DeepEqual(out, want) {
			t.Errorf("wrong return\nout:  %#v\nwant: <nil>", out)
		}

		wantErr := map[string][]string{"b": {"err3"}}
		if d := cmp.Diff(v.Errors, wantErr); d != "" {
			t.Errorf("(-got +want)\n:%s", d)
		}
	}

	{ // pop "b.
		out := v.Pop("b")
		want := []string{"err3"}
		if !reflect.DeepEqual(out, want) {
			t.Errorf("wrong return\nout:  %#v\nwant: <nil>", out)
		}

		wantErr := map[string][]string{}
		if d := cmp.Diff(v.Errors, wantErr); d != "" {
			t.Errorf("(-got +want)\n:%s", d)
		}
	}

	if v.HasErrors() {
		t.Errorf("v.HasErrors(): %#v", v.Errors)
	}
}
