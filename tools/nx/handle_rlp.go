// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"github.com/HalalChain/qitmeer/common/encode/rlp"
)

func rlpEncode(input string) {
	var f interface{}
	err := json.Unmarshal([]byte(input), &f)
	if err != nil {
		errExit(err)
	}
	b, err := rlp.EncodeToBytes(f)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",b)
}

func rlpDecode(input string) {
	data, err := hex.DecodeString(input)
	if err!=nil {
		errExit(err)
	}
	r := bytes.NewReader(data)
	s := rlp.NewStream(r, 0)
	var buffer bytes.Buffer
	for {
		if _, err := dump(s,0,&buffer); err != nil {
			if err != io.EOF {
				errExit(err)
			}
			break
		}
		fmt.Println(buffer.String())
	}
}

func dump(s *rlp.Stream, depth int, buffer *bytes.Buffer) (*bytes.Buffer, error) {
	kind, size, err := s.Kind()
	if err != nil {
		return buffer,err
	}
	switch kind {
	case rlp.Byte, rlp.String:
		str, err := s.Bytes()
		if err != nil {
			return buffer,err
		}
		if len(str) == 0 || isASCII(str) {

			buffer.WriteString(fmt.Sprintf("%s%q", ws(depth), str))
		} else {
			buffer.WriteString(fmt.Sprintf("%s%x", ws(depth), str))
		}
	case rlp.List:
		s.List()
		defer s.ListEnd()
		if size == 0 {
			buffer.WriteString(fmt.Sprint(ws(depth) + "[]"))
		} else {
			buffer.WriteString(fmt.Sprintln(ws(depth) + "["))
			for i := 0; ; i++ {
				if buff, err := dump(s, depth+1,buffer); err == rlp.EOL {
					buff.Truncate(buff.Len() - 2)  //remove the last comma
					buff.WriteString("\n")
					break
				} else if err != nil {
					return buff,err
				} else {
					buff.WriteString(",\n")
				}
			}
			buffer.WriteString(fmt.Sprint(ws(depth) + "]"))
		}
	}
	return buffer,nil
}

func isASCII(b []byte) bool {
	for _, c := range b {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

func ws(n int) string {
	return strings.Repeat("  ", n)
}
