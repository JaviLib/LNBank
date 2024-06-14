package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Typical service log types that will output and store in db
type Log struct {
	date    time.Time
	desc    string
	logType LogType
	service string
}

// log types
type LogType int8

const (
	NORMAL = iota
	FATAL
	ERROR
	WARNING
	INFO
	DEBUG
)
const LogTable = `CREATE TABLE IF NOT EXISTS log (
    date DATETIME NOT NULL PRIMARY KEY,
    log_type_id INT NOT NULL,
    desc TEXT  NOT NULL,
    service VARCHAR(16) NOT NULL
);
`

func (lt LogType) String() string {
	switch lt {
	case FATAL:
		return "FATAL"
	case WARNING:
		return "WARNING"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case ERROR:
		return "ERROR"
	default:
		return "NORMAL"
	}
}

type Service interface {
	start(
		onReady func(),
		onFatal func(*Log),
		onLog func(*Log),
	)
	stop(onStopped func())
	getConfigFile() (string, error)

	// This is used to format the log before showing to the user or insert in db
	fmtLog(LogType, string) *Log
}

// All services running are here
var Services map[string]Service
var Logdb *sql.DB

var (
	ServiceRootDir string
	LogDBFile      string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Panic, can't find UserHomeDir")
	}

	ServiceRootDir = filepath.Join(home, "LNBank")

	err = os.MkdirAll(ServiceRootDir, 0755)
	if err != nil {
		panic("Cannot create root dir")
	}

	LogDBFile = filepath.Join(ServiceRootDir, "log.sqlite3")

	Logdb, err := sql.Open("sqlite3", LogDBFile)
	if err != nil {
		fmt.Println("Error opening/creating SQLite DB:", err)
		panic("Panic, can't create log SQLite DB")
	}
	defer Logdb.Close()

	_, err = Logdb.Exec(LogTable)
	if err != nil {
		fmt.Println("Error creating table:", err)
		panic("Panic, can't create table")
	}

	// PREPARE SERVICES
	Services = make(map[string]Service)

}
