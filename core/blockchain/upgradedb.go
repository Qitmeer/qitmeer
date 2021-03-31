package blockchain

import (
	"fmt"
)

// update db to new version
func (b *BlockChain) upgradeDB() error {
	if b.dbInfo.version == currentDatabaseVersion {
		return nil
	}
	return fmt.Errorf("You can cleanup your block data base by '--cleanup'.Your data is too old (%d -> %d). ", b.dbInfo.version, currentDatabaseVersion)
}
