package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

//go:embed tor-static/tor/dist/bin/tor
var torExe []byte

func InstallTorExe() error {
	path := ServiceRootDir + "/bin/tor"
	info, err := os.Stat(path)
	if err != nil || info.Size() != int64(len(torExe)) {
		err := os.MkdirAll(ServiceRootDir+"/bin", 0755)
		if err != nil {
			return errors.New("Cannot create bin/tor directory: " + err.Error())
		}
		err = os.WriteFile(ServiceRootDir+"/bin/tor", torExe, 0755)
		if err != nil {
			return errors.New("Cannot install Tor executable: " + err.Error())
		}
	}
	// gc the memory
	torExe = nil
	return nil
}

const (
	defaultTorConfig = `
SocksPort localhost:9050
ControlPort localhost:9051
MaxClientCircuitsPending 1024	
	`
)

var (
	TorConfigPath string
	TorConfigFile string
)

type TorService struct {
	onReady func()
	onFatal func(*Log)
	onLog   func(*Log)
}

func (l TorService) fmtLog(lt LogType, desc string) *Log {
	log := &Log{
		date:    time.Now(),
		desc:    desc,
		logType: lt,
		service: "Tor",
	}
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

	if err := InstallTorExe(); err != nil {
		log := ts.fmtLog(FATAL, err.Error())
		go onLog(log)
		go onFatal(log)
		return
	}

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
