// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dcr

// BlockOneLedgerMainNet is the block one output ledger for the main
// network.
var BlockOneLedgerMainNet = []*TokenPayout{
	{"DsaAKsMvZ6HrqhmbhLjV9qVbPkkzF5daowT", 5000 * 1e8},
	{"DsaAtdVq78c4zM7fhHbbRHPd8UoZn91ZoR3", 5000 * 1e8},
	{"DsabRUzLFikdEx62WeSyFjvqVg7i8DkuTjo", 5000 * 1e8},
	{"DsaiQhBpjr2Xq9T6Zso5vAFEPgUEWVFKVC9", 5000 * 1e8},
	{"DsaPv23JyLgjrYFmXnHdiou1DYpyAykCkZb", 5000 * 1e8},
	{"DsaSkJbREyuYvFYgJdatxRJNSREXFWtJc5H", 5000 * 1e8},
	{"DsaxUwRcbWg559c1gWECY9Ei5VPTrd9vXrb", 5000 * 1e8},
	{"DsbaqJC3h9DJmMGdy8hNddNQhiFJi6p5VoM", 5000 * 1e8},
	{"DsbC3ywTnt97xMdhwq9Vma75wZotpFWWS1y", 5000 * 1e8},
	{"DsbcjZNGqTeLhKnBapYTFwqmPxCwEeXCoZ1", 5000 * 1e8},
	{"DsbDKNAwbFiBfaCeCJEmfmr8USbks2N8Mms", 5000 * 1e8},
}

// BlockOneLedgerTestNet is the block one output ledger for the test
// network.
var BlockOneLedgerTestNet = []*TokenPayout{
	{"TsmWaPM77WSyA3aiQ2Q1KnwGDVWvEkhipBc", 100000 * 1e8},
}

// BlockOneLedgerTestNet2 is the block one output ledger for the 2nd test
// network.
var BlockOneLedgerTestNet2 = []*TokenPayout{
	{"TsT5rhHYqJF7sXouh9jHtwQEn5YJ9KKc5L9", 100000 * 1e8},
}

// BlockOneLedgerSimNet is the block one output ledger for the simulation
// network. See under "Decred organization related parameters" in params.go
// for information on how to spend these outputs.
var BlockOneLedgerSimNet = []*TokenPayout{
	{"Sshw6S86G2bV6W32cbc7EhtFy8f93rU6pae", 100000 * 1e8},
	{"SsjXRK6Xz6CFuBt6PugBvrkdAa4xGbcZ18w", 100000 * 1e8},
	{"SsfXiYkYkCoo31CuVQw428N6wWKus2ZEw5X", 100000 * 1e8},
}
