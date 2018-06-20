// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2015-2016 The Decred developers
// Copyright (c) 2013-2016 The btcsuite developers

package main

import (
	"runtime"
	"runtime/debug"
	"fmt"
	"os"
	"github.com/noxproject/nox/log"
	"io"
	"github.com/noxproject/nox/log/term"
	"github.com/mattn/go-colorable"
)

// winServiceMain is only invoked on Windows.  It detects when dcrd is running
// as a service and reacts accordingly.
var winServiceMain func() (bool, error)

var glogger *log.GlogHandler

func init() {

	// init a colorful logger if possible
	usecolor := term.IsTty(os.Stderr.Fd()) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if usecolor {
		output = colorable.NewColorableStderr()
	}
	glogger = log.NewGlogHandler(log.StreamHandler(output, log.TerminalFormat(usecolor)))

	// print log location (file:line) (useful for debug)
	// TODO config & command line flag
	log.PrintOrigins(false)

	// set log level to info
	// TODO config & comand line flag
	glogger.Verbosity(log.LvlInfo)

	log.Root().SetHandler(glogger)

	// Initialize the goroutine count,  Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	// Block and transaction processing can cause bursty allocations.  This
	// limits the garbage collector from excessively overallocating during
	// bursts.  This value was arrived at with the help of profiling live
	// usage.
	debug.SetGCPercent(20)

	// Call serviceMain on Windows to handle running as a service.  When
	// the return isService flag is true, exit now since we ran as a
	// service.  Otherwise, just fall through to normal operation.
	if runtime.GOOS == "windows" {
		isService, err := winServiceMain()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if isService {
			os.Exit(0)
		}
	}

	// Work around defer not working after os.Exit()
	if err := noxdMain(nil); err != nil {
		os.Exit(1)
	}
}

// noxdMain is the real main function for noxd.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func noxdMain(serverChan chan<- *peerServer) error {

	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}

	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := interruptListener()
	defer log.Info("Shutdown complete")

	// Show version and home dir at startup.
	log.Info("System info", "Version", version(), "Go version",runtime.Version())
	log.Info("System info", "Home dir", cfg.HomeDir)
	if cfg.NoFileLogging {
		log.Info("File logging disabled")
	}

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}



