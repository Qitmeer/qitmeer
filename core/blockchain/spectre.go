package blockchain

import (
	"fmt"
	"github.com/golang-collections/collections/stack"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/common/util"
	"strconv"
)

type ISpectre interface {
	Vote(x IBlock, y IBlock) int
}

type Spectre struct {
	dag IBlockDAG

	// voting votes, true value means voting the first candidate, otherwise voting the second
	votes map[hash.Hash]bool

	// the nodes don't exist in future sets of 1 and 2, but they are referenced by the future sets
	//  e.g. node 7 in ByteBall 2
	dangling *BlockSet

	// the candidates to compete with each other
	candidate1, candidate2 IBlock
}

func NewSpectre(dag IBlockDAG) *Spectre {

	s := new(Spectre)
	s.dag=dag
	s.votes = make(map[hash.Hash]bool)
	s.dangling = NewBlockSet()
	return s
}

func (s *Spectre) voteFirst(voter hash.Hash) {
	s.votes[voter] = true
}

func (s *Spectre) voteSecond(voter hash.Hash) {
	s.votes[voter] = false
}

func (s *Spectre) hasVoted(voter hash.Hash) bool {
	if _, ok := s.votes[voter]; ok {
		return true
	}

	return false
}

func (s *Spectre) Votes() map[hash.Hash]bool {
	return s.votes
}

func (s *Spectre) InitVote(b1 IBlock, b2 IBlock) (bool, error) {
	s.candidate1, s.candidate2 = b1, b2
	tiebreak := s.candidate1.GetHash().String() < s.candidate2.GetHash().String()

	exist1 := s.dag.HasBlock(b1.GetHash())
	exist2 := s.dag.HasBlock(b2.GetHash())
	if !exist1 && exist2 {
		return false, fmt.Errorf("block  %v doesn't exist", b1.GetHash())
	} else if exist1 && !exist2 {
		return true, fmt.Errorf("block  %v doesn't exist", b2.GetHash())
	} else if !exist1 && !exist2 {
		return tiebreak, fmt.Errorf("block  %v and %v don't exist", b1.GetHash(), b2.GetHash())
	}

	if b1 == b2 {
		return tiebreak, fmt.Errorf("block %v is identical to block %v", b1.GetHash(), b2.GetHash())
	} else if s.IsInPastOf(b1, b2) {
		return true, fmt.Errorf("block %v is in past of block %v", b1.GetHash(), b2.GetHash())
	} else if s.IsInPastOf(b2, b1) {
		return false, fmt.Errorf("block %v is in past of block %v", b2.GetHash(), b1.GetHash())
	}
	s.voteBySelf(b1, b2)
	s.voteByUniqueFutureSet(b1, b2)

	return tiebreak, nil
}

func (s *Spectre) Vote(b1 IBlock, b2 IBlock) (bool, error) {
	if v, err := s.InitVote(b1, b2); err != nil {
		return v, err
	}

	// For any node n in voted notes, if any child c of n hasn't voted, then add c into outer
	// Outer is the outside layer of the voted notes, which means not all of its children have voted,
	// so we start voting with these pending voting children. Once they have all voted, their children will become
	// the new outside layer, so on and so forth, until all nodes have voted.
	// note: tips are also outside layer, even though they have no children

	outer := util.NewIterativeQueue()
	for h := range s.votes {
		children := s.dag.GetBlock(&h).GetChildren()
		if children == nil { // tips
			outer.Enqueue(h)
		} else { // haven't voted children
			for ch := range children.GetMap() {
				if !s.hasVoted(ch) {
					outer.Enqueue(h)
					break
				}
			}
		}
	}

	for outer.Len() > 0 {
		any := outer.Dequeue().(hash.Hash)

		// having all children of some outer node voted or not, if this is true, that node can be dequeued
		done := true
		children := s.dag.GetBlock(&any).GetChildren()
		if children == nil || children.Len() == 0 { // tips
			// game over once all tips have voted
			if win, err := s.VoteByBlock(nil); err == nil {
				return win, nil
			}
			// go on when some tip hasn't voted
			continue
		}

		for ch := range children.GetMap() {
			if s.hasVoted(ch) {
				continue
			}
			all := true // all parented voted
			chParents:=s.dag.GetBlock(&ch).GetParents()
			for ph := range chParents.GetMap() {
				// note: must ignore dangling parent,
				// e.g. in figure ByteBall2, 7 is a dangling node, so once 17 and 21 are voted, 24 is able to vote
				if !s.hasVoted(ph) && !s.dangling.Has(&ph) {
					all = false
					break
				}
			}
			if all {
				vb := s.dag.GetBlock(&ch)
				s.VoteByBlock(vb)
				outer.Enqueue(ch)
			} else {
				done = false
			}
		}

		// needs processing again
		if !done {
			outer.Enqueue(any)
		}

		// if all the outer nodes vote unanimously, then no need to come up with the rest nodes,
		// because they will inherit the result of outer nodes
		// e.g. in figure ByteBall2, once 17, 21, 20 have voted the same candidate,
		// then 19, 22~27 should follow their votes
		consistent := true
		first := outer.Peek().(hash.Hash)
		vote := s.votes[first]
		for _, o := range outer {
			if s.votes[o.(hash.Hash)] != vote {
				consistent = false
				break
			}
		}
		if consistent {
			return vote, nil
		}
	}

	return s.candidate1.GetHash().String() < s.candidate2.GetHash().String(), nil
}

//  TODO: test if there is ancestor-descendant relationship between b1 and b2
func (s *Spectre) IsInPastOf(b1 IBlock, b2 IBlock) bool {
	return false
}

// intersection of virtual block (if not nil) with its past set and voted nodes,
// note the DIRECTION IS REVERSED than the original graph, which means virtual block is genesis
func (s *Spectre) votedPast(virtualBlock IBlock) IBlockDAG {
	q := util.NewIterativeQueue()

	if virtualBlock == nil {
		for ht := range s.dag.GetTips().GetMap() {
			q.Enqueue(ht)
		}
	} else {
		q.Enqueue(*virtualBlock.GetHash())
	}

	visited := NewBlockSet()
	// cache the intersection of past set and voted notes
	cache := make(map[hash.Hash]IBlock)
	for q.Len() > 0 {
		h := q.Dequeue().(hash.Hash)

		if visited.Has(&h) {
			continue
		} else {
			visited.Add(&h)
		}
		hParents:=s.dag.GetBlock(&h).GetParents()
		for ph := range hParents.GetMap() {
			if s.dangling.Has(&ph) {
				continue
			}
			if s.hasVoted(ph) {
				if !visited.Has(&ph) {
					q.Enqueue(ph)
				}

				// must cache block  due to children index
				sb, ok := cache[ph]
				if !ok {
					sb = NewSpectreBlock(&ph)
					cache[ph] = sb
				}
				sb.GetParents().Add(&h)
			}
		}
	}

	vh := hash.MustHexToHash(strconv.Itoa(int(s.dag.GetBlockCount())))
	if virtualBlock != nil {
		vh = *virtualBlock.GetHash()
	}
	sb := NewSpectreBlock(&vh)
	vp := &BlockSpectre{}
	vp.AddBlock(sb)
	visited = NewBlockSet()

	q = util.NewIterativeQueue()
	if virtualBlock == nil {
		for th := range s.dag.GetTips().GetMap() {
			sb := NewSpectreBlock(&th)
			// create a virtual block as genesis
			sb.GetParents().Add(&vh)
			vp.AddBlock(sb)
			q.Enqueue(th)
		}
	} else {
		q.Enqueue(vh)
	}
	for q.Len() > 0 {
		pos := q.Dequeue().(hash.Hash)
		visited.Add(&pos)
		posParents:=s.dag.GetBlock(&pos).GetParents()
		for ph := range posParents.GetMap() {
			if !s.hasVoted(ph) || s.dangling.Has(&ph) {
				continue
			}
			if visited.Has(&ph) {
				continue
			}

			all := true
			phChildren:=s.dag.GetBlock(&ph).GetChildren()
			for ch := range phChildren.GetMap() {
				if _, ok := cache[ch]; !ok {
					continue
				}
				if !visited.Has(&ch) {
					all = false
				}
			}
			// ensure that add ancestors before descendants
			if all {
				q.Enqueue(ph)
				s.newVoter(ph, vp)
				visited.Add(&ph)
			}
		}
	}

	return vp
}

// Update votes of any node in votedPast
// Note: the ancestors of the node must have their votes updated since the node needs inherit those votes
func (s *Spectre) updateVotes(votedPast IBlockDAG, vh hash.Hash) bool {
	// test if all parents being voted
	canUpdate := true
	maxVotes := -1
	maxParent := new(SpectreBlock)

	// increase votedPast with new nodes, only happening on updating votes in candidates' past sets
	if !votedPast.HasBlock(&vh) {
		vhChildren:=s.dag.GetBlock(&vh).GetChildren()
		for ch := range vhChildren.GetMap() {
			if !votedPast.HasBlock(&ch) && !s.hasVoted(ch) {
				canUpdate = false
				break
			}
		}
		if canUpdate {
			s.newVoter(vh, votedPast)
		} else {
			return false
		}
	}

	parents := votedPast.GetBlock(&vh).GetParents()

	if parents == nil || parents.Len() == 0 {
		log.Error("no parents of ", vh)
	}

	// max parent has more nodes in its future set, which means more votes to inherit
	for ph := range parents.GetMap() {
		if ph.IsEqual(votedPast.GetGenesis().GetHash()) {
			continue
		}

		sb := votedPast.GetBlock(&ph).(*SpectreBlock)
		if sb.Votes1 < 0 || sb.Votes2 < 0 {
			canUpdate = false
			break
		}
		votes := sb.Votes2 + sb.Votes1
		if votes > maxVotes {
			maxVotes = votes
			maxParent = sb
		}
	}

	if canUpdate {
		// first step, inherit votes from max voter
		voter := votedPast.GetBlock(&vh).(*SpectreBlock)
		voter.Votes1, voter.Votes2 = maxParent.Votes1, maxParent.Votes2

		// if it can be updated, it MUST be updated
		if maxParent == nil || maxParent.hash.IsEqual(&hash.Hash{}) {
			log.Error(vh.String())
		}

		// Note: results in s is constant, so we must reference them in the first place
		// then we compare its votes between candidate 1 and 2
		if s.hasVoted(*maxParent.GetHash()) {
			v := s.votes[*maxParent.GetHash()]
			if v {
				voter.Votes1 += 1
			} else {
				voter.Votes2 += 1
			}
		} else {
			if maxParent.Votes2 > maxParent.Votes1 {
				voter.Votes2 += 1
			} else if maxParent.Votes2 < maxParent.Votes1 {
				voter.Votes1 += 1
			}
		}

		// second, add votes from other voters
		if voter.GetParents().Len() > 1 {
			s.updateTipVotes(voter, maxParent, votedPast)
		}
	}
	return canUpdate
}

// once we have inherited votes from max parent (the tip with max votes), we need add votes from other tips
// because each tip and its future set have various nodes in common with another tip and the common part has been
// calculated by inheriting the max parent, we only need to add votes of nodes which are exclusive in each tip

// e.g. in ByteBall2, if virtual block is 21 and the intersection of its past set and candidates' future sets has voted,
// which are 11~18 and 20 in this case. For 10,
// 12 is the max parent (3 votes for candidate 1(c1), 4 votes for candidate 2(c2)), so we inherit its votes.
// for the other tips, in this case it is only 13, and we find the nodes which are inside 13 and 13's future set but
// outside 12 and 12's future set is 13, and 13 has voted c2. so the final votes of 10 would be 3 for c1 and 4 for c2

func (s *Spectre) updateTipVotes(voter *SpectreBlock, maxParent *SpectreBlock, votedPast IBlockDAG) {
	voterParents := voter.GetParents()
	tipStack := stack.New()
	tipSet := NewBlockSet()
	// take out all other tips and add their votes to child
	for h := range voterParents.GetMap() {
		if !h.IsEqual(maxParent.GetHash()) && !h.IsEqual(votedPast.GetGenesis().GetHash()) {
			tipStack.Push(h)
			tipSet.Add(&h)
		}
	}
	for tipStack.Len() > 0 {
		tipHash := tipStack.Pop().(hash.Hash)
		tipVoter := votedPast.GetBlock(&tipHash).(*SpectreBlock)
		if s.hasVoted(tipHash) {
			v := s.votes[tipHash]
			if v {
				voter.Votes1 += 1
			} else {
				voter.Votes2 += 1
			}
		} else {
			if tipVoter.Votes2 > tipVoter.Votes1 {
				voter.Votes2 += 1
			} else if tipVoter.Votes2 < tipVoter.Votes1 {
				voter.Votes1 += 1
			}
		}

		// find nodes exclusively exist in one tip. We save all the nodes visited in tipSet, for a note's parent p,
		// if all children of p exist in tipSet, p is exclusively in that node's future set.
		// e.g. in ByteBall2 with 21 as the virtual block, from 10's view, if we want to find 12's exclusive future,
		// we save 12 into tipSet first, the 14 and 15 are 12's exclusive parents since all their children
		// (just 12 in this case ) exist in tipSet
		for tp := range tipVoter.GetParents().GetMap() {
			if tipSet.Has(&tp) {
				continue
			}
			only := true
			tpChildren:=votedPast.GetBlock(&tp).GetChildren()
			for tc := range tpChildren.GetMap() {
				if !tipSet.Has(&tc) {
					only = false
					break
				}
			}
			if only {
				tipStack.Push(tp)
				tipSet.Add(&tp)
			}
		}
	}
}
func (s *Spectre) followParents(virtualBlock IBlock) (bool, bool, error) {
	// if all parents have voted in consistence, then just follow them
	consistent := true
	last := 0

	parents := NewBlockSet()
	if virtualBlock == nil { // whole graph
		parents = s.dag.GetTips()
	} else { // past set of one node
		parents = virtualBlock.GetParents()
	}
	for ph := range parents.GetMap() {
		if s.dangling.Has(&ph) {
			continue
		}
		if !s.hasVoted(ph) {
			return true, s.candidate1.GetHash().String() < s.candidate2.GetHash().String(), fmt.Errorf("parent %v not ready", ph)
		}
		vote := -1
		if !s.votes[ph] {
			vote = 1
		}
		if last == 0 {
			last = vote
		} else {
			if vote != last {
				consistent = false
				break
			}
		}
	}
	firstWin := false
	if consistent {
		firstWin = last < 0
		h := virtualBlock.GetHash()
		if firstWin {
			s.voteFirst(*h)
		} else {
			s.voteSecond(*h)
		}
		return true, firstWin, nil
	}

	return false, false, nil
}

//2) if z ∈ G is in future (x)∩future (y) then z ’s vote will be determined recursively according to the DAG that is reduced to its past,
// i.e., it has the same vote as virtual (past (z)). If the result of this vote is a tie, z breaks it arbitrarily. 3

// votedPast is a REVERSED graph, which means it starts from a virtual block and traverses backwards

// If virtual block is nil, it is like an imaginary recent coming node which references all the tips and the whole graph
// is its past set,  otherwise it means that some real block is the virtual block of its own past set
func (s *Spectre) VoteByBlock(virtualBlock IBlock) (bool, error) {
	if ok, res, err := s.followParents(virtualBlock); ok {
		return res, err
	}

	votedPast := s.votedPast(virtualBlock)
	// outer of votedPast, note the direction
	outer := s.voteFromFutureSet(votedPast)
	firstWin := s.voteFromPastSet(votedPast, outer)

	//4) if z is the virtual block of G then it will vote the same way as the vote of the majority of blocks in G.
	if virtualBlock != nil {
		h := virtualBlock.GetHash()
		if firstWin {
			s.voteFirst(*h)
		} else {
			s.voteSecond(*h)
		}
	}
	return firstWin, nil
}

//3) if z ∈ G is not in the future of either blocks then it will vote the same way as the vote of the majority of blocks in its own future.

// Update vote numbers of votedPast, e.g. for any node n in votedPast, update its Votes1 and Votes2
// note: because one node's vote number is dependent on its ancestor,
// we need update them from genesis of votePast down to its tips
func (s *Spectre) voteFromFutureSet(votedPast IBlockDAG) *BlockSet {
	tips := NewBlockSet()

	vb := votedPast.GetGenesis()
	unvisited := util.NewIterativeQueue()
	visited := NewBlockSet()

	vChildren := votedPast.GetBlock(vb.GetHash()).GetChildren()
	for ch := range vChildren.GetMap() {
		// only children with one parent (virtual block) will be selected as initial tips,
		// because they are only dependent on their single parent and their votes can be updated directly
		// e.g. note 22 in ByteBall2, 14, 20 can be initialized with 0 votes, but 15 cannot due to its multiple parents
		if votedPast.GetBlock(&ch).GetParents().Len() == 1 {
			sb := votedPast.GetBlock(&ch).(*SpectreBlock)
			sb.Votes1, sb.Votes2 = 0, 0

			parents := s.dag.GetBlock(&ch).GetParents()
			if parents != nil {
				for ph := range parents.GetMap() {
					if !votedPast.HasBlock(&ph) {
						tips.Add(&ch)
						break
					}
				}
			} else {
				if !ch.IsEqual(votedPast.GetGenesis().GetHash()) {
					log.Error("only virtual block can do without parents")
				}
			}
			cChildren := votedPast.GetBlock(&ch).GetChildren()
			if cChildren != nil && cChildren.Len() > 0 {
				unvisited.Enqueue(ch)
			}
		}
	}

	// update votes with BFS order
	for unvisited.Len() > 0 {
		n := unvisited.Dequeue().(hash.Hash)
		childrenUpdated := true
		nChildren:=votedPast.GetBlock(&n).GetChildren()
		for ch := range nChildren.GetMap() {
			if !visited.Has(&ch) {
				if s.updateVotes(votedPast, ch) {
					visited.Add(&ch)

					parents := s.dag.GetBlock(&ch).GetParents()
					if parents != nil {
						for ph := range parents.GetMap() {
							if !votedPast.HasBlock(&ph) {
								tips.Add(&ch)
								break
							}
						}
					} else {
						if !ch.IsEqual(votedPast.GetGenesis().GetHash()) {
							log.Error("only virtual block can do without parents")
						}
					}
					children := votedPast.GetBlock(&ch).GetChildren()
					if children != nil && children.Len() > 0 {
						unvisited.Enqueue(ch)
					}
				} else {
					childrenUpdated = false
				}
			}
		}
		if !childrenUpdated {
			unvisited.Enqueue(n)
		}
	}

	return tips
}

// update votes of candidates's past sets
// tips stand for nodes in votedPast with their children having not fully voted
// and those children are called outer nodes.
// e.g. BB2, VB is 21, 11, 12, 13 are tips, 9 and 10 are outer nodes

// with outer nodes keep voting and become new tips, their children will become the new outer nodes
// repeat this process until all the outer nodes votes for the same candidate, and VB will vote for him as well

func (s *Spectre) voteFromPastSet(votedPast IBlockDAG, tips *BlockSet) bool {
	unvisited := util.NewIterativeQueue()
	outerNodes := NewBlockSet()
	for h := range tips.GetMap() {
		hParents:=s.dag.GetBlock(&h).GetParents()
		for ph := range hParents.GetMap() {
			if !outerNodes.Has(&ph) && !votedPast.HasBlock(&ph) {
				unvisited.Enqueue(ph)
				outerNodes.Add(&ph)
			}
		}
	}

	for unvisited.Len() > 0 {
		n := unvisited.Dequeue().(hash.Hash)

		if !s.updateVotes(votedPast, n) {
			unvisited.Enqueue(n)
		} else {
			lastVote := 0
			consistent := true
			for o := range outerNodes.GetMap() {
				if !votedPast.HasBlock(&o) {
					consistent = false
					break
				}
				sp := votedPast.GetBlock(&o).(*SpectreBlock)
				// we don't consider those with tie votes
				if sp.Votes2 != sp.Votes1 {
					spVote := -1
					if sp.Votes1 < sp.Votes2 {
						spVote = 1
					}
					if lastVote == 0 {
						lastVote = spVote
					} else {
						if lastVote != spVote {
							consistent = false
							break
						}
					}
				}
			}

			// consistent then follow the result
			if consistent {
				return lastVote < 0
			} else {
				removing := NewBlockSet()
				skip := NewBlockSet()
				for o := range outerNodes.GetMap() {
					if skip.Has(&o) {
						continue
					}
					allUpdated := votedPast.HasBlock(&o)
					oParents:=s.dag.GetBlock(&o).GetParents()
					for ph := range oParents.GetMap() {
						if !s.updateVotes(votedPast, ph) {
							allUpdated = false
							break
						} else if outerNodes.Has(&ph) {
							skip.Add(&ph)
						}
					}

					if allUpdated {
						removing.Add(&o)
					}
				}

				for r := range removing.GetMap() {
					outerNodes.Remove(&r)
					rChildren:=votedPast.GetBlock(&r).GetChildren()
					for c := range rChildren.GetMap() {
						if !outerNodes.Has(&c) {
							outerNodes.Add(&c)
							unvisited.Enqueue(c)
						}
					}
				}
			}
		}
	}

	// break the tie
	return s.candidate1.GetHash().String() < s.candidate2.GetHash().String()
}

// add voter into voted past set
func (s *Spectre) newVoter(vh hash.Hash, votedPast IBlockDAG) IBlock {
	sb := NewSpectreBlock(&vh)
	vhChildren:=s.dag.GetBlock(&vh).GetChildren()
	for h := range vhChildren.GetMap() {
		if votedPast.HasBlock(&h) {
			sb.GetParents().Add(&h)
		}
	}
	if votedPast.HasBlock(&vh) {
		log.Error("has already voter ", vh)
	}
	votedPast.(*BlockSpectre).AddBlock(sb)
	return sb
}

//5) finally, (for the case where z equals x or y ), z votes for itself to succeed any block in past (z) and to precede any block outside past (z).
func (s *Spectre) voteBySelf(b1 IBlock, b2 IBlock) {
	s.voteFirst(*b1.GetHash())
	s.voteSecond(*b2.GetHash())
}

// 1) if z ∈ G is in future (x) but not in future (y) then it will vote in favour of x (i.e., for x ≺y ).
func (s *Spectre) voteByUniqueFutureSet(b1 IBlock, b2 IBlock) {
	fs1 := NewBlockSet()
	s.dag.GetFutureSet(fs1, b1)

	fs2 := NewBlockSet()
	s.dag.GetFutureSet(fs2, b2)

	for hf := range fs1.GetMap() {
		hfParents:=s.dag.GetBlock(&hf).GetParents()
		for h := range hfParents.GetMap() {
			if !fs1.Has(&h) && !fs2.Has(&h) {
				s.dangling.Add(&h)
			}
		}
		if fs2.Has(&hf) {
			fs2.Remove(&hf)
		} else {
			s.voteFirst(hf)
		}
	}

	for hf := range fs2.GetMap() {
		hfParents:=s.dag.GetBlock(&hf).GetParents()
		for h := range hfParents.GetMap() {
			if !fs1.Has(&h) && !fs2.Has(&h) {
				s.dangling.Add(&h)
			}
		}
		s.voteSecond(hf)
	}
	s.dangling.Remove(b1.GetHash())
	s.dangling.Remove(b2.GetHash())
	for h := range b1.GetParents().GetMap() {
		s.dangling.Remove(&h)
	}
	for h := range b2.GetParents().GetMap() {
		s.dangling.Remove(&h)
	}
}
