/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:roughtime.go
 * Date:9/3/20 6:49 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package roughtime

import (
	"log"
	"time"

	"github.com/cloudflare/roughtime"
)

const RecalibrationInterval = time.Hour

var offset time.Duration

func Init() {
	log.Println("Init roughtime... ...")
	recalibrateRoughtime()
	runT := time.NewTimer(RecalibrationInterval)
	go func() {
		for {
			select {
			case <-runT.C:
				recalibrateRoughtime()
			}
		}
	}()
}

func recalibrateRoughtime() {
	t0 := time.Now()
	results := roughtime.Do(roughtime.Ecosystem, roughtime.DefaultQueryAttempts, roughtime.DefaultQueryTimeout, nil)
	// Compute the average difference between the system's time and the
	// Roughtime responses from the servers, rejecting responses whose radii
	// are larger than 2 seconds.
	var err error
	offset, err = roughtime.AvgDeltaWithRadiusThresh(results, t0, 2*time.Second)
	if err != nil {
		log.Printf("Failed to calculate roughtime offset, system time will be used by default.(%s)", err)
	} else {
		log.Printf("Roughtime offset:%s", offset)
	}
}

// Since returns the duration since t, based on the roughtime response
func Since(t time.Time) time.Duration {
	return Now().Sub(t)
}

// Until returns the duration until t, based on the roughtime response
func Until(t time.Time) time.Duration {
	return t.Sub(Now())
}

// Now returns the current local time given the roughtime offset.
func Now() time.Time {
	if offset <= 0 {
		return time.Now()
	}
	return time.Now().Add(offset)
}

func Offset() time.Duration {
	return offset
}
