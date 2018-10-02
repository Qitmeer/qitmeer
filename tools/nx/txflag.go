package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/common/encode/base58"
	"math"
	"strconv"
	"strings"
)
type txVersionFlag uint32
type txLockTimeFlag uint32

func (ver txVersionFlag) String() string {
	return fmt.Sprintf("%d",ver)
}
func (ver *txVersionFlag) Set(s string) error {
	v, err :=strconv.ParseUint(s, 10, 32)
	if err !=nil {
		return err
	}
	*ver = txVersionFlag(uint32(v))
	return nil
}

func (lt txLockTimeFlag) String() string {
	return fmt.Sprintf("%d",lt)
}
func (lt *txLockTimeFlag) Set(s string) error {
	v, err :=strconv.ParseUint(s, 10, 32)
	if err !=nil {
		return err
	}
	*lt = txLockTimeFlag(uint32(v))
	return nil
}

type txInputsFlag struct{
 	inputs []txInput
}
type txOutputsFlag struct{
	outputs []txOutput
}

type txInput struct {
	txhash []byte
	index uint64
	sequence uint32
}
type txOutput struct {
	target []byte
	amount uint64
}

func (i txInput) String() string {
	return fmt.Sprintf("%x:%d:%d",i.txhash[:],i.index,i.sequence)
}
func (o txOutput) String() string {
	return fmt.Sprintf("%x:%d",o.target[:],o.amount)
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

func(of txOutputsFlag) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for _,o := range of.outputs{
		buffer.WriteString(o.String())
	}
	buffer.WriteString("}")
	return buffer.String()
}

func (v *txInputsFlag) Set(s string) error {
	input := strings.Split(s,":")
	if len(input) < 2 {
		return fmt.Errorf("error to parse tx input : %s",s)
	}
	data, err :=hex.DecodeString(input[0])
	if err!=nil{
		return err
	}
	if len(data) != 32 {
		return fmt.Errorf("tx hash should be 32 bytes")
	}

	index, err := strconv.ParseUint(input[1], 10, 64)
	if err != nil {
		return err
	}
	var seq = uint32(math.MaxUint32)
	if len(input) == 3 {
		s, err := strconv.ParseUint(input[2], 10, 32)
		if err!= nil {
			return err
		}
		seq = uint32(s)
	}
	i := txInput{
		data,
		index,
		uint32(seq),
	}
	v.inputs = append(v.inputs,i)
	return nil
}

func (of *txOutputsFlag) Set(s string) error {
	output := strings.Split(s,":")
	if len(output) < 2 {
		return fmt.Errorf("error to parse tx output : %s",s)
	}
	target, _, err := base58.CheckDecode(output[0])
	if err!=nil{
		return err
	}
    amount, err := strconv.ParseUint(output[1], 10, 64)
	if err != nil {
		return err
	}
    of.outputs = append(of.outputs,txOutput{
    	target,amount })
	return nil
}

