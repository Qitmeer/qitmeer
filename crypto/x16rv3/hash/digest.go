package hash

type Digest interface {
	// See hash.Hash
	Hash

	// Close the digest by writing the last bits and storing the hash
	// in dst. This prepares the digest for reuse, calls Hash.Reset.
	Close(dst []byte, bits uint8, bcnt uint8) error
}
