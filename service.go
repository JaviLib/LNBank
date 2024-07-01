package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
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

func (lt LogType) String() string {
	switch lt {
	case FATAL:
		return "❌ ❌ ❌ FATAL"
	case WARNING:
		return "❌ WARNING"
	case INFO:
		return "✔"
	case DEBUG:
		return "🐛"
	case ERROR:
		return "❌ ❌ ERROR"
	default:
		return fmt.Sprintf("%d", int(lt))
	}
}

func (l Log) String() string {
	return fmt.Sprintf("%v %v %v: %v", l.service, l.logType, l.date.Format(time.Stamp), l.desc)
}

type Service interface {
	start(
		ctx context.Context,
		onReady func(),
		onStop func(*Log),
		onLog func(*Log),
	)
	onLogHook() func(*Log)
	onReadyHook() func()
	getConfigFile() (string, error)
	name() string

	// given a line of text, return a log
	parseLogEntry(string) (Log, error)
	// given a text from a log, determine if the service is ready to accept connections
	isReady(string) bool

	// This is used to format the log before showing to the user or insert in db
	fmtLog(LogType, string) *Log
}

// All services running are here
var (
	Services map[string]Service
	DB       *sql.DB
)

var (
	ServiceRootDir string
	DBFile         string
)

var (
	ServicesContext    context.Context
	ServicesCancelFunc context.CancelFunc
)

func init() {
	var err error

	ServicesContext, ServicesCancelFunc = context.WithCancel(context.Background())

	home, err := os.UserHomeDir()
	if err != nil {
		panic("Panic, can't find UserHomeDir")
	}

	ServiceRootDir = filepath.Join(home, "LNBank")

	err = os.MkdirAll(ServiceRootDir, 0755)
	if err != nil {
		panic("Cannot create LNBank root dir")
	}

	DBFile = filepath.Join(ServiceRootDir, "lnbank.sqlite3")
	// Check if the log database file exists
	_, err = os.Stat(DBFile)
	if err != nil { // if file does not exist
		DB, err = sql.Open("sqlite3", DBFile)
		if err != nil {
			fmt.Println("Error opening/creating SQLite DB:", err)
			panic("Panic, can't create LNBank SQLite DB")
		} else {
			_, err = DB.ExecContext(ServicesContext, LogTable)
			if err != nil {
				fmt.Println("Error creating table:", err)
				panic("Panic, can't create table")
			}
		}
	} else { // if it does exist, just open the db
		DB, err = sql.Open("sqlite3", DBFile)
		if err != nil {
			panic("Log DB is corrupted")
		}
	}

	// PREPARE SERVICES
	Services = make(map[string]Service)
}

func UnzipReader(rd *bytes.Reader, size int64, dest string, onLog func(*Log)) error {
	r, err := zip.NewReader(rd, size)
	if err != nil {
		return errors.New("Cannot open embedded file: " + err.Error())
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(filepath.Join(dest, f.Name), os.ModePerm)
			if err != nil {
				err := errors.New("Cannot create directory: " + err.Error())
				go onLog(&Log{
					date:    time.Now(),
					service: "LNBank",
					logType: FATAL,
					desc:    err.Error(),
				})
				return err
			}
			continue
		}
		final, err := os.Create(filepath.Join(dest, f.Name))
		if err != nil {
			err := errors.New("Cannot create file: " + err.Error())
			go onLog(&Log{
				date:    time.Now(),
				service: "LNBank",
				logType: FATAL,
				desc:    err.Error(),
			})
			return err
		}
		defer final.Close()
		fcloser, err := f.Open()
		if err != nil {
			err := errors.New("Cannot open embedded file: " + err.Error())
			go onLog(&Log{
				date:    time.Now(),
				service: "LNBank",
				logType: FATAL,
				desc:    err.Error(),
			})
			return err
		}
		_, err = io.Copy(final, fcloser)
		if err != nil {
			err := errors.New("Cannot copy embedded file to disk: " + err.Error())
			go onLog(&Log{
				date:    time.Now(),
				service: "LNBank",
				logType: FATAL,
				desc:    err.Error(),
			})
			return err
		}
		if err := final.Chmod(f.Mode()); err != nil {
			err := errors.New("Cannot change permissions on embeded file: " + err.Error())
			go onLog(&Log{
				date:    time.Now(),
				service: "LNBank",
				logType: FATAL,
				desc:    err.Error(),
			})
			return err
		}
		final.Close()
		fcloser.Close()

		go onLog(&Log{
			date:    time.Now(),
			service: "LNBank",
			logType: INFO,
			desc:    final.Name() + " installed.",
		})
	}
	return nil
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
		insertStmt, err = DB.PrepareContext(ServicesContext,
			"INSERT INTO log (timestamp, type_id, desc, service) VALUES (?, ?, ?, ?)",
		)
		if err != nil {
			insertStmt = nil
			return append(errs, err), true
		}
		// TODO defer stmt.Close()
	}

	_, err := insertStmt.ExecContext(ServicesContext, log.date.Unix(), log.logType, log.desc, log.service)
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

	result, err := DB.QueryContext(ServicesContext, conditions.String(), params...)
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

// Given a prepared command, execute it and scan its output calling onLog and onReady
// functions of the service interface. It exit when the command ends which should be
// when the context is cancelled.
func ScanCommand(ctx context.Context, service Service, cmd *exec.Cmd) *Log {
	onLog := service.onLogHook()
	onReady := service.onReadyHook()
	name := service.name()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		go onLog(service.fmtLog(FATAL, "error preparing "+name+" execution: "+err.Error()))
	}

	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()
	if err != nil {
		go onLog(service.fmtLog(FATAL, "error executing "+name+": "+err.Error()))
	}

	scanComming := make(chan bool, 1000)
	go func() {
		for scanner.Scan() {
			scanComming <- true
		}
		scanComming <- false
		// close(scanComming)
	}()
	log := service.fmtLog(INFO, "exit")
scanLoop:
	for {
		select {
		case <-ctx.Done():
			// send the kill signal and lets hope the scanner will end with some useful logs
			if err := cmd.Process.Kill(); err != nil {
				log = service.fmtLog(FATAL, "cannot kill "+name+" process: "+err.Error())
				go onLog(log)
			}
		case goon := <-scanComming:
			if !goon {
				defer close(scanComming)
				break scanLoop
			}
			text := scanner.Text()
			l, err := service.parseLogEntry(text)
			if err != nil {
				go onLog(service.fmtLog(WARNING, fmt.Sprintf("non-conventional "+name+" log format %v: %s", err, text)))
			} else {
				go onLog(&l)
				if service.isReady(l.desc) {
					go onReady()
				}
			}
		}
	}
	if scanerr := scanner.Err(); scanerr != nil {
		log = service.fmtLog(FATAL, "error receiving stdout from "+name+": %v"+scanerr.Error())
	}
	if err := cmd.Wait(); err != nil {
		if err.Error() != "signal: killed" {
			log = service.fmtLog(FATAL, err.Error())
		}
	}
	go onLog(log)
	return log
}

// Install an embeded executable, using exePath to determine if it is already
// installed and log out the result to onLog
func InstallExe(embededZip []byte, exePath string, onLog func(*Log)) error {
	_, err := os.Stat(exePath)
	if err != nil {
		rd := bytes.NewReader(embededZip)
		if err := UnzipReader(rd, int64(len(embededZip)), ServiceRootDir, onLog); err != nil {
			return errors.New("Cannot unzip embeded zip: " + err.Error())
		}
		// gc the memory
		embededLnd = nil
		return errors.New("installed")
	}
	return nil
}

const LogTable = `
CREATE TABLE IF NOT EXISTS log (
    timestamp INTEGER NOT NULL ,
    type_id TINYINT NOT NULL,
    service VARCHAR(8) NOT NULL COLLATE NOCASE,
    desc TEXT NOT NULL COLLATE NOCASE
);
create index log_idx on log (timestamp, type_id, service COLLATE NOCASE);

INSERT INTO log (timestamp, type_id, desc, service)
  VALUES (strftime('%s'), 0, 'Creation of LNBank', 'LNBank');
`
