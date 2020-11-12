/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"bytes"
	"github.com/Qitmeer/qitmeer/p2p/encoder"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"io"
)

// response code
const (
	ResponseCodeSuccess        = byte(0x00)
	ResponseCodeInvalidRequest = byte(0x01)
	ResponseCodeServerError    = byte(0x02)
)

func (s *Sync) generateErrorResponse(code byte, reason string) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{code})
	resp := &pb.ErrorResponse{
		Message: []byte(reason),
	}
	if _, err := s.Encoding().EncodeWithMaxLength(buf, resp); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ReadRspCode response from a RPC stream.
func ReadRspCode(stream io.Reader, encoding encoder.NetworkEncoding) (uint8, string, error) {
	b := make([]byte, 1)
	_, err := stream.Read(b)
	if err != nil {
		return 0, "", err
	}

	if b[0] == ResponseCodeSuccess {
		return 0, "", nil
	}

	msg := &pb.ErrorResponse{
		Message: []byte{},
	}
	if err := encoding.DecodeWithMaxLength(stream, msg); err != nil {
		return 0, "", err
	}

	return b[0], string(msg.Message), nil
}
