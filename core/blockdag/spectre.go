package blockdag

import (
	"container/list"
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/common/util"
	"github.com/Qitmeer/qng-core/database"
	"github.com/golang-collections/collections/stack"
	"io"
	"strconv"
)

type Spectre struct {
	// The general foundation framework of DAG
	bd *BlockDAG

	// voting votes, true value means voting the first candidate, otherwise voting the second
	votes map[hash.Hash]bool

	// the nodes don't exist in future sets of 1 and 2, but they are referenced by the future sets
	//  e.g. node 7 in ByteBall 2
	dangling *HashSet

	// the candidates to compete with each other
	candidate1, candidate2 IBlock

	// The votes of block
	sblocks map[hash.Hash]*SpectreBlock
}

func (sp *Spectre) GetName() string {
	return spectre
}

func (sp *Spectre) Init(bd *BlockDAG) bool {
	sp.bd = bd

	sp.votes = make(map[hash.Hash]bool)
	sp.dangling = NewHashSet()

	return true
}

func (sp *Spectre) AddBlock(b IBlock) (*list.List, *list.List) {
	if sp.sblocks == nil {
		sp.sblocks = map[hash.Hash]*SpectreBlock{}
	}
	block := SpectreBlock{hash: *b.GetHash(), Votes1: -1, Votes2: -1}
	sp.sblocks[block.hash] = &block

	var result *list.List = list.New()
	result.PushBack(block.GetHash())
	return result, nil
}

// Build self block
func (sp *Spectre) CreateBlock(b *Block) IBlock {
	return b
}

func (sp *Spectre) GetTipsList() []IBlock {
	return nil
}

func (sp *Spectre) voteFirst(voter hash.Hash) {
	sp.votes[voter] = true
}

func (sp *Spectre) voteSecond(voter hash.Hash) {
	sp.votes[voter] = false
}

func (sp *Spectre) hasVoted(voter hash.Hash) bool {
	if _, ok := sp.votes[voter]; ok {
		return true
	}

	return false
}

func (sp *Spectre) Votes() map[hash.Hash]bool {
	return sp.votes
}

func (sp *Spectre) InitVote(b1 IBlock, b2 IBlock) (bool, error) {
	sp.candidate1, sp.candidate2 = b1, b2
	tiebreak := sp.candidate1.GetHash().String() < sp.candidate2.GetHash().String()

	exist1 := sp.bd.hasBlockById(b1.GetID())
	exist2 := sp.bd.hasBlockById(b2.GetID())
	if !exist1 && exist2 {
		return false, fmt.Errorf("block  %v doesn't exist", b1.GetHash())
	} else if exist1 && !exist2 {
		return true, fmt.Errorf("block  %v doesn't exist", b2.GetHash())
	} else if !exist1 && !exist2 {
		return tiebreak, fmt.Errorf("block  %v and %v don't exist", b1.GetHash(), b2.GetHash())
	}

	if b1 == b2 {
		return tiebreak, fmt.Errorf("block %v is identical to block %v", b1.GetHash(), b2.GetHash())
	} else if sp.IsInPastOf(b1, b2) {
		return true, fmt.Errorf("block %v is in past of block %v", b1.GetHash(), b2.GetHash())
	} else if sp.IsInPastOf(b2, b1) {
		return false, fmt.Errorf("block %v is in past of block %v", b2.GetHash(), b1.GetHash())
	}
	sp.voteBySelf(b1, b2)
	sp.voteByUniqueFutureSet(b1, b2)

	return tiebreak, nil
}

func (sp *Spectre) Vote(b1 IBlock, b2 IBlock) (bool, error) {

	if v, err := sp.InitVote(b1, b2); err != nil {
		return v, err
	}

	// For any node n in voted notes, if any child c of n hasn't voted, then add c into outer
	// Outer is the outside layer of the voted notes, which means not all of its children have voted,
	// so we start voting with these pending voting children. Once they have all voted, their children will become
	// the new outside layer, so on and so forth, until all nodes have voted.
	// note: tips are also outside layer, even though they have no children

	outer := util.NewIterativeQueue()
	for h := range sp.votes {
		children := sp.bd.getBlock(&h).GetChildren()
		if children == nil { // tips
			outer.Enqueue(h)
		} else { // haven't voted children
			for _, ch := range children.GetMap() {
				if !sp.hasVoted(*ch.(IBlock).GetHash()) {
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
		children := sp.bd.getBlock(&any).GetChildren()
		if children == nil || children.Size() == 0 { // tips
			// game over once all tips have voted
			if win, err := sp.VoteByBlock(nil); err == nil {
				return win, nil
			}
			// go on when some tip hasn't voted
			continue
		}

		for _, ch := range children.GetMap() {
			if sp.hasVoted(*ch.(IBlock).GetHash()) {
				continue
			}
			all := true // all parented voted
			chParents := sp.bd.getBlock(ch.(IBlock).GetHash()).GetParents()
			for _, ph := range chParents.GetMap() {
				// note: must ignore dangling parent,
				// e.g. in figure ByteBall2, 7 is a dangling node, so once 17 and 21 are voted, 24 is able to vote
				if !sp.hasVoted(*ph.(IBlock).GetHash()) && !sp.dangling.Has(ph.(IBlock).GetHash()) {
					all = false
					break
				}
			}
			if all {
				vb := sp.bd.getBlock(ch.(IBlock).GetHash())
				sp.VoteByBlock(vb)
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
		vote := sp.votes[first]
		for _, o := range outer {
			if sp.votes[o.(hash.Hash)] != vote {
				consistent = false
				break
			}
		}
		if consistent {
			return vote, nil
		}
	}

	return sp.candidate1.GetHash().String() < sp.candidate2.GetHash().String(), nil
}

//  TODO: test if there is ancestor-descendant relationship between b1 and b2
func (sp *Spectre) IsInPastOf(b1 IBlock, b2 IBlock) bool {
	return false
}

// intersection of virtual block (if not nil) with its past set and voted nodes,
// note the DIRECTION IS REVERSED than the original graph, which means virtual block is genesis
func (sp *Spectre) votedPast(virtualBlock IBlock) *BlockDAG {
	q := util.NewIterativeQueue()

	if virtualBlock == nil {
		for ht := range sp.bd.tips.GetMap() {
			q.Enqueue(ht)
		}
	} else {
		q.Enqueue(*virtualBlock.GetHash())
	}

	visited := NewHashSet()
	// cache the intersection of past set and voted notes
	cache := make(map[hash.Hash]IBlock)
	for q.Len() > 0 {
		h := q.Dequeue().(hash.Hash)

		if visited.Has(&h) {
			continue
		} else {
			visited.Add(&h)
		}
		hParents := sp.bd.getBlock(&h).GetParents()
		for _, ib := range hParents.GetMap() {
			ph := *ib.(IBlock).GetHash()
			if sp.dangling.Has(&ph) {
				continue
			}
			if sp.hasVoted(ph) {
				if !visited.Has(&ph) {
					q.Enqueue(ph)
				}

				// must cache block  due to children index
				sb, ok := cache[ph]
				if !ok {
					sb = &Block{hash: ph, parents: NewIdSet(), id: ib.(IBlock).GetID()}
					cache[ph] = sb
				}
				sb.GetParents().AddPair(sp.bd.getBlock(&h).GetID(), sp.bd.getBlock(&h))
			}
		}
	}

	vh := hash.MustHexToDecodedHash(strconv.Itoa(int(sp.bd.blockTotal)))
	if virtualBlock != nil {
		vh = *virtualBlock.GetHash()
	}
	sb := &SpectreBlockData{hash: vh}
	vp := &BlockDAG{}
	vp.Init(spectre, nil, -1, nil, nil)
	vp.AddBlock(sb)
	visited = NewHashSet()

	q = util.NewIterativeQueue()
	if virtualBlock == nil {
		for _, th := range sp.bd.tips.GetMap() {
			sb := &SpectreBlockData{hash: *th.(IBlock).GetHash()}
			sb.parents = []*hash.Hash{}
			// create a virtual block as genesis
			sb.parents = append(sb.parents, &vh)
			vp.AddBlock(sb)
			q.Enqueue(th)
		}
	} else {
		q.Enqueue(vh)
	}
	for q.Len() > 0 {
		pos := q.Dequeue().(hash.Hash)
		visited.Add(&pos)
		posParents := sp.bd.getBlock(&pos).GetParents()
		for _, ib := range posParents.GetMap() {
			ph := *ib.(IBlock).GetHash()
			if !sp.hasVoted(ph) || sp.dangling.Has(&ph) {
				continue
			}
			if visited.Has(&ph) {
				continue
			}

			all := true
			phChildren := sp.bd.getBlock(&ph).GetChildren()
			for _, ch := range phChildren.GetMap() {
				if _, ok := cache[*ch.(IBlock).GetHash()]; !ok {
					continue
				}
				if !visited.Has(ch.(IBlock).GetHash()) {
					all = false
				}
			}
			// ensure that add ancestors before descendants
			if all {
				q.Enqueue(ph)
				sp.newVoter(ph, vp)
				visited.Add(&ph)
			}
		}
	}

	return vp
}

// Update votes of any node in votedPast
// Note: the ancestors of the node must have their votes updated since the node needs inherit those votes
func (sp *Spectre) updateVotes(votedPast *BlockDAG, vh hash.Hash) bool {
	// test if all parents being voted
	canUpdate := true
	maxVotes := -1
	maxParent := new(SpectreBlock)

	// increase votedPast with new nodes, only happening on updating votes in candidates' past sets
	if !votedPast.hasBlockById(votedPast.getBlock(&vh).GetID()) {
		vhChildren := sp.bd.getBlock(&vh).GetChildren()
		for id, ch := range vhChildren.GetMap() {
			if !votedPast.hasBlockById(id) && !sp.hasVoted(*ch.(IBlock).GetHash()) {
				canUpdate = false
				break
			}
		}
		if canUpdate {
			sp.newVoter(vh, votedPast)
		} else {
			return false
		}
	}
	parents := votedPast.getBlock(&vh).GetParents()

	if parents == nil || parents.Size() == 0 {
		log.Error("no parents of ", vh)
	}

	// max parent has more nodes in its future set, which means more votes to inherit
	for id, ph := range parents.GetMap() {
		if ph.(IBlock).GetHash().IsEqual(votedPast.getGenesis().GetHash()) {
			continue
		}
		b := votedPast.getBlockById(id)
		sb := votedPast.instance.(*Spectre).sblocks[*b.GetHash()]
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
		b := votedPast.getBlock(&vh)
		voter := votedPast.instance.(*Spectre).sblocks[*b.GetHash()]

		voter.Votes1, voter.Votes2 = maxParent.Votes1, maxParent.Votes2

		// if it can be updated, it MUST be updated
		if maxParent == nil || maxParent.hash.IsEqual(&hash.Hash{}) {
			log.Error(vh.String())
		}

		// Note: results in s is constant, so we must reference them in the first place
		// then we compare its votes between candidate 1 and 2
		if sp.hasVoted(*maxParent.GetHash()) {
			v := sp.votes[*maxParent.GetHash()]
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
		if b.GetParents().Size() > 1 {
			sp.updateTipVotes(voter, maxParent, votedPast)
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

func (sp *Spectre) updateTipVotes(voter *SpectreBlock, maxParent *SpectreBlock, votedPast *BlockDAG) {
	vb := votedPast.getBlock(voter.GetHash())
	voterParents := vb.GetParents()
	tipStack := stack.New()
	tipSet := NewHashSet()
	// take out all other tips and add their votes to child
	for _, h := range voterParents.GetMap() {
		if !h.(IBlock).GetHash().IsEqual(maxParent.GetHash()) && !h.(IBlock).GetHash().IsEqual(votedPast.getGenesis().GetHash()) {
			tipStack.Push(h)
			tipSet.Add(h.(IBlock).GetHash())
		}
	}
	for tipStack.Len() > 0 {
		tipHash := tipStack.Pop().(hash.Hash)
		tb := votedPast.getBlock(&tipHash)
		tipVoter := votedPast.instance.(*Spectre).sblocks[*tb.GetHash()]
		if sp.hasVoted(tipHash) {
			v := sp.votes[tipHash]
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
		for _, ib := range tb.GetParents().GetMap() {
			tp := *ib.(IBlock).GetHash()
			if tipSet.Has(&tp) {
				continue
			}
			only := true
			tpChildren := votedPast.getBlock(&tp).GetChildren()
			for _, tc := range tpChildren.GetMap() {
				if !tipSet.Has(tc.(IBlock).GetHash()) {
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
func (sp *Spectre) followParents(virtualBlock IBlock) (bool, bool, error) {
	// if all parents have voted in consistence, then just follow them
	consistent := true
	last := 0

	parents := NewIdSet()
	if virtualBlock == nil { // whole graph
		parents = sp.bd.tips
	} else { // past set of one node
		parents = virtualBlock.GetParents()
	}
	for _, id := range parents.GetMap() {
		ph := *id.(IBlock).GetHash()
		if sp.dangling.Has(&ph) {
			continue
		}
		if !sp.hasVoted(ph) {
			return true, sp.candidate1.GetHash().String() < sp.candidate2.GetHash().String(), fmt.Errorf("parent %v not ready", ph)
		}
		vote := -1
		if !sp.votes[ph] {
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
			sp.voteFirst(*h)
		} else {
			sp.voteSecond(*h)
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
func (sp *Spectre) VoteByBlock(virtualBlock IBlock) (bool, error) {
	if ok, res, err := sp.followParents(virtualBlock); ok {
		return res, err
	}

	votedPast := sp.votedPast(virtualBlock)
	// outer of votedPast, note the direction
	outer := sp.voteFromFutureSet(votedPast)
	firstWin := sp.voteFromPastSet(votedPast, outer)

	//4) if z is the virtual block of G then it will vote the same way as the vote of the majority of blocks in G.
	if virtualBlock != nil {
		h := virtualBlock.GetHash()
		if firstWin {
			sp.voteFirst(*h)
		} else {
			sp.voteSecond(*h)
		}
	}
	return firstWin, nil
}

//3) if z ∈ G is not in the future of either blocks then it will vote the same way as the vote of the majority of blocks in its own future.

// Update vote numbers of votedPast, e.g. for any node n in votedPast, update its Votes1 and Votes2
// note: because one node's vote number is dependent on its ancestor,
// we need update them from genesis of votePast down to its tips
func (sp *Spectre) voteFromFutureSet(votedPast *BlockDAG) *HashSet {
	tips := NewHashSet()

	vb := votedPast.getGenesis()
	unvisited := util.NewIterativeQueue()
	visited := NewHashSet()

	vChildren := votedPast.getBlock(vb.GetHash()).GetChildren()
	for _, ib := range vChildren.GetMap() {
		ch := *ib.(IBlock).GetHash()
		// only children with one parent (virtual block) will be selected as initial tips,
		// because they are only dependent on their single parent and their votes can be updated directly
		// e.g. note 22 in ByteBall2, 14, 20 can be initialized with 0 votes, but 15 cannot due to its multiple parents
		if votedPast.getBlock(&ch).GetParents().Size() == 1 {
			sb := votedPast.instance.(*Spectre).sblocks[ch]
			sb.Votes1, sb.Votes2 = 0, 0

			parents := sp.bd.getBlock(&ch).GetParents()
			if parents != nil {
				for ph := range parents.GetMap() {
					if !votedPast.hasBlockById(ph) {
						tips.Add(&ch)
						break
					}
				}
			} else {
				if !ch.IsEqual(votedPast.getGenesis().GetHash()) {
					log.Error("only virtual block can do without parents")
				}
			}
			cChildren := votedPast.getBlock(&ch).GetChildren()
			if cChildren != nil && cChildren.Size() > 0 {
				unvisited.Enqueue(ch)
			}
		}
	}

	// update votes with BFS order
	for unvisited.Len() > 0 {
		n := unvisited.Dequeue().(hash.Hash)
		childrenUpdated := true
		nChildren := votedPast.getBlock(&n).GetChildren()
		for _, id := range nChildren.GetMap() {
			ch := *id.(IBlock).GetHash()
			if !visited.Has(&ch) {
				if sp.updateVotes(votedPast, ch) {
					visited.Add(&ch)

					parents := sp.bd.getBlock(&ch).GetParents()
					if parents != nil {
						for ph := range parents.GetMap() {
							if !votedPast.hasBlockById(ph) {
								tips.Add(&ch)
								break
							}
						}
					} else {
						if !ch.IsEqual(votedPast.getGenesis().GetHash()) {
							log.Error("only virtual block can do without parents")
						}
					}
					children := votedPast.getBlock(&ch).GetChildren()
					if children != nil && children.Size() > 0 {
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

func (sp *Spectre) voteFromPastSet(votedPast *BlockDAG, tips *HashSet) bool {
	unvisited := util.NewIterativeQueue()
	outerNodes := NewHashSet()
	for h := range tips.GetMap() {
		hParents := sp.bd.getBlock(&h).GetParents()
		for _, ib := range hParents.GetMap() {
			ph := *ib.(IBlock).GetHash()
			if !outerNodes.Has(&ph) && !votedPast.hasBlockById(ib.(IBlock).GetID()) {
				unvisited.Enqueue(ph)
				outerNodes.AddPair(&ph, ib)
			}
		}
	}

	for unvisited.Len() > 0 {
		n := unvisited.Dequeue().(hash.Hash)

		if !sp.updateVotes(votedPast, n) {
			unvisited.Enqueue(n)
		} else {
			lastVote := 0
			consistent := true
			for o, ib := range outerNodes.GetMap() {
				if !votedPast.hasBlockById(ib.(IBlock).GetID()) {
					consistent = false
					break
				}
				sp := votedPast.instance.(*Spectre).sblocks[o]
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
				removing := NewHashSet()
				skip := NewHashSet()
				for o, ib := range outerNodes.GetMap() {
					if skip.Has(&o) {
						continue
					}
					allUpdated := votedPast.hasBlockById(ib.(IBlock).GetID())
					oParents := sp.bd.getBlock(&o).GetParents()
					for _, oib := range oParents.GetMap() {
						ph := *oib.(IBlock).GetHash()
						if !sp.updateVotes(votedPast, ph) {
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
					rChildren := votedPast.getBlock(&r).GetChildren()
					for _, c := range rChildren.GetMap() {
						if !outerNodes.Has(c.(IBlock).GetHash()) {
							outerNodes.Add(c.(IBlock).GetHash())
							unvisited.Enqueue(c)
						}
					}
				}
			}
		}
	}

	// break the tie
	return sp.candidate1.GetHash().String() < sp.candidate2.GetHash().String()
}

// add voter into voted past set
func (sp *Spectre) newVoter(vh hash.Hash, votedPast *BlockDAG) IBlock {
	sb := SpectreBlockData{hash: vh}
	sb.parents = []*hash.Hash{}
	vhChildren := sp.bd.getBlock(&vh).GetChildren()
	for _, ib := range vhChildren.GetMap() {
		hash := *ib.(IBlock).GetHash()
		if votedPast.hasBlockById(ib.(IBlock).GetID()) {
			sb.parents = append(sb.parents, &hash)
		}
	}
	if votedPast.hasBlockById(votedPast.getBlock(&vh).GetID()) {
		log.Error("has already voter ", vh)
	}
	//votedPast.AddBlock(&sb)
	block := Block{hash: *sb.GetHash(), weight: 1}
	if sb.parents != nil {
		block.parents = NewIdSet()
		for _, h := range sb.parents {
			hash := *h
			block.parents.Add(votedPast.getBlock(&hash).GetID())
			parent := votedPast.getBlock(&hash)
			parent.AddChild(&block)
		}
	}
	if votedPast.blocks == nil {
		votedPast.blocks = map[uint]IBlock{}
	}
	votedPast.blocks[block.id] = &block
	if votedPast.blockTotal == 0 {
		votedPast.genesis = *block.GetHash()
	}
	votedPast.blockTotal++
	votedPast.updateTips(&block)
	votedPast.instance.AddBlock(&block)
	return &block
}

//5) finally, (for the case where z equals x or y ), z votes for itself to succeed any block in past (z) and to precede any block outside past (z).
func (sp *Spectre) voteBySelf(b1 IBlock, b2 IBlock) {
	sp.voteFirst(*b1.GetHash())
	sp.voteSecond(*b2.GetHash())
}

// 1) if z ∈ G is in future (x) but not in future (y) then it will vote in favour of x (i.e., for x ≺y ).
func (sp *Spectre) voteByUniqueFutureSet(b1 IBlock, b2 IBlock) {
	fs1 := NewIdSet()
	sp.bd.getFutureSet(fs1, b1)

	fs2 := NewIdSet()
	sp.bd.getFutureSet(fs2, b2)

	for hf, ibf := range fs1.GetMap() {
		hfParents := sp.bd.getBlockById(hf).GetParents()
		for h, ib := range hfParents.GetMap() {
			if !fs1.Has(h) && !fs2.Has(h) {
				sp.dangling.Add(ib.(IBlock).GetHash())
			}
		}
		if fs2.Has(hf) {
			fs2.Remove(hf)
		} else {
			sp.voteFirst(*ibf.(IBlock).GetHash())
		}
	}

	for _, ib := range fs2.GetMap() {
		hfParents := ib.(IBlock).GetParents()
		for h, ibp := range hfParents.GetMap() {
			if !fs1.Has(h) && !fs2.Has(h) {
				sp.dangling.Add(ibp.(IBlock).GetHash())
			}
		}
		sp.voteSecond(*ib.(IBlock).GetHash())
	}
	sp.dangling.Remove(b1.GetHash())
	sp.dangling.Remove(b2.GetHash())
	for _, ib := range b1.GetParents().GetMap() {
		sp.dangling.Remove(ib.(IBlock).GetHash())
	}
	for _, ib := range b2.GetParents().GetMap() {
		sp.dangling.Remove(ib.(IBlock).GetHash())
	}
}

// Currently not supported
func (sp *Spectre) IsOnMainChain(b IBlock) bool {
	return false
}

// return the tip of main chain
func (sp *Spectre) GetMainChainTip() IBlock {
	return nil
}

func (sp *Spectre) GetMainChainTipId() uint {
	return 0
}

// return the main parent in the parents
func (sp *Spectre) GetMainParent(parents *IdSet) IBlock {
	return nil
}

// encode
func (sp *Spectre) Encode(w io.Writer) error {
	return nil
}

// decode
func (sp *Spectre) Decode(r io.Reader) error {
	return nil
}

func (sp *Spectre) Load(dbTx database.Tx) error {
	return nil
}

// IsDAG
func (sp *Spectre) IsDAG(parents []IBlock) bool {
	return true
}

// GetBlues
func (sp *Spectre) GetBlues(parents *IdSet) uint {
	return 0
}

// IsBlue
func (sp *Spectre) IsBlue(id uint) bool {
	return false
}

// getMaxParents
func (sp *Spectre) getMaxParents() int {
	return 0
}

// The main parent concurrency of block
func (sp *Spectre) GetMainParentConcurrency(b IBlock) int {
	return 0
}
