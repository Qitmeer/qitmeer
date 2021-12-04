/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

// Package dbnamespace contains constants that define the database namespaces
// for the purpose of the blockdag, so that external callers may easily access
// this data.
package blockdag

import (
	"encoding/binary"
)

var (
	// ByteOrder is the preferred byte order used for serializing numeric
	// fields for storage in the database.
	ByteOrder = binary.LittleEndian

	// BlockIndexBucketName is the name of the db bucket used to house the
	// block which consists of metadata for all known blocks in DAG.
	BlockIndexBucketName = []byte("blockidx")

	// DagInfoBucketName is the name of the db bucket used to house the
	// dag information
	DagInfoBucketName = []byte("daginfo")

	// DAG Main Chain Blocks
	DagMainChainBucketName = []byte("dagmainchain")

	// OrderIdBucketName is the name of the db bucket used to house to
	// the block order -> block DAG Id.
	OrderIdBucketName = []byte("orderid")

	// BlockIdBucketName is the name of the db bucket used to house to
	// the block hash -> block DAG Id.
	BlockIdBucketName = []byte("blockid")
)
