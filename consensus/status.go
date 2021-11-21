/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package consensus

import (
	"fmt"
)

type Status byte

const (
	Unknown Status = iota
	Processing
	Rejected
	Accepted
)

func (s Status) MarshalJSON() ([]byte, error) {
	if err := s.Valid(); err != nil {
		return nil, err
	}
	return []byte("\"" + s.String() + "\""), nil
}

func (s *Status) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" {
		return nil
	}
	switch str {
	case "\"Unknown\"":
		*s = Unknown
	case "\"Processing\"":
		*s = Processing
	case "\"Rejected\"":
		*s = Rejected
	case "\"Accepted\"":
		*s = Accepted
	default:
		return fmt.Errorf("unknown status:%s", str)
	}
	return nil
}

func (s Status) Fetched() bool {
	switch s {
	case Processing:
		return true
	default:
		return s.Decided()
	}
}

func (s Status) Decided() bool {
	switch s {
	case Rejected, Accepted:
		return true
	default:
		return false
	}
}

func (s Status) Valid() error {
	switch s {
	case Unknown, Processing, Rejected, Accepted:
		return nil
	default:
		return fmt.Errorf("unknown status")
	}
}

func (s Status) String() string {
	switch s {
	case Unknown:
		return "Unknown"
	case Processing:
		return "Processing"
	case Rejected:
		return "Rejected"
	case Accepted:
		return "Accepted"
	default:
		return "Invalid status"
	}
}

func (s Status) Bytes() []byte {
	return []byte{byte(s)}
}
