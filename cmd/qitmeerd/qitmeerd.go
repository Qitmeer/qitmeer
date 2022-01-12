// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2015-2016 The Decred developers
// Copyright (c) 2013-2016 The btcsuite developers

package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/config"
	_ "github.com/Qitmeer/qitmeer/database/ffldb"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/node"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/services/index"
	"github.com/Qitmeer/qitmeer/version"
	"os"
	"runtime"
	"runtime/debug"
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
		log.Error(err.Error())
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
	cfg, _, err := common.LoadConfig()
	if err != nil {
		return err
	}

	defer func() {
		if log.LogWrite() != nil {
			log.LogWrite().Close()
		}
	}()
	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := interruptListener()
	defer log.Info("Shutdown complete")

	// Show version and home dir at startup.
	log.Info("System info", "Qitmeer Version", version.String(), "Go version", runtime.Version())
	log.Info("System info", "Home dir", cfg.HomeDir)

	if cfg.NoFileLogging {
		log.Info("File logging disabled")
	}

	// Load the block database.
	db, err := common.LoadBlockDB(cfg)
	if err != nil {
		log.Error("load block database", "error", err)
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
		db.Close()
		common.CleanupBlockDB(cfg)
		return nil
	}

	// Create node and start it.
	n, err := node.NewNode(cfg, db, params.ActiveNetParams.Params, shutdownRequestChannel)
	if err != nil {
		log.Error("Unable to start server", "listeners", cfg.Listener, "error", err)
		return err
	}
	err = n.RegisterService()
	if err != nil {
		return err
	}
	defer func() {
		log.Info("Gracefully shutting down the server...")
		err := n.Stop()
		if err != nil {
			log.Warn("node stop error", "error", err)
		}
		n.WaitForShutdown()
	}()
	err = n.Start()
	if err != nil {
		log.Error("Uable to start server", "error", err)
		return err
	}
	showLogo(cfg)
	//
	if nodeChan != nil {
		nodeChan <- n
	}
	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}

func showLogo(cfg *config.Config) {
	logo := `

         .__  __                                                                    
    _____|__|/  |_  _____   ____   ___________    Qitmeer %s
   / ____/  \   __\/     \_/ __ \_/ __ \_  __ \   Port: %d
  < <_|  |  ||  | |  Y Y  \  ___/\  ___/|  | \/   PID : %d
   \__   |__||__| |__|_|  /\___  >\___  >__|      Network : %s                      
      |__|              \/     \/     \/          https://github.com/Qitmeer/qitmeer



`
	fmt.Printf(logo, version.String(), cfg.P2PTCPPort, os.Getpid(), params.ActiveNetParams.Name)
}
