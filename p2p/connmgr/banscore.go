package connmgr

import (
	"fmt"
	"time"
)

type BanScore uint8

// Score of penalty node
const (
	// None
	NoneScore = 0

	// Sligh
	SlightScore = 1

	// A few
	FewScore = 5

	// Many
	ManyScore = 10

	// Excessive
	ExcessiveScore = 50

	// Serious
	SeriousScore = 100
)

var banScoreStrings = map[BanScore]string{
	NoneScore:      "NONE_SCORE",
	SlightScore:    "SLIGHT_SCORE",
	FewScore:       "FEW_SCORE",
	ManyScore:      "MANY_SCORE",
	ExcessiveScore: "EXCESSIVE_SCORE",
	SeriousScore:   "SERIOUS_SCORE",
}

func (bs BanScore) String() string {
	if s, ok := banScoreStrings[bs]; ok {
		return s
	}
	return fmt.Sprintf("Unknown Ban Score (%d)", uint8(bs))
}

const (
	defaultBanDuration  = time.Hour * 24
	defaultBanThreshold = 100
)

var BanDuration time.Duration = defaultBanDuration

var BanThreshold uint32 = defaultBanThreshold
