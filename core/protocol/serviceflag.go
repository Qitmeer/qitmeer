package protocol

import (
	"strconv"
	"strings"
)

// Map of service flags back to their constant names for pretty printing.
var sfStrings = map[ServiceFlag]string{
	Full: "Full",
	Bloom:   "Bloom",
	CF:      "CF",
}

// orderedSFStrings is an ordered list of service flags from highest to
// lowest.
var orderedSFStrings = []ServiceFlag{
	Full,
	Bloom,
	CF,
}

// String returns the ServiceFlag in human-readable form.
func (f ServiceFlag) String() string {
	// No flags are set.
	if f == 0 {
		return "0x0"
	}

	// Add individual bit flags.
	s := ""
	for _, flag := range orderedSFStrings {
		if f&flag == flag {
			s += sfStrings[flag] + "|"
			f -= flag
		}
	}

	// Add any remaining flags which aren't accounted for as hex.
	s = strings.TrimRight(s, "|")
	if f != 0 {
		s += "|0x" + strconv.FormatUint(uint64(f), 16)
	}
	s = strings.TrimLeft(s, "|")
	return s
}

// hasServices returns whether or not the provided advertised service flags have
// all of the provided desired service flags set.
func HasServices(advertised, desired ServiceFlag) bool {
	return advertised&desired == desired
}

// MissingServices returns what missing service flags from the advertised flags
// to the desired flags set
func MissingServices(advertised, desired ServiceFlag) ServiceFlag  {
	return desired & ^advertised
}
