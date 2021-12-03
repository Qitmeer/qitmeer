package main

import (
	"github.com/Qitmeer/qng-core/database"
	"github.com/Qitmeer/qng-core/log"
	"github.com/Qitmeer/qng-core/params"
	"path/filepath"
)

const (
	// blockDbNamePrefix is the prefix for the block database name.  The
	// database type is appended to this value to form the full block
	// database name.
	blockDbNamePrefix = "blocks"
)

// loadBlockDB loads (or creates when needed) the block database taking into
// account the selected database backend and returns a handle to it.  It also
// contains additional logic such warning the user if there are multiple
// databases which consume space on the file system and ensuring the regression
// test database is clean when in regression test mode.
func LoadBlockDB(cfg *Config) (database.DB, error) {
	// The database name is based on the database type.
	dbPath := blockDbPath(cfg.DbType,cfg)

	log.Info("Loading block database", "dbPath", dbPath)
	db, err := database.Open(cfg.DbType, dbPath, params.ActiveNetParams.Net)
	if err != nil {
		return nil, err
	}
	log.Info("Block database loaded")
	return db, nil
}

// blockDbPath returns the path to the block database given a database type.
func blockDbPath(dbType string,cfg *Config) string {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + dbType
	dbPath := filepath.Join(cfg.DataDir, dbName)
	return dbPath
}
