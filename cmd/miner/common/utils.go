/**
Qitmeer
james
*/

package common

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"log"
	"math"
	"math/big"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func BlockBitsToTarget(bits string, width int) []byte {
	nbits, err := hex.DecodeString(bits[0:2])
	if err != nil {
		fmt.Println("error", err.Error())
	}
	shift := nbits[0] - 3
	value, _ := hex.DecodeString(bits[2:])
	target0 := make([]byte, int(shift))
	tmp := string(value) + string(target0)
	target1 := []byte(tmp)
	if len(target1) < width {
		head := make([]byte, width-len(target1))
		target := string(head) + string(target1)
		return []byte(target)
	}
	return target1
}

func Int2varinthex(x int64) string {
	if x < 0xfd {
		return fmt.Sprintf("%02x", x)
	} else if x < 0xffff {
		return "fd" + Int2lehex(x, 2)
	} else if x < 0xffffffff {
		return "fe" + Int2lehex(x, 4)
	} else {
		return "ff" + Int2lehex(x, 8)
	}
}
func Int2lehex(x int64, width int) string {
	if x <= 0 {
		return fmt.Sprintf("%016x", x)
	}
	bs := make([]byte, width)
	switch width {
	case 2:
		binary.LittleEndian.PutUint16(bs, uint16(x))
	case 4:
		binary.LittleEndian.PutUint32(bs, uint32(x))
	case 8:
		binary.LittleEndian.PutUint64(bs, uint64(x))
	}
	return hex.EncodeToString(bs)
}

// Reverse reverses a byte array.
func Reverse(src []byte) []byte {
	dst := make([]byte, len(src))
	for i := len(src); i > 0; i-- {
		dst[len(src)-i] = src[i-1]
	}
	return dst
}

// FormatHashRate sets the units properly when displaying a hashrate.
func FormatHashRate(h float64, unit string) string {
	if h > 1000000000000 {
		return fmt.Sprintf("%.2f T%s", h/1000000000000, unit)
	} else if h > 1000000000 {
		return fmt.Sprintf("%.2f G%s", h/1000000000, unit)
	} else if h > 1000000 {
		return fmt.Sprintf("%.2f M%s", h/1000000, unit)
	} else if h > 1000 {
		return fmt.Sprintf("%.2f k%s", h/1000, unit)
	} else if h == 0 {
		return fmt.Sprintf("0%s", unit)
	} else if h > 0 {
		return fmt.Sprintf("%.2f %s", h, unit)
	}

	return fmt.Sprintf("%.2f T%s", h, unit)
}

func ReverseByWidth(s []byte, width int) []byte {
	newS := make([]byte, len(s))
	for i := 0; i < (len(s) / width); i += 1 {
		j := i * width
		copy(newS[len(s)-j-width:len(s)-j], s[j:j+width])
	}
	return newS
}

func DiffToTarget(diff float64, powLimit *big.Int, powType pow.PowType) (*big.Int, error) {
	if diff <= 0 {
		return nil, fmt.Errorf("invalid pool difficulty %v (0 or less than "+
			"zero passed)", diff)
	}

	// Round down in the case of a non-integer diff since we only support
	// ints (unless diff < 1 since we don't allow 0)..
	if diff < 1 {
		diff = 1
	} else {
		diff = math.Floor(diff)
	}
	divisor := new(big.Int).SetInt64(int64(diff))
	max := powLimit
	target := new(big.Int)
	if powType == pow.BLAKE2BD || powType == pow.X8R16 ||
		powType == pow.QITMEERKECCAK256 || powType == pow.X16RV3 || powType == pow.MEERXKECCAKV1 {
		target.Div(max, divisor)
	} else {
		target.Div(divisor, max)
	}

	return target, nil
}

// appDataDir returns an operating system specific directory to be used for
// storing application data for an application.  See AppDataDir for more
// details.  This unexported version takes an operating system argument
// primarily to enable the testing package to properly test the function by
// forcing an operating system that is not the currently one.
func appDataDir(goos, appName string, roaming bool) string {
	if appName == "" || appName == "." {
		return "."
	}

	// The caller really shouldn't prepend the appName with a period, but
	// if they do, handle it gracefully by stripping it.
	appName = strings.TrimPrefix(appName, ".")
	appNameUpper := string(unicode.ToUpper(rune(appName[0]))) + appName[1:]
	appNameLower := string(unicode.ToLower(rune(appName[0]))) + appName[1:]

	// Get the OS specific home directory via the Go standard lib.
	var homeDir string
	usr, err := user.Current()
	if err == nil {
		homeDir = usr.HomeDir
	}

	// Fall back to standard HOME environment variable that works
	// for most POSIX OSes if the directory from the Go standard
	// lib failed.
	if err != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}

	switch goos {
	// Attempt to use the LOCALAPPDATA or APPDATA environment variable on
	// Windows.
	case "windows":
		// Windows XP and before didn't have a LOCALAPPDATA, so fallback
		// to regular APPDATA when LOCALAPPDATA is not set.
		appData := os.Getenv("LOCALAPPDATA")
		if roaming || appData == "" {
			appData = os.Getenv("APPDATA")
		}

		if appData != "" {
			return filepath.Join(appData, appNameUpper)
		}

	case "darwin":
		if homeDir != "" {
			return filepath.Join(homeDir, "Library",
				"Application Support", appNameUpper)
		}

	case "plan9":
		if homeDir != "" {
			return filepath.Join(homeDir, appNameLower)
		}

	default:
		if homeDir != "" {
			return filepath.Join(homeDir, "."+appNameLower)
		}
	}

	// Fall back to the current directory if all else fails.
	return "."
}

func GetCurrentDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func RandUint64() uint64 {
	return rand.Uint64()
}

func InArray(val interface{}, arr interface{}) bool {
	switch arr.(type) {
	case []string:
		for _, v := range arr.([]string) {
			if v == val {
				return true
			}
		}
	case []int:
		for _, v := range arr.([]int) {
			if v == val {
				return true
			}
		}
	}

	return false
}

func GetNeedHashTimesByTarget(target string) *big.Int {
	times := big.NewInt(1)
	for i := 0; i < len(target)-1; i++ {
		tmp := target[i : i+1]
		if tmp == "0" {
			times.Lsh(times, 4)
		} else {
			n, _ := strconv.ParseUint(tmp, 16, 32)
			if n <= 1 {
				n = 1
			}
			n1 := int64(16 / n)
			times.Mul(times, big.NewInt(n1))
			break
		}
	}
	return times
}

func Timeout(fun func(), t int64, callBack func()) {
	tim := time.NewTicker(time.Duration(t) * time.Second)
	defer tim.Stop()
	complete := make(chan int)
	go func() {
		fun()
		complete <- 1
	}()
	select {
	case <-complete:
		return
	case <-tim.C:
		callBack()
		MinerLoger.Warn("ws timeout!!!")
		return
	}
}
