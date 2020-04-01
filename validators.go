package zvalidate

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Required validates that the value is not the type's zero value.
//
// Currently supported types are string, int, int64, uint, uint64, bool,
// []string, and mail.Address. It will panic if the type is not supported.
func (v *Validator) Required(key string, value interface{}, message ...string) {
	msg := getMessage(message, MessageRequired)

	if value == nil {
		v.Append(key, msg)
		return
	}

check:
	switch val := value.(type) {
	default:
		// This is an appropiate use of panic, as it's a programming error that
		// should be displayed ASAP. Adding a "validation error" would be
		// inappropriate, and returning an error cumbersome.
		panic(fmt.Sprintf("zvalidate: not a supported type: %T", value))

	case string:
		if strings.TrimSpace(val) == "" {
			v.Append(key, msg)
		}
	case int:
		if val == int(0) {
			v.Append(key, msg)
		}
	case int64:
		if val == int64(0) {
			v.Append(key, msg)
		}
	case uint:
		if val == uint(0) {
			v.Append(key, msg)
		}
	case uint64:
		if val == uint64(0) {
			v.Append(key, msg)
		}
	case bool:
		if !val {
			v.Append(key, msg)
		}

	case []int64:
		if len(val) == 0 {
			v.Append(key, msg)
		}
	case []string:
		if len(val) == 0 {
			v.Append(key, msg)
			return
		}

		// Make sure there is at least one non-empty entry.
		nonEmpty := false
		for i := range val {
			if val[i] != "" { // Consider " " to be non-empty on purpose.
				nonEmpty = true
				break
			}
		}
		if !nonEmpty {
			v.Append(key, msg)
		}

	case mail.Address:
		if val.Address == "" {
			v.Append(key, msg)
		}
	case time.Time:
		if val.IsZero() {
			v.Append(key, msg)
		}

	case *string:
		value = *val
		goto check
	case *int:
		value = *val
		goto check
	case *int64:
		value = *val
		goto check
	case *uint:
		value = *val
		goto check
	case *uint64:
		value = *val
		goto check
	case *mail.Address:
		value = *val
		goto check
	case *time.Time:
		value = *val
		goto check
	}
}

// Exclude validates that the value is not in the exclude list.
//
// This list is matched case-insensitive.
func (v *Validator) Exclude(key, value string, exclude []string, message ...string) {
	msg := getMessage(message, "")

	value = strings.TrimSpace(strings.ToLower(value))
	for _, e := range exclude {
		if strings.EqualFold(e, value) {
			if msg != "" {
				v.Append(key, msg)
			} else {
				v.Append(key, fmt.Sprintf(MessageExclude, e))
			}
			return
		}
	}
}

// Include validates that the value is in the include list.
//
// This list is matched case-insensitive.
func (v *Validator) Include(key, value string, include []string, message ...string) {
	if len(include) == 0 {
		return
	}

	value = strings.TrimSpace(strings.ToLower(value))
	for _, e := range include {
		if strings.EqualFold(e, value) {
			return
		}
	}

	msg := getMessage(message, "")
	if msg != "" {
		v.Append(key, msg)
	} else {
		v.Append(key, fmt.Sprintf(MessageInclude, strings.Join(include, ", ")))
	}
}

// Range sets the minimum and maximum value of a integer.
//
// A maximum of 0 indicates there is no upper limit.
func (v *Validator) Range(key string, value, min, max int64, message ...string) {
	msg := getMessage(message, "")

	if value < min {
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageRangeHigher, min))
		}
	}
	if max > 0 && value > max {
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageRangeLower, max))
		}
	}
}

// Domain parses a domain as individual labels.
//
// A domain must consist of at least two labels. So "com" or "localhost" – while
// technically valid domain names – are not accepted, whereas "example.com" or
// "me.localhost" are. For the overwhelming majority of applications this makes
// the most sense.
//
// This works for internationalized domain names (IDN), either as UTF-8
// characters or as punycode.
func (v *Validator) Domain(key, value string, message ...string) []string {
	if value == "" {
		return nil
	}

	msg := getMessage(message, MessageDomain)
	labels, err := validDomain(value, 2)
	if err != nil {
		v.Append(key, fmt.Sprintf("%s: %s", msg, err))
	}
	return labels
}

// Hostname checks if this is a valid hostname.
//
// This is different from Domain in that it considers any hostname valid,
// whereas Domain() is a bit stricter and validates if this is likely to be a
// publicly accessible domain. e.g. "localhost" is valid in Hostname(), but not
// Domain().
func (v *Validator) Hostname(key, value string, message ...string) []string {
	if value == "" {
		return nil
	}

	msg := getMessage(message, MessageHostname)
	labels, err := validDomain(value, 1)
	if err != nil {
		v.Append(key, fmt.Sprintf("%s: %s", msg, err))
	}
	return labels
}

func validDomain(value string, minLabels int) ([]string, error) {
	if len(value) < 3 || value[0] == '.' {
		return nil, fmt.Errorf("too short")
	}
	if value[len(value)-1] == '.' {
		value = value[:len(value)-1]
	}

	labels := strings.Split(value, ".")
	if len(labels) < minLabels {
		return nil, fmt.Errorf("need at least %d labels", minLabels)
	}

	var total int
	for i, l := range labels {
		// See RFC 1034, section 3.1, RFC 1035, secion 2.3.1
		//
		// - Only allow letters, numbers
		// - Max size of a single label is 63 bytes
		// - Need at least two labels
		if len(l) > 63 {
			return nil, errors.New("label is longer than 63 bytes")
		}
		total += len(l)

		if strings.HasPrefix(l, "xn--") {
			var err error
			l, err = punyDecode(l[4:])
			if err != nil {
				return nil, fmt.Errorf("not valid punycode: %q", l)
			}
			labels[i] = l
		}

		for _, c := range l {
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' {
				return nil, fmt.Errorf("invalid character: %q", c)
			}
		}
	}

	if total > 255 {
		return nil, fmt.Errorf("domain is longer than 255 bytes")
	}

	return labels, nil
}

// URL parses an URL.
//
// The URL may consist of a scheme, host, path, and query parameters. Only the
// host is required.
//
// The host is validated with the Domain() validation. If the scheme is not
// given "http" will be prepended.
func (v *Validator) URL(key, value string, message ...string) *url.URL {
	if value == "" {
		return nil
	}

	msg := getMessage(message, MessageURL)

	u, err := url.Parse(value)
	if err != nil && u == nil {
		v.Append(key, "%s: %s", msg, err)
		return nil
	}

	// If we don't have a scheme the parse may or may not fail according to the
	// go docs. "Trying to parse a hostname and path without a scheme is invalid
	// but may not necessarily return an error, due to parsing ambiguities."
	if u.Scheme == "" {
		u.Scheme = "http"
		u, err = url.Parse(u.String())
	}

	if err != nil {
		v.Append(key, "%s: %s", msg, err)
		return nil
	}

	if u.Host == "" {
		v.Append(key, msg)
		return nil
	}

	host := u.Host
	if h, _, err := net.SplitHostPort(u.Host); err == nil {
		host = h
	}

	_, err = validDomain(host, 2)
	if err != nil {
		v.Append(key, msg)
		return nil
	}

	return u
}

// Email parses an email address.
func (v *Validator) Email(key, value string, message ...string) mail.Address {
	if value == "" {
		return mail.Address{}
	}

	msg := getMessage(message, MessageEmail)
	addr, err := mail.ParseAddress(value)
	if err != nil {
		v.Append(key, msg)
		return mail.Address{}
	}

	// "foo@domain" is technically valid, but practically never what's intended.
	_, err = validDomain(addr.Address[strings.LastIndex(addr.Address, "@")+1:], 2)
	if err != nil {
		v.Append(key, msg)
		return mail.Address{}
	}

	return *addr
}

// IPv4 parses an IPv4 address.
func (v *Validator) IPv4(key, value string, message ...string) net.IP {
	if value == "" {
		return net.IP{}
	}

	msg := getMessage(message, MessageIPv4)
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() == nil {
		v.Append(key, msg)
	}
	return ip
}

// IP parses an IPv4 or IPv6 address.
func (v *Validator) IP(key, value string, message ...string) net.IP {
	if value == "" {
		return net.IP{}
	}

	msg := getMessage(message, MessageIP)
	ip := net.ParseIP(value)
	if ip == nil {
		v.Append(key, msg)
	}
	return ip
}

// HexColor parses a color as a hex triplet (e.g. #ffffff or #fff).
func (v *Validator) HexColor(key, value string, message ...string) (uint8, uint8, uint8) {
	if value == "" {
		return 0, 0, 0
	}

	msg := getMessage(message, MessageHexColor)

	if value[0] != '#' {
		v.Append(key, msg)
		return 0, 0, 0
	}

	var rgb []byte
	if len(value) == 4 {
		value = "#" +
			strings.Repeat(string(value[1]), 2) +
			strings.Repeat(string(value[2]), 2) +
			strings.Repeat(string(value[3]), 2)
	}

	n, err := fmt.Sscanf(strings.ToLower(value), "#%x", &rgb)
	if n != 1 || len(rgb) != 3 || err != nil {
		v.Append(key, msg)
		return 0, 0, 0
	}

	return rgb[0], rgb[1], rgb[2]
}

// UTF8 validates that this string is valid UTF-8.
//
// Caveat: this will consider NULL bytes *invalid* even though they're valid in
// UTF-8. Many tools don't accept it (e.g. PostgreSQL and SQLite), there's very
// rarely a reason to include them in strings, and most uses I've seen is from
// people trying to insert exploits. So the practical thing to do is just to
// reject it.
func (v *Validator) UTF8(key, value string, message ...string) {
	msg := getMessage(message, MessageUTF8)
	if !validString(value) {
		v.Append(key, msg)
	}
}

// Len validates the character (rune) length of a string.
//
// A maximum of 0 indicates there is no upper limit.
func (v *Validator) Len(key, value string, min, max int, message ...string) int {
	msg := getMessage(message, "")

	l := utf8.RuneCountInString(value)
	switch {
	case l < min:
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageLenLonger, min))
		}
	case max > 0 && l > max:
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageLenShorter, max))
		}
	}
	return l
}

// Integer parses a string as an integer.
func (v *Validator) Integer(key, value string, message ...string) int64 {
	if value == "" {
		return 0
	}

	i, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		v.Append(key, getMessage(message, MessageInteger))
	}
	return i
}

// Boolean parses as string as a boolean.
func (v *Validator) Boolean(key, value string, message ...string) bool {
	if value == "" {
		return false
	}

	switch strings.ToLower(value) {
	case "1", "y", "yes", "t", "true":
		return true
	case "0", "n", "no", "f", "false":
		return false
	}
	v.Append(key, getMessage(message, MessageBool))
	return false
}

// Date parses a string in the given date layout.
func (v *Validator) Date(key, value, layout string, message ...string) time.Time {
	if value == "" {
		return time.Time{}
	}

	msg := getMessage(message, "")
	t, err := time.Parse(layout, value)
	if err != nil {
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageDate, layout))
		}
	}
	return t
}

var rePhone = regexp.MustCompile(`^[0123456789+\-() .]{5,20}$`)

// Phone parses a phone number.
//
// There are a great amount of writing conventions for phone numbers:
// https://en.wikipedia.org/wiki/National_conventions_for_writing_telephone_numbers
//
// This merely checks a field contains 5 to 20 characters "0123456789+\-() .",
// which is not very strict but should cover all conventions.
//
// Returns the phone number with grouping/spacing characters removed.
func (v *Validator) Phone(key, value string, message ...string) string {
	if value == "" {
		return ""
	}

	msg := getMessage(message, MessagePhone)
	if !rePhone.MatchString(value) {
		v.Append(key, msg)
	}

	return strings.NewReplacer("-", "", "(", "", ")", "", " ", "", ".", "").
		Replace(value)
}
