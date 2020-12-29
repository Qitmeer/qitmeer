/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package client

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc/websocket"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrInvalidAuth     = errors.New("authentication failure")
	ErrInvalidEndpoint = errors.New("the endpoint either does not support " +
		"websockets or does not exist")
	ErrClientShutdown     = errors.New("the client has been shutdown")
	ErrClientNotConnected = errors.New("the client was never connected")
	ErrClientDisconnect   = errors.New("the client has been disconnected")
)

const (
	// sendBufferSize is the number of elements the websocket send channel
	// can queue before blocking.
	sendBufferSize = 50

	// sendPostBufferSize is the number of elements the HTTP POST send
	// channel can queue before blocking.
	sendPostBufferSize = 100

	// connectionRetryInterval is the amount of time to wait in between
	// retries when automatically reconnecting to an RPC server.
	connectionRetryInterval = time.Second * 5
)

type sendPostDetails struct {
	httpRequest *http.Request
	jsonRequest *jsonRequest
}

// jsonRequest holds information about a json request that is used to properly
// detect, interpret, and deliver a reply to it.
type jsonRequest struct {
	id             uint64
	method         string
	cmd            interface{}
	marshalledJSON []byte
	responseChan   chan *response
}

type inMessage struct {
	ID *float64 `json:"id"`
	*rawNotification
	*rawResponse
}

type rawNotification struct {
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

func newHTTPClient(config *ConnConfig) (*http.Client, error) {
	// Configure TLS if needed.
	var tlsConfig *tls.Config
	if !config.DisableTLS {
		if len(config.Certificates) > 0 {
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(config.Certificates)
			tlsConfig = &tls.Config{
				RootCAs: pool,
			}
		}
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &client, nil
}

func dial(config *ConnConfig) (*websocket.Conn, error) {
	// Setup TLS if not disabled.
	var tlsConfig *tls.Config
	var scheme = "ws"
	if !config.DisableTLS {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if len(config.Certificates) > 0 {
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(config.Certificates)
			tlsConfig.RootCAs = pool
		}
		scheme = "wss"
	}

	// Create a websocket dialer that will be used to make the connection.
	// It is modified by the proxy setting below as needed.
	dialer := websocket.Dialer{TLSClientConfig: tlsConfig}

	// The RPC server requires basic authorization, so create a custom
	// request header with the Authorization header set.
	user, pass, err := config.getAuth()
	if err != nil {
		return nil, err
	}
	login := user + ":" + pass
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
	requestHeader := make(http.Header)
	requestHeader.Add("Authorization", auth)
	for key, value := range config.ExtraHeaders {
		requestHeader.Add(key, value)
	}

	// Dial the connection.
	url := fmt.Sprintf("%s://%s/%s", scheme, config.Host, config.Endpoint)
	wsConn, resp, err := dialer.Dial(url, requestHeader)
	if err != nil {
		if err != websocket.ErrBadHandshake || resp == nil {
			return nil, err
		}

		// Detect HTTP authentication error status codes.
		if resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == http.StatusForbidden {
			return nil, ErrInvalidAuth
		}

		// The connection was authenticated and the status response was
		// ok, but the websocket handshake still failed, so the endpoint
		// is invalid in some way.
		if resp.StatusCode == http.StatusOK {
			return nil, ErrInvalidEndpoint
		}

		// Return the status text from the server if none of the special
		// cases above apply.
		return nil, errors.New(resp.Status)
	}
	return wsConn, nil
}

func readCookieFile(path string) (username, password string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	err = scanner.Err()
	if err != nil {
		return
	}
	s := scanner.Text()

	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("malformed cookie file")
		return
	}

	username, password = parts[0], parts[1]
	return
}

type wrongNumParams int

// Error satisifies the builtin error interface.
func (e wrongNumParams) Error() string {
	return fmt.Sprintf("wrong number of parameters (%d)", e)
}

func receiveFuture(f chan *response) ([]byte, error) {
	// Wait for a response on the returned channel.
	r := <-f
	return r.result, r.err
}

func newFutureError(err error) chan *response {
	responseChan := make(chan *response, 1)
	responseChan <- &response{err: err}
	return responseChan
}

func newNilFutureResult() chan *response {
	responseChan := make(chan *response, 1)
	responseChan <- &response{result: nil, err: nil}
	return responseChan
}
