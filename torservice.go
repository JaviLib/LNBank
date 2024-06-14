package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultTorConfig = `
SocksPort localhost:9050
ControlPort localhost:9051
MaxClientCircuitsPending 1024	
	`
)

var TorConfigPath string
var TorConfigFile string

type TorService struct {
	onReady func()
	onFatal func(*Log)
	onLog   func(*Log)
}

func (l TorService) fmtLog(lt LogType, desc string) *Log {
	log := &Log{time.Now(), desc, lt, "Tor"}
	return log
}

// Implement the Service interface
func (ts TorService) start(onReady func(), onFatal func(*Log), onLog func(*Log)) {
	if onReady == nil || onFatal == nil || onLog == nil {
		panic("Some function parameter to tor start method is missing. Bad implementation.")
	}
	ts.onReady = onReady
	ts.onFatal = onFatal
	ts.onLog = onLog

	TorConfigPath = filepath.Join(ServiceRootDir, "tor")
	err := os.MkdirAll(TorConfigPath, 0755) // make sure the log directory exists
	if err != nil {
		log := ts.fmtLog(FATAL, fmt.Sprintf("Cannot create tor directory %v: ", err))
		go onLog(log)
		go onFatal(log)
		return
	}

	TorConfigFile = filepath.Join(TorConfigPath, "torrc")
	if _, err := os.Stat(TorConfigFile); os.IsNotExist(err) {
		err = os.WriteFile(TorConfigFile, []byte(defaultTorConfig), 0644)
		if err != nil {
			log := ts.fmtLog(FATAL,
				fmt.Sprintf("Cannot create tor configuration file %v: ", err))
			go onLog(log)
			go onFatal(log)
			return
		}
	}
}

func (ts TorService) stop(onStopped func()) {
	// TODO: end the program
	ts.onLog(ts.fmtLog(NORMAL, "Stopping Tor"))
	onStopped()
}

func (ts TorService) getConfigFile() (string, error) {
	return TorConfigFile, nil // assuming that the torrc file path is stored as a global variable TorConfigPath
}
