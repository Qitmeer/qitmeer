package common

import (
	"time"
)

func Usleep(sec int) {
	time.Sleep(time.Second * time.Duration(sec))
}
