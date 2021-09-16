package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	_ "github.com/Qitmeer/qitmeer/database/ffldb"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"os"
)

func main() {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	cfg, _, err := LoadConfig()
	if err != nil {
		log.Error(err.Error())
		return
	}

	defer func() {
		if log.LogWrite() != nil {
			log.LogWrite().Close()
		}
	}()

	// Load the block database.
	db, err := LoadBlockDB(cfg)
	if err != nil {
		log.Error("load block database", "error", err)
		return
	}
	defer func() {
		// Ensure the database is sync'd and closed on shutdown.
		log.Info("Gracefully shutting down the database...")
		db.Close()
	}()

	// find
	bc, err := blockchain.New(&blockchain.Config{
		DB:          db,
		ChainParams: params.ActiveNetParams.Params,
		TimeSource:  blockchain.NewMedianTime(),
		DAGType:     cfg.DAGType,
	})
	if err != nil {
		log.Error(err.Error())
		return
	}
	if processIsCheckpoint(bc, cfg) {
		return
	}
	// Find checkpoint candidates.
	candidates, err := findCandidates(bc, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to identify candidates:", err)
		return
	}

	// No candidates.
	if len(candidates) == 0 {
		fmt.Println("No candidates found.")
		return
	}

	// Show the candidates.
	for i, checkpoint := range candidates {
		showCandidate(i+1, checkpoint, cfg)
	}
}

func findCandidates(chain *blockchain.BlockChain, cfg *Config) ([]*params.Checkpoint, error) {
	// Start with the latest block of the main chain.
	block := chain.BlockDAG().GetMainChainTip()
	if block == nil {
		return nil, fmt.Errorf("Can't find main chain tip")
	}

	// Get the latest known checkpoint.
	latestCheckpoint := chain.LatestCheckpoint()
	if latestCheckpoint == nil {
		// Set the latest checkpoint to the genesis block if there isn't
		// already one.
		latestCheckpoint = &params.Checkpoint{
			Hash:  params.ActiveNetParams.GenesisHash,
			Layer: 0,
		}
	}

	// The latest known block must be at least the last known checkpoint
	// plus required checkpoint confirmations.
	checkpointConfirmations := uint64(blockchain.CheckpointConfirmations)
	requiredLayer := latestCheckpoint.Layer + checkpointConfirmations
	if uint64(block.GetLayer()) < requiredLayer {
		return nil, fmt.Errorf("the block database is only at layer "+
			"%d which is less than the latest checkpoint height "+
			"of %d plus required confirmations of %d",
			block.GetLayer(), latestCheckpoint.Layer,
			checkpointConfirmations)
	}

	// For the first checkpoint, the required layer is any block after the
	// genesis block, so long as the chain has at least the required number
	// of confirmations (which is enforced above).
	if len(params.ActiveNetParams.Checkpoints) == 0 {
		requiredLayer = 1
	}

	// Indeterminate progress setup.
	numBlocksToTest := uint64(block.GetLayer()) - requiredLayer
	progressInterval := (numBlocksToTest / 100) + 1 // min 1
	fmt.Print("Searching for candidates")
	defer fmt.Println()

	// Loop backwards through the chain to find checkpoint candidates.
	candidates := make([]*params.Checkpoint, 0, cfg.NumCandidates)
	numTested := uint64(0)
	preblock := block
	for len(candidates) < cfg.NumCandidates && uint64(block.GetLayer()) > requiredLayer {
		// Display progress.
		if numTested%progressInterval == 0 {
			fmt.Print(".")
		}
		// Determine if this block is a checkpoint candidate.
		isCandidate, err := chain.IsCheckpointCandidate(preblock, block)
		if err != nil {
			return nil, err
		}

		// All checks passed, so this node seems like a reasonable
		// checkpoint candidate.
		if isCandidate {
			checkpoint := params.Checkpoint{
				Layer: uint64(block.GetLayer()),
				Hash:  block.GetHash(),
			}
			candidates = append(candidates, &checkpoint)
		}

		if block.GetMainParent() == blockdag.MaxId {
			break
		}
		preblock = block
		block = chain.BlockDAG().GetBlockById(block.GetMainParent())
		if block == nil {
			return nil, fmt.Errorf("Can't find block")
		}
		numTested++
	}
	return candidates, nil
}

func showCandidate(candidateNum int, checkpoint *params.Checkpoint, cfg *Config) {
	if cfg.UseGoOutput {
		fmt.Printf("Candidate %d -- {%d, newShaHashFromStr(\"%v\")},\n",
			candidateNum, checkpoint.Layer, checkpoint.Hash)
		return
	}

	fmt.Printf("Candidate %d -- Layer: %d, Hash: %v\n", candidateNum,
		checkpoint.Layer, checkpoint.Hash)

}

func processIsCheckpoint(chain *blockchain.BlockChain, cfg *Config) bool {
	if len(cfg.IsCheckPoint) == 0 {
		return false
	}
	blockhash, err := hash.NewHashFromStr(cfg.IsCheckPoint)
	if err != nil {
		log.Error(err.Error())
		return true
	}
	block := chain.GetBlock(blockhash)
	if block == nil {
		log.Error(fmt.Sprintf("%s is not check point", blockhash.String()))
		return true
	}
	checkpoints := chain.Checkpoints()
	if checkpoints != nil {
		for _, checkpoint := range checkpoints {
			if blockhash.IsEqual(checkpoint.Hash) {
				log.Error(fmt.Sprintf("%s is check point", blockhash.String()))
				return true
			}
		}
	}
	latestCheckpoint := chain.LatestCheckpoint()
	if latestCheckpoint == nil {
		latestCheckpoint = &params.Checkpoint{
			Hash:  params.ActiveNetParams.GenesisHash,
			Layer: 0,
		}
	}

	checkpointConfirmations := uint64(blockchain.CheckpointConfirmations)
	requiredLayer := latestCheckpoint.Layer + checkpointConfirmations
	if len(params.ActiveNetParams.Checkpoints) == 0 {
		requiredLayer = 1
	}
	if uint64(block.GetLayer()) < requiredLayer {
		log.Error(fmt.Sprintf("%s is not check point", blockhash.String()))
		return true
	}

	var preblock blockdag.IBlock
	if block.HasChildren() {
		for k, v := range block.GetChildren().GetMap() {
			if chain.BlockDAG().IsOnMainChain(k) {
				preblock = v.(blockdag.IBlock)
				break
			}
		}
	}

	if preblock == nil {
		log.Error(fmt.Sprintf("%s is not check point", blockhash.String()))
		return true
	}
	isCandidate, err := chain.IsCheckpointCandidate(preblock, block)
	if err != nil {
		log.Error(fmt.Sprintf("%s is not check point", blockhash.String()))
		return true
	}

	if isCandidate {
		log.Error(fmt.Sprintf("%s is candidate check point", blockhash.String()))
	}
	return true
}
