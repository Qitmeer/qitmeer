// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2015-2016 The Decred developers
// Copyright (c) 2013-2016 The btcsuite developers

package main

import (
	"runtime"
	"runtime/debug"
	"os"
	"path/filepath"
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/node"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/database"
	 _ "github.com/noxproject/nox/database/ffldb"
)

const (
	// blockDbNamePrefix is the prefix for the block database name.  The
	// database type is appended to this value to form the full block
	// database name.
	blockDbNamePrefix = "blocks"
)

var (
	cfg *config.Config
)

func main() {

	// Initialize the goroutine count,  Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Block and transaction processing can cause bursty allocations.  This
	// limits the garbage collector from excessively overallocating during
	// bursts.  This value was arrived at with the help of profiling live
	// usage.
	debug.SetGCPercent(20)

	// Work around defer not working after os.Exit()
	if err := noxdMain(nil); err != nil {
		os.Exit(1)
	}
}

// noxdMain is the real main function for noxd.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional nodeChan parameter is mainly used by the service code to be
// notified with the node once it is setup so it can gracefully stop it when
// requested from the service control manager.
func noxdMain(nodeChan chan<- *node.Node) error {

	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg

	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := interruptListener()
	defer log.Info("Shutdown complete")

	// Show version and home dir at startup.
	log.Info("System info", "Nox Version", version(), "Go version",runtime.Version())
	log.Info("System info", "Home dir", cfg.HomeDir)
	if cfg.NoFileLogging {
		log.Info("File logging disabled")
	}

	// Load the block database.
	db, err := loadBlockDB()
	if err != nil {
		log.Error("load block database","error", err)
		return err
	}
	defer func() {
		// Ensure the database is sync'd and closed on shutdown.
		log.Info("Gracefully shutting down the database...")
		db.Close()
	}()

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}

	// Create node and start it.
	node, err := makeNode(db, activeNetParams.Params, interrupt)
	if err != nil {
		log.Error("Unable to start server","listeners",cfg.Listeners,"error", err)
		return err
	}
	defer func() {
		log.Info("Gracefully shutting down the server...")
		node.Stop()
		node.WaitForShutdown()
		log.Info("Server shutdown complete")
	}()
	node.Start()
	if nodeChan != nil {
		nodeChan <- node
	}
	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}

// loadBlockDB loads (or creates when needed) the block database taking into
// account the selected database backend and returns a handle to it.  It also
// contains additional logic such warning the user if there are multiple
// databases which consume space on the file system and ensuring the regression
// test database is clean when in regression test mode.
func loadBlockDB() (database.DB, error) {

	// The database name is based on the database type.
	dbPath := blockDbPath(cfg.DbType)

	log.Info("Loading block database", "dbPath", dbPath)
	db, err := database.Open(cfg.DbType, dbPath, activeNetParams.Net)
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
		db, err = database.Create(cfg.DbType, dbPath, activeNetParams.Net)
		if err != nil {
			return nil, err
		}
	}
	log.Info("Block database loaded")
	return db, nil
}

// blockDbPath returns the path to the block database given a database type.
func blockDbPath(dbType string) string {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + dbType
	dbPath := filepath.Join(cfg.DataDir, dbName)
	return dbPath
}

// newNode returns a new nox node which configured to listen on addr for the
// nox network type specified by the network Params.
func makeNode(db database.DB, params *params.Params, interrupt <-chan struct{}) (*node.Node, error) {
	node := node.Node{
	}
	return &node,nil
}


