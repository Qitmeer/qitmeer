package lib

import (
	"bytes"
	"github.com/Qitmeer/qng-core/core/json"
	s "github.com/Qitmeer/qng-core/core/serialization"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/core/types/pow"
	"github.com/Qitmeer/qng-core/common/hash"
	"io"
	"sync"
)

//qitmeer block header
type BlockHeader struct {
	sync.Mutex
	// block version
	Version uint32 `json:"version"`
	// The merkle root of the previous parent blocks (the dag layer)
	ParentRoot hash.Hash `json:"previousblockhash"`
	// The merkle root of the tx tree  (tx of the block)
	// included Witness here instead of the separated witness commitment
	TxRoot hash.Hash `json:"tx_root"`
	// The Multiset hash of UTXO set or(?) merkle range/path or(?) tire tree root
	// can all of the state data (stake, receipt, utxo) in state root?
	StateRoot hash.Hash `json:"stateroot"`

	// block number
	Height     uint64 `json:"height"`
	Difficulty uint64 `json:"difficulty"`

	// TimeStamp
	Curtime uint32 `json:"curtime"`

	Pow pow.IPow

	// Nonce
	Target string `json:"target"`

	PowDiffReference json.PowDiffReference `json:"pow_diff_reference"`

	Coinbasevalue   int64 `json:"coinbasevalue"`
	HasCoinbasePack bool
	TotalFee        uint64
	transactions    []*types.Tx
	BlockFeesMap    map[int]int64 `json:"block_fees_map"`
	NodeInfo        string        `json:"nodeinfo"`
	CoinbaseVersion string        `json:"coinbase_version"`
	GBTID           int64
	WorkData        []byte
}

func (h *BlockHeader) SetTxs(transactions []*types.Tx) {
	h.transactions = transactions
}

//qitmeer block header
func BlockDataWithProof(h *types.BlockHeader) []byte {
	var buf bytes.Buffer
	// TODO, redefine the protocol version and storage
	_ = writeBlockHeaderWithProof(&buf, 0, h)
	return buf.Bytes()
}

func writeBlockHeaderWithProof(w io.Writer, pver uint32, bh *types.BlockHeader) error {
	sec := uint32(bh.Timestamp.Unix())
	return s.WriteElements(w, bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, bh.Difficulty, sec, bh.Pow.Bytes())
}

// readBlockHeader reads a block header from io reader.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the type.
// TODO, redefine the protocol version and storage
func ReadBlockHeader(b []byte, bh *types.BlockHeader) error {
	r := bytes.NewReader(b)
	// TODO fix time ambiguous
	return s.ReadElements(r, &bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, &bh.Difficulty, (*s.Uint32Time)(&bh.Timestamp),
		&bh.Pow)
}
