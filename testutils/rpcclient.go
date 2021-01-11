// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"sync/atomic"
)

// Dial connects to a JSON-RPC server at the specified url,
// return a new client which can perform rpc Call. the context controls
// the entire lifetime of the client.
func Dial(URL, user, pass string, certs string) (*Client, error) {
	if u, err := url.Parse(URL); err == nil {
		switch u.Scheme {
		case "http", "https":
			return dailHTTP(URL, user, pass, certs)
		case "ws", "wss":
			//TODO websocket
			return nil, fmt.Errorf("websockect is unsupported yet")
		default:
			return nil, fmt.Errorf("unknown URL scheme: %v", u.Scheme)
		}
	} else {
		return nil, err
	}
}

const contentType = "application/json"

func dailHTTP(URL string, user, pass string, certs string) (*Client, error) {
	httpClient := new(http.Client)
	header := make(http.Header, 2)
	header.Set("accept", contentType) //TODO, refactor rpc.contentType
	header.Set("content-type", contentType)
	auth := "Basic " +
		base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
	header.Set("Authorization", auth)
	cert, err := ioutil.ReadFile(certs)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(cert)
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	}
	var httpReconn reconnect = func() (connect, error) {
		hc := &httpConn{
			client: httpClient,
			header: &header,
			url:    URL,
		}
		return hc, nil
	}
	return newClient(httpReconn)
}

type connType uint8

const (
	httpConnType connType = 0x0
)

type connect interface {
	Type() connType
}
type reconnect func() (connect, error)

// Client represents a connection to an RPC server.
type Client struct {
	// call when connection is lost, need to reconnect
	reconnect   reconnect
	connect     connect
	connectType connType
	idCounter   uint32
}

// the abstract connect
func newClient(reconn reconnect) (*Client, error) {
	if con, err := reconn(); err != nil {
		return nil, err
	} else {
		c := &Client{
			connect:     con,
			connectType: con.Type(),
			reconnect:   reconn,
			idCounter:   0,
		}
		return c, nil
	}
}

// Call wraps CallWithContext using the background context.
func (c *Client) Call(result interface{}, method string, args ...interface{}) error {
	ctx := context.Background()
	return c.CallWithContext(ctx, result, method, args...)
}

// CallWithContext performs a JSON-RPC call with the given arguments and unmarshalls into
// result if no error occurred.
func (c *Client) CallWithContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	// The result must be a pointer so that package json can unmarshal into it.
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("result must be pointer or nil interface: %v", result)
	}
	req, err := cmds.NewRequest(c.nextID(), method, args[:])
	op := &requestOp{ids: []uint32{req.ID.(uint32)}, resp: make(chan *cmds.Response, 1)}
	err = c.sendHTTP(ctx, op, req)
	if err != nil {
		return err
	}
	// dispatch has accepted the request and will close the channel when it quits.
	switch resp, err := op.wait(ctx, c); {
	case err != nil:
		return err
	case resp.Error != nil:
		return resp.Error
	case len(resp.Result) == 0:
		return fmt.Errorf("no result")
	default:
		return json.Unmarshal(resp.Result, &result)
	}
}

type httpConn struct {
	client *http.Client
	url    string
	mu     sync.Mutex // protects headers
	header *http.Header
}

func (hc *httpConn) Type() connType {
	return httpConnType
}

func (hc *httpConn) doRequest(ctx context.Context, msg interface{}) (io.ReadCloser, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", hc.url, ioutil.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(body))

	// set headers
	hc.mu.Lock()
	req.Header = hc.header.Clone()
	hc.mu.Unlock()

	// do request
	resp, err := hc.client.Do(req)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.Body, fmt.Errorf("%v", resp.Status)
	}
	return resp.Body, nil
}

func (c *Client) nextID() uint32 {
	id := atomic.AddUint32(&c.idCounter, 1)
	return id
}

func (c *Client) sendHTTP(ctx context.Context, op *requestOp, req interface{}) error {
	hc := c.connect.(*httpConn)
	respBody, err := hc.doRequest(ctx, req)
	if respBody != nil {
		defer respBody.Close()
	}

	if err != nil {
		if respBody != nil {
			buf := new(bytes.Buffer)
			if _, err2 := buf.ReadFrom(respBody); err2 == nil {
				return fmt.Errorf("%v: %v", err, buf.String())
			}
		}
		return err
	}
	var respmsg cmds.Response
	if err := json.NewDecoder(respBody).Decode(&respmsg); err != nil {
		return err
	}
	op.resp <- &respmsg
	return nil

}

type requestOp struct {
	ids  []uint32
	err  error
	resp chan *cmds.Response
}

func (op *requestOp) wait(ctx context.Context, c *Client) (*cmds.Response, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-op.resp:
		return resp, op.err
	}
}
