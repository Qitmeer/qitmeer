package main

import (
	"github.com/Qitmeer/qng-core/database"
	"github.com/Qitmeer/qng-core/log"
	"github.com/Qitmeer/qng-core/params"
	"os"
	"path/filepath"
)

const (
	// blockDbNamePrefix is the prefix for the block database name.  The
	// database type is appended to this value to form the full block
	// database name.
	blockDbNamePrefix = "blocks"
)

var (
	// DebugAddrInfoBucketName is the name of the db bucket used to house the
	// debug address info
	DebugAddrInfoBucketName = []byte("debugaddrinfo")

	// DebugAddrBucketName is the name of the db bucket used to house the
	// debug address
	DebugAddrBucketName = []byte("debugaddr")
)

// loadBlockDB loads (or creates when needed) the block database taking into
// account the selected database backend and returns a handle to it.  It also
// contains additional logic such warning the user if there are multiple
// databases which consume space on the file system and ensuring the regression
// test database is clean when in regression test mode.
func LoadBlockDB(DbType string, DataDir string, nocreate bool) (database.DB, error) {
	// The database name is based on the database type.
	dbPath := blockDbPath(DbType, DataDir)

	log.Trace("Loading block database", "dbPath", dbPath)
	db, err := database.Open(DbType, dbPath, params.ActiveNetParams.Net)
	if err != nil {
		if nocreate {
			// Return the error if it's not because the database doesn't
			// exist.
			if dbErr, ok := err.(database.Error); !ok || dbErr.ErrorCode !=
				database.ErrDbDoesNotExist {

				return nil, err
			}
			// Create the db if it does not exist.
			err = os.MkdirAll(DataDir, 0700)
			if err != nil {
				return nil, err
			}
			db, err = database.Create(DbType, dbPath, params.ActiveNetParams.Net)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	log.Trace("Block database loaded")
	return db, nil
}

// blockDbPath returns the path to the block database given a database type.
func blockDbPath(DbType string, DataDir string) string {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + DbType
	dbPath := filepath.Join(DataDir, dbName)
	return dbPath
}
