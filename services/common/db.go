package common

import (
	"fmt"
	"github.com/Qitmeer/qng-core/config"
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

// loadBlockDB loads (or creates when needed) the block database taking into
// account the selected database backend and returns a handle to it.  It also
// contains additional logic such warning the user if there are multiple
// databases which consume space on the file system and ensuring the regression
// test database is clean when in regression test mode.
func LoadBlockDB(cfg *config.Config) (database.DB, error) {
	// The database name is based on the database type.
	dbPath := blockDbPath(cfg.DbType, cfg)

	log.Info("Loading block database", "dbPath", dbPath)
	db, err := database.Open(cfg.DbType, dbPath, params.ActiveNetParams.Net)
	if err != nil {
		// Return the error if it's not because the database doesn't
		// exist.
		if dbErr, ok := err.(database.Error); !ok || dbErr.ErrorCode !=
			database.ErrDbDoesNotExist {

			return nil, err
		}
		// Create the db if it does not exist.
		err = os.MkdirAll(cfg.DataDir, 0700)
		if err != nil {
			return nil, err
		}
		db, err = database.Create(cfg.DbType, dbPath, params.ActiveNetParams.Net)
		if err != nil {
			return nil, err
		}
	}
	log.Info("Block database loaded")
	return db, nil
}

// blockDbPath returns the path to the block database given a database type.
func blockDbPath(dbType string, cfg *config.Config) string {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + dbType
	dbPath := filepath.Join(cfg.DataDir, dbName)
	return dbPath
}

// removeBlockDB removes the existing database
func removeBlockDB(dbPath string) error {
	// Remove the old database if it already exists.
	fi, err := os.Stat(dbPath)
	if err == nil {
		log.Info(fmt.Sprintf("Removing block database from '%s'", dbPath))
		if fi.IsDir() {
			err := os.RemoveAll(dbPath)
			if err != nil {
				return err
			}
		} else {
			err := os.Remove(dbPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CleanupBlockDB(cfg *config.Config) {
	dbPath := blockDbPath(cfg.DbType, cfg)
	err := removeBlockDB(dbPath)
	if err != nil {
		log.Error(err.Error())
	}
	log.Info("Finished cleanup")
}
