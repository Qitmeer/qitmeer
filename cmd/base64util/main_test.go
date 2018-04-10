// Copyright 2017-2018 The nox developers
package main

import (
	"io/ioutil"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
)

func Example1HexString() {
	args = []string{"-hex", "01020304", "-D", "std"}
	main()
	// Output:
	// AQIDBA==
}

func Example2HexString() {
	args = []string{"-D","std","04030201"}
	main()
	// Output:
	// BAMCAQ==
}

func Test1HexString(t *testing.T) {
	args = []string{"-hex", "01", "-D", "std"}
	main()
}

func Test2HexString(t *testing.T) {
	args = []string{"-D","std","02"}
	main()
}


func TestDebugStd(t *testing.T) {
	args = []string{"-hex", "03", "-D", "std"}
	main()
}

func TestDebugLog(t *testing.T) {
	args = []string{"-hex", "04", "-D", "log"}
	main()
}

func TestPipe(t *testing.T) {
	r, w, _ := os.Pipe()
	stdin := os.Stdin
	defer func() {
		r.Close()
		os.Stdin = stdin
	}()
	w.WriteString("05")
	w.Close()
	os.Stdin = r
	args = []string{"-D", "std"}
	main()
}

func TestStdInFile(t *testing.T) {
	content := []byte("06")
	// create a tmp file & write content
	tmpfile, _ := ioutil.TempFile("", "base64test")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	// open tmpfile
	f, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	stdin := os.Stdin
	defer func() {
		os.Stdin = stdin
		f.Close()
		os.Remove(tmpfile.Name())
	}()
	// wire file to stdin
	os.Stdin = f
	args = []string{"-D", "std"}
	main()
}

func TestHexStr2base64Str(t *testing.T) {

	hexStr1 := "01020304"
	hexStr1_enc := "AQIDBA=="

	hexStr2 := "f90260f901f9a083cafc574e1f51ba9dc0568fc617a08ea2429fb384059c972f13b19fa1c8dd55a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948888f1f195afa192cfee860698584c030f4c9db1a0ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017a05fe50b260da6308036625b850b5d6ced6d0a9f814c0688bc91ffb7b7a3a54b67a0bc37d79753ad738a6dac4921e57392f145d8887476de3f783dfa7edae9283e52b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008302000001832fefd8825208845506eb0780a0bd4472abb6659ebe3ee06ee4d7b72a00a9f4d001caca51342001075469aff49888a13a5a8c8f2bb1c4f861f85f800a82c35094095e7baea6a6c7c4c2dfeb977efac326af552d870a801ba09bea4c4daac7c7c52e093e6a4c35dbbcf8856f1af7b059ba20253e70848d094fa08a8fae537ce25ed8cb5af9adac3f141af69bd515bd2ba031522df09b97dd72b1c0"
	hexStr2_enc := "+QJg+QH5oIPK/FdOH1G6ncBWj8YXoI6iQp+zhAWcly8TsZ+hyN1VoB3MTejex116q4W1Z7bM1BrTEkUblIp0E/ChQv1A1JNHlIiI8fGVr6GSz+6GBphYTAMPTJ2xoO8VUqQLcWXDzXc4BrngwWW3U1bgMUvwcG8nnHKfUeAXoF/lCyYNpjCANmJbhQtdbO1tCp+BTAaIvJH/t7ejpUtnoLw315dTrXOKbaxJIeVzkvFF2Ih0dt4/eD36ftrpKD5SuQEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIMCAAABgy/v2IJSCIRVBusHgKC9RHKrtmWevj7gbuTXtyoAqfTQAcrKUTQgAQdUaa/0mIihOlqMjyuxxPhh+F+ACoLDUJQJXnuupqbHxMLf65d++sMmr1UthwqAG6Cb6kxNqsfHxS4JPmpMNdu8+IVvGvewWbogJT5whI0JT6CKj65TfOJe2Mta+a2sPxQa9pvVFb0roDFSLfCbl91yscA="

	tests := []struct {
		hex    string
		base64 string
	}{
		{hexStr1, hexStr1_enc},
		{hexStr2, hexStr2_enc},
	}

	for _, v := range tests {
		enc, _ := HexStr2base64Str(v.hex)
		assert.Equal(t, enc, v.base64)
	}
}
