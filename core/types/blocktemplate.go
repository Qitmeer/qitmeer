package types

// this standard target use for miner to verify Their work
// for different pow work diff
// blake2bd on hash compare hash <= target
// cuckoo diff formula ：scale * 2^64 / hash(front 8bytes) >= base diff
type PowDiffStandard struct {
	//blake2b diff hash target
	Blake2bDTarget         uint32
	X16rv3DTarget          uint32
	X8r16DTarget           uint32
	QitmeerKeccak256Target uint32

	//cuckoo base difficultuy
	CuckarooBaseDiff  uint64
	CuckatooBaseDiff  uint64
	CuckaroomBaseDiff uint64

	//cuckoo hash convert diff scale
	CuckarooDiffScale  uint64
	CuckatooDiffScale  uint64
	CuckaroomDiffScale uint64
}

// BlockTemplate houses a block that has yet to be solved along with additional
// details about the fees and the number of signature operations for each
// transaction in the block.
type BlockTemplate struct {
	// Block is a block that is ready to be solved by miners.  Thus, it is
	// completely valid with the exception of satisfying the proof-of-work
	// requirement.
	Block *Block

	// Fees contains the amount of fees each transaction in the generated
	// template pays in base units.  Since the first transaction is the
	// coinbase, the first entry (offset 0) will contain the negative of the
	// sum of the fees of all other transactions.
	Fees []int64

	// SigOpCounts contains the number of signature operations each
	// transaction in the generated template performs.
	SigOpCounts []int64

	// Height is the height at which the block template connects to the main
	// chain.
	Height uint64

	// Blues is the count of the blue set in the block past set to
	// the DAG
	Blues int64

	// ValidPayAddress indicates whether or not the template coinbase pays
	// to an address or is redeemable by anyone.  See the documentation on
	// NewBlockTemplate for details on which this can be useful to generate
	// templates without a coinbase payment address.
	ValidPayAddress bool

	//pow diff standard
	PowDiffData PowDiffStandard
}
