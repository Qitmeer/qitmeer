// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qng-core/common/encode/rlp"
	"io"
	"strings"
)

func RlpEncode(input string) {
	var f interface{}
	err := json.Unmarshal([]byte(input), &f)
	if err != nil {
		ErrExit(err)
	}
	b, err := rlp.EncodeToBytes(f)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", b)
}

func RlpDecode(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	r := bytes.NewReader(data)
	s := rlp.NewStream(r, 0)
	var buffer bytes.Buffer
	for {
		if _, err := Dump(s, 0, &buffer); err != nil {
			if err != io.EOF {
				ErrExit(err)
			}
			break
		}
		fmt.Println(buffer.String())
	}
}

func Dump(s *rlp.Stream, depth int, buffer *bytes.Buffer) (*bytes.Buffer, error) {
	kind, size, err := s.Kind()
	if err != nil {
		return buffer, err
	}
	switch kind {
	case rlp.Byte, rlp.String:
		str, err := s.Bytes()
		if err != nil {
			return buffer, err
		}
		if len(str) == 0 || IsASCII(str) {

			buffer.WriteString(fmt.Sprintf("%s%q", Ws(depth), str))
		} else {
			buffer.WriteString(fmt.Sprintf("%s%x", Ws(depth), str))
		}
	case rlp.List:
		s.List()
		defer s.ListEnd()
		if size == 0 {
			buffer.WriteString(fmt.Sprint(Ws(depth) + "[]"))
		} else {
			buffer.WriteString(fmt.Sprintln(Ws(depth) + "["))
			for i := 0; ; i++ {
				if buff, err := Dump(s, depth+1, buffer); err == rlp.EOL {
					buff.Truncate(buff.Len() - 2) //remove the last comma
					buff.WriteString("\n")
					break
				} else if err != nil {
					return buff, err
				} else {
					buff.WriteString(",\n")
				}
			}
			buffer.WriteString(fmt.Sprint(Ws(depth) + "]"))
		}
	}
	return buffer, nil
}

func IsASCII(b []byte) bool {
	for _, c := range b {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

func Ws(n int) string {
	return strings.Repeat("  ", n)
}
