package consensus

import (
	"math/big"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/blockchain"
)

// Any ConsensusState
type ConsensusState interface {
}

type ConsensusAlgorithm interface {
	SetState(state ConsensusState) (ConsensusState, error)
}

// agree on a consensus state
type Consensus interface {
	GetCurrentState() (ConsensusState, error)
	Commit(state ConsensusState) (ConsensusState, error)
	SetAlgorithm(algorithm ConsensusAlgorithm)
}

type Operation interface {
	ApplyTo(state ConsensusState)(ConsensusState, error)
}

// agree on a operation that update the consensus state
type Agreement interface {
	Commit(operation Operation) (ConsensusState, error)
	GetHeadState() (ConsensusState, error)
	Rollback(state ConsensusState) error
}


// the algorithm agnostic consensus engine.
type BlockChainConsensue interface {

	// VerifySeal checks whether the crypto seal on a header is valid according to
	// the consensus rules of the given engine.
	Verify(chain blockchain.BlockChain, header *types.BlockHeader) error

	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(chain blockchain.BlockChain, header *types.BlockHeader) error

	// Finalize runs any post-transaction state modifications (e.g. block rewards)
	// and assembles the final block.
	// Note: The block header and state database might be updated to reflect any
	// consensus rules that happen at finalization (e.g. block rewards).
	Finalize() (*types.Block, error)

	// Generates a new block for the given input block with the local miner's
	// seal place on top.
	Generate(chain blockchain.BlockChain, block *types.Block, stop <-chan struct{}) (*types.Block, error)

}
// PoW is a consensus engine based on proof-of-work.
type PoW interface {

  	BlockChainConsensue
	// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
	// that a new block should have.
	CalcDifficulty(chain blockchain.BlockChain, time uint64, parent *types.BlockHeader) *big.Int

	// Hashrate returns the current mining hashrate of a PoW consensus engine.
	Hashrate() float64
}

type PoA interface {
	BlockChainConsensue
}
