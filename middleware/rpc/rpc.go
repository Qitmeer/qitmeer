package rpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

const url = "https://127.0.0.1:1234"

//Params类型不确定，可以是数组，也可以是map参数
type RequestData struct {
	JsonRpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int64       `json:"id"`
}

//Result类型不确定，可以是string，也可以是int等类型
type RpcResult struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      int64       `json:"id"`
	Result  interface{} `json:"result"`
}

func (rpcResult *RpcResult) SendRpc(data RequestData) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	dataByte, _ := json.Marshal(data)

	httpRequest, err := http.NewRequest("POST", url, bytes.NewReader(dataByte))
	if err != nil {
		log.Fatal(err)
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.SetBasicAuth("test", "test")

	response, err := client.Do(httpRequest)
	body := response.Body

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatal("io read error")
	}

	str:=string(bodyBytes)

	log.Println(str)

	unmarshalError := json.Unmarshal(bodyBytes, rpcResult)
	if unmarshalError != nil {
		log.Fatal(unmarshalError)
	}
}

func (response *RpcResult) ToString() string {
	respByte, err := json.Marshal(response)
	if (err != nil) {
		log.Fatal("Marshal error")
	}
	return string(respByte)
}
