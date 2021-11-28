package zvalidate

type Messages struct {
	Required    func() string
	Domain      func() string
	Hostname    func() string
	URL         func() string
	Email       func() string
	IPv4        func() string
	IP          func() string
	HexColor    func() string
	LenLonger   func() string
	LenShorter  func() string
	Exclude     func() string
	Include     func() string
	Integer     func() string
	Bool        func() string
	Date        func() string
	Phone       func() string
	RangeHigher func() string
	RangeLower  func() string
	UTF8        func() string
	Contains    func() string
}

var DefaultMessages = Messages{
	Required:    func() string { return "must be set" },
	Domain:      func() string { return "must be a valid domain" },
	Hostname:    func() string { return "must be a valid hostname" },
	URL:         func() string { return "must be a valid url" },
	Email:       func() string { return "must be a valid email address" },
	IPv4:        func() string { return "must be a valid IPv4 address" },
	IP:          func() string { return "must be a valid IPv4 or IPv6 address" },
	HexColor:    func() string { return "must be a valid color code" },
	LenLonger:   func() string { return "must be longer than %d characters" },
	LenShorter:  func() string { return "must be shorter than %d characters" },
	Exclude:     func() string { return "cannot be ‘%s’" },
	Include:     func() string { return "must be one of ‘%s’" },
	Integer:     func() string { return "must be a whole number" },
	Bool:        func() string { return "must be a boolean" },
	Date:        func() string { return "must be a date as ‘%s’" },
	Phone:       func() string { return "must be a valid phone number" },
	RangeHigher: func() string { return "must be %d or higher" },
	RangeLower:  func() string { return "must be %d or lower" },
	UTF8:        func() string { return "must be UTF-8" },
	Contains:    func() string { return "cannot contain the characters %s" },
}

func (v Validator) getMessage(in []string, f func() string) string {
	switch len(in) {
	case 0:
		return f()
	case 1:
		return in[0]
	default:
		panic("can only pass one message")
	}
}
