/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qng-core/meerdag"
)

// ThresholdState define the various threshold states used when voting on
// consensus changes.
type ThresholdState byte

// These constants are used to identify specific threshold states.
const (
	// ThresholdDefined is the first state for each deployment and is the
	// state for the genesis block has by definition for all deployments.
	ThresholdDefined ThresholdState = iota

	// ThresholdStarted is the state for a deployment once its start time
	// has been reached.
	ThresholdStarted

	// ThresholdLockedIn is the state for a deployment during the retarget
	// period which is after the ThresholdStarted state period and the
	// number of blocks that have voted for the deployment equal or exceed
	// the required number of votes for the deployment.
	ThresholdLockedIn

	// ThresholdActive is the state for a deployment for all blocks after a
	// retarget period in which the deployment was in the ThresholdLockedIn
	// state.
	ThresholdActive

	// ThresholdFailed is the state for a deployment once its expiration
	// time has been reached and it did not reach the ThresholdLockedIn
	// state.
	ThresholdFailed

	// numThresholdsStates is the maximum number of threshold states used in
	// tests.
	numThresholdsStates
)

// thresholdStateStrings is a map of ThresholdState values back to their
// constant names for pretty printing.
var thresholdStateStrings = map[ThresholdState]string{
	ThresholdDefined:  "ThresholdDefined",
	ThresholdStarted:  "ThresholdStarted",
	ThresholdLockedIn: "ThresholdLockedIn",
	ThresholdActive:   "ThresholdActive",
	ThresholdFailed:   "ThresholdFailed",
}

// String returns the ThresholdState as a human-readable name.
func (t ThresholdState) String() string {
	if s := thresholdStateStrings[t]; s != "" {
		return s
	}
	return fmt.Sprintf("Unknown ThresholdState (%d)", int(t))
}

func (t ThresholdState) HumanString() string {
	switch t {
	case ThresholdDefined:
		return "defined"
	case ThresholdStarted:
		return "started"
	case ThresholdLockedIn:
		return "lockedin"
	case ThresholdActive:
		return "active"
	case ThresholdFailed:
		return "failed"
	default:
		return fmt.Sprintf("unknown deployment state: %v", t)
	}
}

// thresholdConditionChecker provides a generic interface that is invoked to
// determine when a consensus rule change threshold should be changed.
type thresholdConditionChecker interface {
	// BeginTime returns the unix timestamp for the median block time after
	// which voting on a rule change starts (at the next window).
	BeginTime() uint64

	// EndTime returns the unix timestamp for the median block time after
	// which voting on a rule change end (at the next window).
	EndTime() uint64

	// PerformTime returns the unix timestamp for the median block time after
	// which an attempted rule change fails if it has not already been
	// locked in or activated.
	PerformTime() uint64

	// RuleChangeActivationThreshold is the number of blocks for which the
	// condition must be true in order to lock in a rule change.
	RuleChangeActivationThreshold() uint32

	// MinerConfirmationWindow is the number of blocks in each threshold
	// state retarget window.
	MinerConfirmationWindow() uint32

	// Condition returns whether or not the rule change activation condition
	// has been met.  This typically involves checking whether or not the
	// bit associated with the condition is set, but can be more complex as
	// needed.
	Condition(blockdag.IBlock) (bool, error)
}

// thresholdStateCache provides a type to cache the threshold states of each
// threshold window for a set of IDs.
type thresholdStateCache struct {
	entries map[uint]ThresholdState
}

// Lookup returns the threshold state associated with the given hash along with
// a boolean that indicates whether or not it is valid.
func (c *thresholdStateCache) Lookup(id uint) (ThresholdState, bool) {
	state, ok := c.entries[id]
	return state, ok
}

// Update updates the cache to contain the provided hash to threshold state
// mapping.
func (c *thresholdStateCache) Update(id uint, state ThresholdState) {
	c.entries[id] = state
}

// newThresholdCaches returns a new array of caches to be used when calculating
// threshold states.
func newThresholdCaches(numCaches uint32) []thresholdStateCache {
	caches := make([]thresholdStateCache, numCaches)
	for i := 0; i < len(caches); i++ {
		caches[i] = thresholdStateCache{
			entries: make(map[uint]ThresholdState),
		}
	}
	return caches
}

func isCheckerTimeMode(checker thresholdConditionChecker) bool {
	return checker.BeginTime() >= CheckerTimeThreshold
}

// thresholdState returns the current rule change threshold state for the block
// AFTER the given node and deployment ID.  The cache is used to ensure the
// threshold states for previous windows are only calculated once.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) thresholdState(prevNode blockdag.IBlock, checker thresholdConditionChecker, cache *thresholdStateCache) (ThresholdState, error) {
	// The threshold state for the window that contains the genesis block is
	// defined by definition.
	confirmationWindow := int64(checker.MinerConfirmationWindow())
	if prevNode == nil || confirmationWindow <= 0 || int64(prevNode.GetHeight()+1) < confirmationWindow {
		return ThresholdDefined, nil
	}

	// Get the ancestor that is the last block of the previous confirmation
	// window in order to get its threshold state.  This can be done because
	// the state is the same for all blocks within a given window.
	prevNode = b.bd.GetMainAncestor(prevNode, int64(prevNode.GetHeight())-(int64(prevNode.GetHeight()+1))%confirmationWindow)

	// Iterate backwards through each of the previous confirmation windows
	// to find the most recently cached threshold state.
	neededStates := []blockdag.IBlock{}
	for prevNode != nil {
		// Nothing more to do if the state of the block is already
		// cached.
		if _, ok := cache.Lookup(prevNode.GetID()); ok {
			break
		}

		if isCheckerTimeMode(checker) {
			// The start and expiration times are based on the median block
			// time, so calculate it now.
			medianTime := b.CalcPastMedianTime(prevNode)

			// The state is simply defined if the start time hasn't been
			// been reached yet.
			if uint64(medianTime.Unix()) < checker.BeginTime() {
				cache.Update(prevNode.GetID(), ThresholdDefined)
				break
			}
		} else {
			if uint64(prevNode.GetHeight()) < checker.BeginTime() {
				cache.Update(prevNode.GetID(), ThresholdDefined)
				break
			}
		}

		// Add this node to the list of nodes that need the state
		// calculated and cached.
		neededStates = append(neededStates, prevNode)

		// Get the ancestor that is the last block of the previous
		// confirmation window.
		prevNode = b.bd.RelativeMainAncestor(prevNode, confirmationWindow)
	}

	// Start with the threshold state for the most recent confirmation
	// window that has a cached state.
	state := ThresholdDefined
	if prevNode != nil {
		var ok bool
		state, ok = cache.Lookup(prevNode.GetID())
		if !ok {
			return ThresholdFailed, AssertError(fmt.Sprintf(
				"thresholdState: cache lookup failed for %v",
				prevNode.GetHash()))
		}
	}

	// Since each threshold state depends on the state of the previous
	// window, iterate starting from the oldest unknown window.
	for neededNum := len(neededStates) - 1; neededNum >= 0; neededNum-- {
		prevNode := neededStates[neededNum]

		switch state {
		case ThresholdDefined:
			if isCheckerTimeMode(checker) {
				// The deployment of the rule change fails if it expires
				// before it is accepted and locked in.
				medianTime := b.CalcPastMedianTime(prevNode)
				medianTimeUnix := uint64(medianTime.Unix())
				if medianTimeUnix >= checker.EndTime() {
					state = ThresholdFailed
					break
				}

				// The state for the rule moves to the started state
				// once its start time has been reached (and it hasn't
				// already expired per the above).
				if medianTimeUnix >= checker.BeginTime() {
					state = ThresholdStarted
				}
			} else {
				if uint64(prevNode.GetHeight()) >= checker.EndTime() {
					state = ThresholdFailed
					break
				}
				if uint64(prevNode.GetHeight()) >= checker.BeginTime() {
					state = ThresholdStarted
				}
			}

		case ThresholdStarted:
			// The deployment of the rule change fails if it expires
			// before it is accepted and locked in.
			if isCheckerTimeMode(checker) {
				medianTime := b.CalcPastMedianTime(prevNode)
				if uint64(medianTime.Unix()) >= checker.EndTime() {
					state = ThresholdFailed
					break
				}
			} else {
				if uint64(prevNode.GetHeight()) >= checker.EndTime() {
					state = ThresholdFailed
					break
				}
			}

			// At this point, the rule change is still being voted
			// on by the miners, so iterate backwards through the
			// confirmation window to count all of the votes in it.
			var count uint32
			countNode := prevNode
			for i := int64(0); i < confirmationWindow; i++ {
				condition, err := checker.Condition(countNode)
				if err != nil {
					return ThresholdFailed, err
				}
				if condition {
					count++
				}

				// Get the previous block node.
				countNode = b.bd.GetBlockById(countNode.GetMainParent())
				if countNode == nil {
					break
				}
			}

			// The state is locked in if the number of blocks in the
			// period that voted for the rule change meets the
			// activation threshold.
			if count >= checker.RuleChangeActivationThreshold() {
				state = ThresholdLockedIn
			}

		case ThresholdLockedIn:
			if checker.PerformTime() == 0 {
				// The new rule becomes active when its previous state
				// was locked in.
				state = ThresholdActive
			} else {
				if isCheckerTimeMode(checker) {
					medianTime := b.CalcPastMedianTime(prevNode)
					if uint64(medianTime.Unix()) >= checker.PerformTime() {
						state = ThresholdActive
					}
				} else {
					if uint64(prevNode.GetHeight()) >= checker.PerformTime() {
						state = ThresholdActive
					}
				}

			}

		// Nothing to do if the previous state is active or failed since
		// they are both terminal states.
		case ThresholdActive:
		case ThresholdFailed:
		}

		// Update the cache to avoid recalculating the state in the
		// future.
		cache.Update(prevNode.GetID(), state)
	}

	return state, nil
}

// ThresholdState returns the current rule change threshold state of the given
// deployment ID for the block AFTER the end of the current best chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) ThresholdState(deploymentID uint32) (ThresholdState, error) {
	mtip := b.bd.GetMainChainTip()
	b.chainLock.Lock()
	state, err := b.deploymentState(mtip, deploymentID)
	b.chainLock.Unlock()

	return state, err
}

// IsDeploymentActive returns true if the target deploymentID is active, and
// false otherwise.
//
// This function is safe for concurrent access.
func (b *BlockChain) IsDeploymentActive(deploymentID uint32) (bool, error) {
	b.chainLock.Lock()
	state, err := b.isDeploymentActive(deploymentID)
	b.chainLock.Unlock()
	return state, err
}

func (b *BlockChain) isDeploymentActive(deploymentID uint32) (bool, error) {
	state, err := b.deploymentState(b.bd.GetMainChainTip(), deploymentID)
	if err != nil {
		return false, err
	}
	return state == ThresholdActive, nil
}

// deploymentState returns the current rule change threshold for a given
// deploymentID. The threshold is evaluated from the point of view of the block
// node passed in as the first argument to this method.
//
// It is important to note that, as the variable name indicates, this function
// expects the block node prior to the block for which the deployment state is
// desired.  In other words, the returned deployment state is for the block
// AFTER the passed node.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) deploymentState(prevNode blockdag.IBlock, deploymentID uint32) (ThresholdState, error) {
	if deploymentID > uint32(len(b.params.Deployments)) {
		return ThresholdFailed, fmt.Errorf(DeploymentError(deploymentID).Error())
	}

	deployment := &b.params.Deployments[deploymentID]
	checker := deploymentChecker{deployment: deployment, chain: b}
	cache := &b.deploymentCaches[deploymentID]

	return b.thresholdState(prevNode, checker, cache)
}

// initThresholdCaches initializes the threshold state caches for each warning
// bit and defined deployment and provides warnings if the chain is current per
// the warnUnknownRuleActivations function.
func (b *BlockChain) initThresholdCaches() error {
	// Initialize the warning and deployment caches by calculating the
	// threshold state for each of them.  This will ensure the caches are
	// populated and any states that needed to be recalculated due to
	// definition changes is done now.
	prevNode := b.bd.GetBlockById(b.bd.GetMainChainTip().GetMainParent())

	for bit := uint32(0); bit < VBNumBits; bit++ {
		checker := bitConditionChecker{bit: bit, chain: b}
		cache := &b.warningCaches[bit]
		_, err := b.thresholdState(prevNode, checker, cache)
		if err != nil {
			return err
		}
	}
	for id := 0; id < len(b.params.Deployments); id++ {
		deployment := &b.params.Deployments[id]
		cache := &b.deploymentCaches[id]
		checker := deploymentChecker{deployment: deployment, chain: b}
		_, err := b.thresholdState(prevNode, checker, cache)
		if err != nil {
			return err
		}
	}

	// No warnings about unknown rules until the chain is current.
	if b.isCurrent() {
		// Warn if any unknown new rules are either about to activate or
		// have already been activated.
		if err := b.warnUnknownRuleActivations(b.bd.GetMainChainTip()); err != nil {
			return err
		}
	}

	return nil
}
