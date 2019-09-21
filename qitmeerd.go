// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2015-2016 The Decred developers
// Copyright (c) 2013-2016 The btcsuite developers

package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/config"
	"github.com/Qitmeer/qitmeer-lib/core/message"
	"github.com/Qitmeer/qitmeer-lib/log"
	_ "github.com/Qitmeer/qitmeer/database/ffldb"
	"github.com/Qitmeer/qitmeer/node"
	"github.com/Qitmeer/qitmeer/services/index"
	"github.com/Qitmeer/qitmeer/version"
	"os"
	"runtime"
	"runtime/debug"
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
	if err := qitmeerdMain(nil); err != nil {
		os.Exit(1)
	}
}

// qitmeerdMain is the real main function for qitmeerd.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional nodeChan parameter is mainly used by the service code to be
// notified with the node once it is setup so it can gracefully stop it when
// requested from the service control manager.
func qitmeerdMain(nodeChan chan<- *node.Node) error {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg
	defer func() {
		if logWrite != nil {
			logWrite.Close()
		}
	}()
	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := interruptListener()
	defer log.Info("Shutdown complete")

	// Show version and home dir at startup.
	log.Info("System info", "Qitmeer Version", version.String(), "Go version",runtime.Version())
	log.Info("System info","UUID",message.UUID)
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
	// Drop indexes and exit if requested.
	if cfg.DropAddrIndex {
		if err := index.DropAddrIndex(db, interrupt); err != nil {
			log.Error("%v", err)
			return err
		}

		return nil
	}
	if cfg.DropTxIndex {
		if err := index.DropTxIndex(db, interrupt); err != nil {
			log.Error(fmt.Sprintf("%v", err))
			return err
		}

		return nil
	}

	// Cleanup the block database
	if cfg.Cleanup {
		cleanupBlockDB()
		return nil
	}

	// Create node and start it.
	n, err := node.NewNode(cfg,db,activeNetParams.Params,shutdownRequestChannel)
	if err != nil {
		log.Error("Unable to start server","listeners",cfg.Listeners,"error", err)
		return err
	}
	err = n.RegisterService()
	if err != nil {
		return err
	}
	defer func() {
		log.Info("Gracefully shutting down the server...")
		err := n.Stop()
		if err!=nil{
			log.Warn("node stop error","error",err)
		}
		n.WaitForShutdown()
	}()
	err = n.Start()
	if err != nil {
		log.Error("Uable to start server", "error",err)
		return err
	}

	if nodeChan != nil {
		nodeChan <- n
	}
	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}
