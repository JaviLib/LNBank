package main

import (
	"database/sql"
	"errors"
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

const LogTable = `
CREATE TABLE IF NOT EXISTS log (
    timestamp INTEGER NOT NULL ,
    type_id TINYINT NOT NULL,
    service VARCHAR(8) NOT NULL COLLATE NOCASE,
    desc TEXT NOT NULL COLLATE NOCASE
);
create index log_idx on log (timestamp, type_id, service COLLATE NOCASE);

CREATE VIEW last_hour_logs AS
  SELECT * FROM log
  WHERE timestamp >= strftime('%s', 'now', '-1 hour')
  ORDER BY timestamp DESC;
CREATE VIEW last_day_logs AS
  SELECT * FROM log
  WHERE timestamp >= strftime('%s', 'now', '-1 day')
  ORDER BY timestamp DESC;
CREATE VIEW last_week_logs AS
  SELECT * FROM log
  WHERE timestamp >= strftime('%s', 'now', '-7 day')
  ORDER BY timestamp DESC;
CREATE VIEW last_month_logs AS
  SELECT * FROM log
  WHERE timestamp >= strftime('%s', 'now', '-1 month')
  ORDER BY timestamp DESC;
CREATE VIEW last_year_logs AS
  SELECT * FROM log
  WHERE timestamp >= strftime('%s', 'now', '-1 year')
  ORDER BY timestamp DESC;

INSERT INTO log (timestamp, type_id, desc, service)
  VALUES (strftime('%s'), 0, 'Creation of LNBank', 'LNBank');
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
		return fmt.Sprintf("%d", int(lt))
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
var (
	Services map[string]Service
	Logdb    *sql.DB
)

var (
	ServiceRootDir string
	LogDBFile      string
)

func init() {
	var err error
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Panic, can't find UserHomeDir")
	}

	ServiceRootDir = filepath.Join(home, "LNBank")

	err = os.MkdirAll(ServiceRootDir, 0755)
	if err != nil {
		panic("Cannot create LNBank root dir")
	}

	LogDBFile = filepath.Join(ServiceRootDir, "log.sqlite3")
	// Check if the log database file exists
	_, err = os.Stat(LogDBFile)
	if err != nil { // if file does not exist
		Logdb, err = sql.Open("sqlite3", LogDBFile)
		if err != nil {
			fmt.Println("Error opening/creating SQLite DB:", err)
			panic("Panic, can't create log SQLite DB")
		} else {
			_, err = Logdb.Exec(LogTable)
			if err != nil {
				fmt.Println("Error creating table:", err)
				panic("Panic, can't create table")
			}
		}
	} else { // if it does exist, just open the db
		Logdb, err = sql.Open("sqlite3", LogDBFile)
		if err != nil {
			panic("Log DB is corrupted")
		}
	}

	// PREPARE SERVICES
	Services = make(map[string]Service)
}

// Check that the log is valid and modify it to make it valid. If not, returns
// a list of errors found.
func (log *Log) Validate() (err []error) {
	err = make([]error, 0)
	if log == nil {
		return append(err, errors.New("trying to log a null log entry"))
	}
	unixtimestamp := log.date.Unix()
	if unixtimestamp == 0 {
		err = append(err, errors.New("incorrect time '0'"))
		log.date = time.Now()
	} else if log.date.After(time.Now()) {
		err = append(err, errors.New("has a time in the future"))
		log.date = time.Now()
	}
	if log.logType > DEBUG {
		err = append(err, fmt.Errorf("incorrect log type %v", log.logType))
		log.logType = NORMAL
	}
	if log.desc == "" {
		err = append(err, errors.New("no description"))
		log.desc = "undefined"
	}
	if log.service == "" {
		err = append(err, errors.New("no service provided"))
		log.service = "LNBANK"
	}
	if len(err) == 0 {
		return nil
	}
	return err
}

// Log to the database, it returns a list of errors found and if any of them is fatal
func LogToDb(log *Log) (errs []error, fatal bool) {
	// Validate the log entry and return any errors found
	errs = log.Validate()
	// Prepare a statement for inserting the log into the database
	stmt, err := Logdb.Prepare(
		"INSERT INTO log (timestamp, type_id, desc, service) VALUES (?, ?, ?, ?)",
	)
	if err != nil {
		return append(errs, err), true
	}
	defer stmt.Close()

	// Execute the prepared statement with the log data
	_, err = stmt.Exec(log.date.Unix(), log.logType, log.desc, log.service)
	if err != nil {
		return append(errs, err), true
	}

	return errs, false
}
