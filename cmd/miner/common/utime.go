package common

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"time"
)

const ntpEpochOffset = 2208988800

type packet struct {
	Settings       uint8
	Stratum        uint8
	Poll           int8
	Precision      int8
	RootDelay      uint32
	RootDispersion uint32
	ReferenceID    uint32
	RefTimeSec     uint32
	RefTimeFrac    uint32
	OrigTimeSec    uint32
	OrigTimeFrac   uint32
	RxTimeSec      uint32
	RxTimeFrac     uint32
	TxTimeSec      uint32
	TxTimeFrac     uint32
}

func UpdateSysTime() {
	var timeLayoutStr = "2006-01-02 15:04:05"

	ntime := getremotetime()
	ts := ntime.Format(timeLayoutStr) //time转string
	fmt.Print(ts)
	// 2021-08-29 15:53:35.922579627 +0800 CST
	UpdateSystemDate(ts)
}

func getremotetime() time.Time {
	host := "time.asia.apple.com:123"
	conn, err := net.Dial("udp", host)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		log.Fatalf("failed to set deadline: %v", err)
	}

	req := &packet{Settings: 0x1B}

	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		log.Fatalf("failed to send request: %v", err)
	}

	rsp := &packet{}
	if err := binary.Read(conn, binary.BigEndian, rsp); err != nil {
		log.Fatalf("failed to read server response: %v", err)
	}

	secs := float64(rsp.TxTimeSec) - ntpEpochOffset
	nanos := (int64(rsp.TxTimeFrac) * 1e9) >> 32

	showtime := time.Unix(int64(secs), nanos)

	return showtime
}

func UpdateSystemDate(dateTime string) bool {
	cmd := exec.Command(`date`, `-s`, `"`+dateTime+`"`)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return false
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		log.Println(err)
		return false
	}
	_, err = ioutil.ReadAll(stdout)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
