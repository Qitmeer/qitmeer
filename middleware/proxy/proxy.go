package main

import (
	"fmt"
	"github.com/noxproject/nox/middleware/rpc"
	"log"
	"net/http"
	"strconv"
)

func checkAndSend(r *http.Request, method string) rpc.RpcResult {
	r.ParseForm() //解析参数，默认是不会解析的
	params := r.PostFormValue("params")
	id := r.PostFormValue("id")
	idInt64, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	data := rpc.RequestData{
		JsonRpc: rpc.JSONRPC,
		Method:  method,
		Params:  params,
		Id:      idInt64,
	}

	result := rpc.RpcResult{}
	result.SendRpc(data)

	return result
}

func getBlockCount(w http.ResponseWriter, r *http.Request) {
	result := checkAndSend(r, rpc.GetBlockCount)
	resultString := result.ToString();
	fmt.Fprintln(w, resultString)
}

func getBlockInfo(w http.ResponseWriter, r *http.Request) {
	result := checkAndSend(r, rpc.GetBlockByHeight)
	resultString := result.ToString();
	fmt.Fprintln(w, resultString)
}

func main() {
	http.HandleFunc("/block/count", getBlockCount)
	http.HandleFunc("/block/info", getBlockInfo)

	err := http.ListenAndServe(":9998", nil)
	if err != nil {
		log.Fatal(err)
	}

}
