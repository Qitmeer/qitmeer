/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qng-core/meerdag"
	"github.com/Qitmeer/qng-core/params"
	"math"
)

const (
	// VBTopBits defines the bits to set in the version to signal that the
	// version bits scheme is being used.
	VBTopBits = 0x20000000

	// VBTopMask is the bitmask to use to determine whether or not the
	// version bits scheme is in use.
	VBTopMask = 0xe0000000

	// VBNumBits is the total number of bits available for use with the
	// version bits scheme.
	VBNumBits = 29

	// time or main height threshold
	CheckerTimeThreshold = 0x60000000 // 2021-01-14 08:25:36 +0000 UTC
)

// bitConditionChecker provides a thresholdConditionChecker which can be used to
// test whether or not a specific bit is set when it's not supposed to be
// according to the expected version based on the known deployments and the
// current state of the chain.  This is useful for detecting and warning about
// unknown rule activations.
type bitConditionChecker struct {
	bit   uint32
	chain *BlockChain
}

// Ensure the bitConditionChecker type implements the thresholdConditionChecker
// interface.
var _ thresholdConditionChecker = bitConditionChecker{}

// BeginTime returns the unix timestamp for the median block time after which
// voting on a rule change starts (at the next window).
//
// Since this implementation checks for unknown rules, it returns 0 so the rule
// is always treated as active.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c bitConditionChecker) BeginTime() uint64 {
	return 0
}

// EndTime returns the unix timestamp for the median block time after which
// voting on a rule change end (at the next window).
//
// Since this implementation checks for unknown rules, it returns the maximum
// possible timestamp so the rule is always treated as active.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c bitConditionChecker) EndTime() uint64 {
	return math.MaxUint64
}

// PerformTime returns the unix timestamp for the median block time after which an
// attempted rule change fails if it has not already been locked in or
// activated.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c bitConditionChecker) PerformTime() uint64 {
	return 0
}

// RuleChangeActivationThreshold is the number of blocks for which the condition
// must be true in order to lock in a rule change.
//
// This implementation returns the value defined by the chain params the checker
// is associated with.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c bitConditionChecker) RuleChangeActivationThreshold() uint32 {
	return c.chain.params.RuleChangeActivationThreshold
}

// MinerConfirmationWindow is the number of blocks in each threshold state
// retarget window.
//
// This implementation returns the value defined by the chain params the checker
// is associated with.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c bitConditionChecker) MinerConfirmationWindow() uint32 {
	return c.chain.params.MinerConfirmationWindow
}

// Condition returns true when the specific bit associated with the checker is
// set and it's not supposed to be according to the expected version based on
// the known deployments and the current state of the chain.
//
// This function MUST be called with the chain state lock held (for writes).
//
// This is part of the thresholdConditionChecker interface implementation.
func (c bitConditionChecker) Condition(node meerdag.IBlock) (bool, error) {
	bn := c.chain.GetBlockNode(node)
	if bn == nil {
		return false, nil
	}
	conditionMask := uint32(1) << c.bit
	version := bn.GetHeader().Version
	if version&VBTopMask != VBTopBits {
		return false, nil
	}
	if version&conditionMask == 0 {
		return false, nil
	}
	expectedVersion, err := c.chain.calcNextBlockVersion(c.chain.bd.GetBlockById(node.GetMainParent()))
	if err != nil {
		return false, err
	}
	return uint32(expectedVersion)&conditionMask == 0, nil
}

// deploymentChecker provides a thresholdConditionChecker which can be used to
// test a specific deployment rule.  This is required for properly detecting
// and activating consensus rule changes.
type deploymentChecker struct {
	deployment *params.ConsensusDeployment
	chain      *BlockChain
}

// Ensure the deploymentChecker type implements the thresholdConditionChecker
// interface.
var _ thresholdConditionChecker = deploymentChecker{}

// BeginTime returns the unix timestamp for the median block time after which
// voting on a rule change starts (at the next window).
//
// This implementation returns the value defined by the specific deployment the
// checker is associated with.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c deploymentChecker) BeginTime() uint64 {
	return c.deployment.StartTime
}

// EndTime returns the unix timestamp for the median block time after which
// voting on a rule change end (at the next window).
//
// This implementation returns the value defined by the specific deployment the
// checker is associated with.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c deploymentChecker) EndTime() uint64 {
	return c.deployment.ExpireTime
}

// PerformTime returns the unix timestamp for the median block time after which an
// attempted rule change fails if it has not already been locked in or
// activated.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c deploymentChecker) PerformTime() uint64 {
	return c.deployment.PerformTime
}

// RuleChangeActivationThreshold is the number of blocks for which the condition
// must be true in order to lock in a rule change.
//
// This implementation returns the value defined by the chain params the checker
// is associated with.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c deploymentChecker) RuleChangeActivationThreshold() uint32 {
	return c.chain.params.RuleChangeActivationThreshold
}

// MinerConfirmationWindow is the number of blocks in each threshold state
// retarget window.
//
// This implementation returns the value defined by the chain params the checker
// is associated with.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c deploymentChecker) MinerConfirmationWindow() uint32 {
	return c.chain.params.MinerConfirmationWindow
}

// Condition returns true when the specific bit defined by the deployment
// associated with the checker is set.
//
// This is part of the thresholdConditionChecker interface implementation.
func (c deploymentChecker) Condition(node meerdag.IBlock) (bool, error) {
	bn := c.chain.GetBlockNode(node)
	if bn == nil {
		return false, nil
	}

	conditionMask := uint32(1) << c.deployment.BitNumber
	version := bn.GetHeader().Version
	return (version&VBTopMask == VBTopBits) && (version&conditionMask != 0),
		nil
}

// calcNextBlockVersion calculates the expected version of the block after the
// passed previous block node based on the state of started and locked in
// rule change deployments.
//
// This function differs from the exported CalcNextBlockVersion in that the
// exported version uses the current best chain as the previous block node
// while this function accepts any block node.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) calcNextBlockVersion(prevNode meerdag.IBlock) (uint32, error) {
	// Set the appropriate bits for each actively defined rule deployment
	// that is either in the process of being voted on, or locked in for the
	// activation at the next threshold window change.
	expectedVersion := uint32(VBTopBits)
	for id := 0; id < len(b.params.Deployments); id++ {
		deployment := &b.params.Deployments[id]
		cache := &b.deploymentCaches[id]
		checker := deploymentChecker{deployment: deployment, chain: b}
		state, err := b.thresholdState(prevNode, checker, cache)
		if err != nil {
			return 0, err
		}
		if state == ThresholdStarted || state == ThresholdLockedIn {
			expectedVersion |= uint32(1) << deployment.BitNumber
		}
	}
	return expectedVersion, nil
}

// CalcNextBlockVersion calculates the expected version of the block after the
// end of the current best chain based on the state of started and locked in
// rule change deployments.
//
// This function is safe for concurrent access.
func (b *BlockChain) CalcNextBlockVersion() (uint32, error) {
	b.ChainLock()
	version, err := b.calcNextBlockVersion(b.bd.GetMainChainTip())
	b.ChainUnlock()
	return version, err
}

// warnUnknownRuleActivations displays a warning when any unknown new rules are
// either about to activate or have been activated.  This will only happen once
// when new rules have been activated and every block for those about to be
// activated.
//
// This function MUST be called with the chain state lock held (for writes)
func (b *BlockChain) warnUnknownRuleActivations(node meerdag.IBlock) error {
	if node == nil {
		return fmt.Errorf("No block:%s\n", node.GetHash())
	}
	mp := b.bd.GetBlockById(node.GetMainParent())
	// Warn if any unknown new rules are either about to activate or have
	// already been activated.
	for bit := uint32(0); bit < VBNumBits; bit++ {
		checker := bitConditionChecker{bit: bit, chain: b}
		cache := &b.warningCaches[bit]

		state, err := b.thresholdState(mp, checker, cache)
		if err != nil {
			return err
		}

		switch state {
		case ThresholdActive:
			if !b.unknownRulesWarned {
				log.Warn(fmt.Sprintf("Unknown new rules activated (bit %d)",
					bit))
				b.unknownRulesWarned = true
			}

		case ThresholdLockedIn:
			window := int32(checker.MinerConfirmationWindow())
			activationHeight := window - (int32(node.GetHeight()) % window)
			log.Warn(fmt.Sprintf("Unknown new rules are about to activate in "+
				"%d blocks (bit %d)", activationHeight, bit))
		}
	}

	return nil
}
