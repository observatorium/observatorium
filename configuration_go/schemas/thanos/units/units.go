package units

import (
	"fmt"
	"strings"
)

// Taken from https://github.com/alecthomas/units/blob/b94a6e3cc13755c0a75fffecbb089eb346fc4289/bytes.go

type Bytes int64

// Base-2 byte units.
const (
	Kibibyte Bytes = 1024
	KiB            = Kibibyte
	Mebibyte       = Kibibyte * 1024
	MiB            = Mebibyte
	Gibibyte       = Mebibyte * 1024
	GiB            = Gibibyte
	Tebibyte       = Gibibyte * 1024
	TiB            = Tebibyte
	Pebibyte       = Tebibyte * 1024
	PiB            = Pebibyte
	Exbibyte       = Pebibyte * 1024
	EiB            = Exbibyte
)

func (b Bytes) String() string {
	return ToString(int64(b), 1024, "iB", "B")
}

var (
	siUnits = []string{"", "K", "M", "G", "T", "P", "E"}
)

func ToString(n int64, scale int64, suffix, baseSuffix string) string {
	mn := len(siUnits)
	out := make([]string, mn)
	for i, m := range siUnits {
		if n%scale != 0 || i == 0 && n == 0 {
			s := suffix
			if i == 0 {
				s = baseSuffix
			}
			out[mn-1-i] = fmt.Sprintf("%d%s%s", n%scale, m, s)
		}
		n /= scale
		if n == 0 {
			break
		}
	}
	return strings.Join(out, "")
}
