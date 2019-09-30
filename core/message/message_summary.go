package message

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/engine/txscript"
	"strings"
	"time"
)

const (
	// maxRejectReasonLen is the maximum length of a sanitized reject reason
	// that will be logged.
	maxRejectReasonLen = 250
)

// message.Summary returns a human-readable string which summarizes a message.
// Not all messages have or need a summary.  This is used for debug logging.
func Summary(msg Message) string {
	switch msg := msg.(type) {
	case *MsgVersion:
		return fmt.Sprintf("agent %s, pver %d, gs %s",
			msg.UserAgent, msg.ProtocolVersion, msg.LastGS.String())

	case *MsgVerAck:
		// No summary.

	case *MsgGetAddr:
		// No summary.

	case *MsgAddr:
		return fmt.Sprintf("%d addr", len(msg.AddrList))

	case *MsgPing:
		// No summary - perhaps add nonce.

	case *MsgPong:
		// No summary - perhaps add nonce.

	case *MsgTx:
		return fmt.Sprintf("hash %s, %d inputs, %d outputs, lock %s",
			msg.Tx.TxHash(), len(msg.Tx.TxIn), len(msg.Tx.TxOut),
			formatLockTime(msg.Tx.LockTime))
	// TODO
	/*
	case *MsgMemPool:
		// No summary.
	case *MsgBlock:
		header := &msg.Header
		return fmt.Sprintf("hash %s, ver %d, %d tx, %s", msg.BlockHash(),
			header.Version, len(msg.Transactions), header.Timestamp)

	case *MsgInv:
		return invSummary(msg.InvList)

	case *MsgNotFound:
		return invSummary(msg.InvList)

	case *MsgGetData:
		return invSummary(msg.InvList)

	case *MsgGetBlocks:
		return locatorSummary(msg.BlockLocatorHashes, &msg.HashStop)
*/
	case *MsgGetHeaders:
		return msg.String()

	case *MsgHeaders:
		return msg.String()

	case *MsgReject:
		// Ensure the variable length strings don't contain any
		// characters which are even remotely dangerous such as HTML
		// control characters, etc.  Also limit them to sane length for
		// logging.
		rejCommand := sanitizeString(msg.Cmd, CommandSize)
		rejReason := sanitizeString(msg.Reason, maxRejectReasonLen)
		summary := fmt.Sprintf("cmd %v, code %v, reason %v", rejCommand,
			msg.Code, rejReason)
		if rejCommand == CmdBlock || rejCommand == CmdTx {
			summary += fmt.Sprintf(", hash %v", msg.Hash)
		}
		return summary
	}

	// No summary for other messages.
	return ""
}

// formatLockTime returns a transaction lock time as a human-readable string.
func formatLockTime(lockTime uint32) string {
	// The lock time field of a transaction is either a block height at
	// which the transaction is finalized or a timestamp depending on if the
	// value is before the lockTimeThreshold.  When it is under the
	// threshold it is a block height.
	if lockTime < txscript.LockTimeThreshold {
		return fmt.Sprintf("height %d", lockTime)
	}

	return time.Unix(int64(lockTime), 0).String()
}

// sanitizeString strips any characters which are even remotely dangerous, such
// as html control characters, from the passed string.  It also limits it to
// the passed maximum size, which can be 0 for unlimited.  When the string is
// limited, it will also add "..." to the string to indicate it was truncated.
func sanitizeString(str string, maxLength uint) string {
	const safeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY" +
		"Z01234567890 .,;_/:?@"

	// Strip any characters not in the safeChars string removed.
	str = strings.Map(func(r rune) rune {
		if strings.ContainsRune(safeChars, r) {
			return r
		}
		return -1
	}, str)

	// Limit the string to the max allowed length.
	if maxLength > 0 && uint(len(str)) > maxLength {
		str = str[:maxLength]
		str = str + "..."
	}
	return str
}
