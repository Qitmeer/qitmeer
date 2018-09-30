package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type txInputsFlag struct{
 	inputs []txInput
}

type txInput struct {
	txhash []byte
	index uint64
}

func (i txInput) String() string {
	return fmt.Sprintf("%x:%d",i.txhash[:],i.index)
}

func (v txInputsFlag) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for _,input := range v.inputs{
		buffer.WriteString(input.String())
	}
	buffer.WriteString("}")
	return buffer.String()
}

func (v *txInputsFlag) Set(s string) error {
	inputs := strings.Split(s,":")
	if cap(inputs) < 2 {
		return fmt.Errorf("error to parse tx input : %s",s)
	}
	data, err :=hex.DecodeString(inputs[0])
	if err!=nil{
		return err
	}
	if len(data) != 32 {
		return fmt.Errorf("tx hash should be 32 bytes")
	}

	index, err := strconv.ParseUint(inputs[1], 10, 64)
	if err != nil {
		return err
	}
	input := txInput{
		 data,
		index,
	}
	v.inputs = append(v.inputs,input)
	return nil
}
