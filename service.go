package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Typical service log types that will output and store in db
type Log struct {
	date    time.Time
	desc    string
	service string
	logType LogType
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
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 10;

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

var insertStmt *sql.Stmt

// Log to the database, it returns a list of errors found and if any of them is fatal
func LogToDb(log *Log) (errs []error, fatal bool) {
	// Validate the log entry and return any errors found
	errs = log.Validate()

	// Prepare a statement for inserting the log into the database but only if it doesn't exist
	if insertStmt == nil {
		var err error
		insertStmt, err = Logdb.Prepare(
			"INSERT INTO log (timestamp, type_id, desc, service) VALUES (?, ?, ?, ?)",
		)
		if err != nil {
			insertStmt = nil
			return append(errs, err), true
		}
		// TODO defer stmt.Close()
	}

	_, err := insertStmt.Exec(log.date.Unix(), log.logType, log.desc, log.service)
	if err != nil {
		return append(errs, err), true
	}

	// Errors in logs are also logged to the db
	if errs != nil {
		LogToDb(&Log{
			date:    time.Now(),
			service: "LNBank",
			desc:    fmt.Sprintf("Service %v provided an incorrect log: %v", log.service, errs),
			logType: WARNING,
		})
	}

	return errs, false
}

type LogQuery struct {
	next  func() (Log, error)
	close func()
	getn  func(n uint) ([]Log, error)
}

// Query the logs and returns a closure useful to iterate over the rows.
// desc is no case and limit is the maximum number of rows allowed.
func QueryLog(duration time.Duration,
	logtypes []LogType,
	services []string,
	desc string,
	limit uint,
) (LogQuery, error) {
	var conditions strings.Builder
	conditions.WriteString("SELECT timestamp, type_id, service, desc FROM log WHERE timestamp >= ?")

	if logtypes != nil {
		conditions.WriteString(" AND (")
		first := true
		for range logtypes {
			if !first {
				conditions.WriteString(" OR")
			} else {
				first = false
			}
			conditions.WriteString(" type_id=?")
		}
		conditions.WriteString(" )")
	}

	if services != nil {
		conditions.WriteString(" AND (")
		first := true
		for range services {
			if !first {
				conditions.WriteString(" OR")
			} else {
				first = false
			}
			conditions.WriteString(" service=?")
		}
		conditions.WriteString(" )")
	}
	if desc != "" {
		conditions.WriteString(" AND desc LIKE ?")
	}

	conditions.WriteString(" ORDER BY timestamp DESC")

	if limit != 0 {
		conditions.WriteString(" LIMIT ?")
	}

	var params []interface{}

	// timestamp
	params = append(params, time.Now().Add(-duration).Unix())

	// logtypes
	for _, logtype := range logtypes {
		params = append(params, logtype)
	}

	// services
	for _, service := range services {
		params = append(params, service) // services
	}

	// desc and limit
	if desc != "" {
		params = append(params, "%"+desc+"%")
	}
	if limit != 0 {
		params = append(params, limit)
	}

	result, err := Logdb.Query(conditions.String(), params...)
	if err != nil {
		return LogQuery{}, err
	}
	logQuery := LogQuery{
		next: func() (Log, error) {
			var log Log
			if !result.Next() {
				result.Close()
				return log, errors.New("end")
			}
			var unixdate int64
			err := result.Scan(&unixdate, &log.logType, &log.service, &log.desc)
			if err != nil {
				return Log{}, err
			}

			log.date = time.Unix(unixdate, 0)
			return log, nil
		},
		close: func() { result.Close() },
		getn: func(n uint) ([]Log, error) {
			var logs []Log
			for result.Next() {
				var unixdate int64
				var log Log
				err := result.Scan(&unixdate, &log.logType, &log.service, &log.desc)
				if err != nil {
					_, _ = LogToDb(&Log{
						date:    time.Now(),
						service: "LNBank",
						logType: WARNING,
						desc:    "unable to scan from the db: " + err.Error(),
					})
					return []Log{}, errors.New("unable to scan from the db: " + err.Error())
				}
				log.date = time.Unix(unixdate, 0)
				logs = append(logs, log)
			}
			return logs, nil
		},
	}
	return logQuery, nil
}
