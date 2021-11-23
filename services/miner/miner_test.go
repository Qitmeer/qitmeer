package miner

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/types/pow"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"
)

type RpcClient struct {
	Server string
	User   string
	Pwd    string
}

// newHTTPClient returns a new HTTP client that is configured according to the
// proxy and TLS settings in the associated connection configuration.
func (rpc *RpcClient) newHTTPClient() (*http.Client, error) {
	// Configure proxy if needed.
	var dial func(network, addr string) (net.Conn, error)

	// Configure TLS if needed.
	var tlsConfig *tls.Config

	// Create and return the new HTTP client potentially configured with a
	// proxy and TLS.
	client := http.Client{
		Transport: &http.Transport{
			Dial:            dial,
			TLSClientConfig: tlsConfig,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
				DualStack: true,
			}).DialContext,
		},
	}
	return &client, nil
}

func (rpc *RpcClient) RpcResult(method string, params []interface{}) []byte {
	paramStr, err := json.Marshal(params)
	if err != nil {
		fmt.Println("rpc params error:", err)
		return nil
	}
	jsonStr := []byte(`{"jsonrpc": "2.0", "method": "` + method + `", "params": ` + string(paramStr) + `, "id": 1}`)
	bodyBuff := bytes.NewBuffer(jsonStr)
	httpRequest, err := http.NewRequest("POST", rpc.Server, bodyBuff)
	if err != nil {
		fmt.Println("rpc connect failed", err)
		return nil
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")
	// Configure basic access authorization.
	httpRequest.SetBasicAuth(rpc.User, rpc.Pwd)

	// Create the new HTTP client that is configured according to the user-
	// specified options and submit the request.
	httpClient, err := rpc.newHTTPClient()
	if err != nil {
		fmt.Println("rpc auth faild", err)
		return nil
	}
	httpClient.Timeout = 10 * time.Second
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		fmt.Println("rpc request faild", err)
		return nil
	}
	body, err := ioutil.ReadAll(httpResponse.Body)
	defer func() {
		_ = httpResponse.Body.Close()
	}()
	if err != nil {
		fmt.Println("error reading json reply:", err)
		return nil
	}

	if httpResponse.Status != "200 OK" {
		fmt.Println("error http response :", httpResponse.Status, body)
		return nil
	}
	return body
}

func TestMining(t *testing.T) {
	rpcC := &RpcClient{
		Server: "http://47.244.17.119:2234",
		User:   "test",
		Pwd:    "test",
	}
	for {
		b := rpcC.RpcResult("getBlockTemplate", []interface{}{
			[]string{"coinbasetxn", "coinbasevalue"}, 8,
		})
		type GbtTemplate struct {
			Result struct {
				WorkData         string `json:"workdata"`
				Height           int    `json:"height"`
				PowDiffReference struct {
					Target string `json:"target"`
				} `json:"pow_diff_reference"`
			} `json:"result"`
		}
		type Submit struct {
			Result string `json:"result"`
		}
		var gbtR GbtTemplate
		err := json.Unmarshal(b, &gbtR)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("New work height and target ", gbtR.Result.Height, gbtR.Result.PowDiffReference.Target)
		workData, err := hex.DecodeString(gbtR.Result.WorkData)
		if err != nil {
			fmt.Println(err)
			return
		}
		b1, _ := hex.DecodeString(gbtR.Result.PowDiffReference.Target)
		var r [32]byte
		copy(r[:], Reverse(b1)[:])
		r1 := hash.Hash(r)
		targetDiff := pow.HashToBig(&r1)
		start := time.Now().UnixNano()
		i := uint64(0)
		for ; i < ^uint64(0); i++ {
			nonce := make([]byte, 8)
			binary.LittleEndian.PutUint64(nonce, i)
			copy(workData[109:117], nonce)
			h := hash.HashMeerXKeccakV1(workData[:117])
			if pow.HashToBig(&h).Cmp(targetDiff) <= 0 {
				fmt.Println(i, hex.EncodeToString(nonce), h)
				// find hash
				b = rpcC.RpcResult("submitBlock", []interface{}{
					hex.EncodeToString(workData),
				})
				var subR Submit
				err = json.Unmarshal(b, &subR)
				if err != nil {
					fmt.Println(err)
					continue
				}
				break
			}
		}
		end := time.Now().UnixNano()
		gap := end - start
		hashrate := float64(i) / float64(gap)
		fmt.Printf("Hashrate: %.6f hash/s\n", 1/hashrate/1000)
	}
}

// Reverse reverses a byte array.
func Reverse(src []byte) []byte {
	dst := make([]byte, len(src))
	for i := len(src); i > 0; i-- {
		dst[len(src)-i] = src[i-1]
	}
	return dst
}
