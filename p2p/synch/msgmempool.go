package synch

// MsgMemPool implements the Message interface and represents a bitcoin mempool
// message.  It is used to request a list of transactions still in the active
// memory pool of a relay.
//
// This message has no payload and was not added until protocol versions
type MsgMemPool struct{}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgMemPool) MaxPayloadLength(pver uint32) uint32 {
	return 0
}
