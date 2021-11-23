// Copyright (c) 2017-2018 The qitmeer developers

package rpc

import (
	"errors"
	"fmt"
	"github.com/Qitmeer/qng-core/core/types"
	"mime"
	"net/http"
)

const (
	contentType             = "application/json"
	maxRequestContentLength = types.MaxBlockPayload
)

// validateRequest returns a non-zero response code and error message if the
// request is invalid.
func validateRequest(r *http.Request) (int, error) {
	if r.Method == http.MethodPut || r.Method == http.MethodDelete {
		return http.StatusMethodNotAllowed, errors.New("method not allowed")
	}
	if r.ContentLength > maxRequestContentLength {
		err := fmt.Errorf("content length too large (%d>%d)", r.ContentLength, maxRequestContentLength)
		return http.StatusRequestEntityTooLarge, err
	}
	mt, _, err := mime.ParseMediaType(r.Header.Get("content-type"))
	if r.Method != http.MethodOptions && (err != nil || mt != contentType) {
		err := fmt.Errorf("invalid content type, only %s is supported", contentType)
		return http.StatusUnsupportedMediaType, err
	}
	return 0, nil
}

func emptyRequest(r *http.Request) bool {
	return r.Method == http.MethodGet && r.ContentLength == 0 && r.URL.RawQuery == ""
}
