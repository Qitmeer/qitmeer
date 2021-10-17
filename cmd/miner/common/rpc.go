package common

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/Qitmeer/qitmeer/cmd/miner/common/socks"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type RpcClient struct {
	Cfg      *GlobalConfig
	GbtID    int64
	SubmitID int64
}

// newHTTPClient returns a new HTTP client that is configured according to the
// proxy and TLS settings in the associated connection configuration.
func (rpc *RpcClient) newHTTPClient() (*http.Client, error) {
	// Configure proxy if needed.
	var dial func(network, addr string) (net.Conn, error)
	if rpc.Cfg.OptionConfig.Proxy != "" {
		proxy := &socks.Proxy{
			Addr:     rpc.Cfg.OptionConfig.Proxy,
			Username: rpc.Cfg.OptionConfig.ProxyUser,
			Password: rpc.Cfg.OptionConfig.ProxyPass,
		}
		dial = func(network, addr string) (net.Conn, error) {
			c, err := proxy.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			return c, nil
		}
	}

	// Configure TLS if needed.
	var tlsConfig *tls.Config
	if !rpc.Cfg.SoloConfig.NoTLS && rpc.Cfg.SoloConfig.RPCCert != "" {
		pem, err := ioutil.ReadFile(rpc.Cfg.SoloConfig.RPCCert)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(pem)
		tlsConfig = &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: rpc.Cfg.SoloConfig.NoTLS,
		}
	} else {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: rpc.Cfg.SoloConfig.NoTLS,
		}
	}

	// Create and return the new HTTP client potentially configured with a
	// proxy and TLS.
	client := http.Client{
		Transport: &http.Transport{
			Dial:            dial,
			TLSClientConfig: tlsConfig,
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(rpc.Cfg.OptionConfig.Timeout) * time.Second,
				KeepAlive: time.Duration(rpc.Cfg.OptionConfig.Timeout) * time.Second,
				DualStack: true,
			}).DialContext,
		},
	}
	return &client, nil
}

func (rpc *RpcClient) RpcResult(method string, params []interface{}, id string) []byte {
	protocol := "http"
	if !rpc.Cfg.SoloConfig.NoTLS {
		protocol = "https"
	}
	paramStr, err := json.Marshal(params)
	if err != nil {
		MinerLoger.Error("rpc params error", "error", err)
		return nil
	}
	url := rpc.Cfg.SoloConfig.RPCServer
	if !strings.Contains(rpc.Cfg.SoloConfig.RPCServer, "://") {
		url = protocol + "://" + url
	}
	jsonStr := []byte(`{"jsonrpc": "2.0", "method": "` + method +
		`", "params": ` + string(paramStr) + `, "id": "` + id + `"}`)
	bodyBuff := bytes.NewBuffer(jsonStr)
	httpRequest, err := http.NewRequest("POST", url, bodyBuff)
	if err != nil {
		MinerLoger.Error("rpc connect failed ", "error", err)
		return nil
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")
	// Configure basic access authorization.
	httpRequest.SetBasicAuth(rpc.Cfg.SoloConfig.RPCUser, rpc.Cfg.SoloConfig.RPCPassword)

	// Create the new HTTP client that is configured according to the user-
	// specified options and submit the request.
	httpClient, err := rpc.newHTTPClient()
	if err != nil {
		MinerLoger.Error("rpc auth faild ", "error", err)
		return nil
	}
	defer httpClient.CloseIdleConnections()
	httpClient.Timeout = time.Duration(rpc.Cfg.OptionConfig.Timeout) * time.Second
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		MinerLoger.Error("rpc request faild ", "error", err)
		return nil
	}
	defer func() {
		_ = httpResponse.Body.Close()
	}()
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		MinerLoger.Error("error reading json reply", "error", err)
		return nil
	}

	if httpResponse.StatusCode != 200 {
		time.Sleep(30 * time.Second)
		MinerLoger.Error("error http response", "status", httpResponse.Status, "body", string(body), "wait sec", 30)
		return nil
	}
	return body
}
