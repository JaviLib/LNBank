package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func CompareErrorSlices(expected []error, actual []error) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i := range expected {
		if !strings.Contains(actual[i].Error(), expected[i].Error()) {
			return false
		}
	}

	return true
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		log         *Log
		expectedErr []error
	}{
		{
			name: "Valid log entry",
			log: &Log{
				date:    time.Now(),
				logType: NORMAL,
				desc:    "Test description",
				service: "TEST_SERVICE",
			},
			expectedErr: nil, // <-- Rewritten portion
		},
		{
			name: "Log with incorrect time",
			log: &Log{
				date:    time.Unix(0, 0),
				logType: NORMAL,
				desc:    "Test description",
				service: "TEST_SERVICE",
			},
			expectedErr: []error{errors.New("incorrect time '0'")}, // <-- Rewritten portion
		},
		{
			name: "Log with future date",
			log: &Log{
				date:    time.Date(2057, 12, 31, 23, 59, 59, 0, time.UTC),
				logType: NORMAL,
				desc:    "Test description",
				service: "TEST_SERVICE",
			},
			expectedErr: []error{errors.New("has a time in the future")},
		},
		{
			name: "Log with invalid log type",
			log: &Log{
				date:    time.Now(),
				logType: 99, // Invalid value
				desc:    "Test description",
				service: "TEST_SERVICE",
			},
			expectedErr: []error{errors.New("incorrect log type 99")},
		},
		{
			name: "Log with empty description",
			log: &Log{
				date:    time.Now(),
				logType: NORMAL,
				desc:    "",
				service: "TEST_SERVICE",
			},
			expectedErr: []error{errors.New("no description")},
		},
		{
			name: "Log with empty service",
			log: &Log{
				date:    time.Now(),
				logType: NORMAL,
				desc:    "Test description",
				service: "",
			},
			expectedErr: []error{errors.New("no service provided")},
		},
		{
			name:        "Nil log",
			log:         nil,
			expectedErr: []error{errors.New("trying to log a null log entry")},
		},
		{
			name: "test several errors",
			log: &Log{
				date:    time.Now(),
				logType: 99,
				desc:    "Test description",
				service: "",
			},
			expectedErr: []error{
				errors.New("incorrect log type 99"),
				errors.New("no service provided"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.log.Validate()
			if !CompareErrorSlices(err, tt.expectedErr) { // <-- Rewritten portion
				t.Errorf("Wanted: %v\nGot: %v\n", tt.expectedErr, err)
			}
		})
	}
}

func TestLogToDb(t *testing.T) {
	db := Logdb

	log := &Log{
		date:    time.Now(),
		logType: INFO,
		desc:    "Test log message",
		service: "test_service",
	}
	errs, fatal := LogToDb(log)
	assert.False(t, fatal, "%v", errs)
	assert.Empty(t, errs)
	var id int
	err := db.QueryRow(
		"SELECT timestamp FROM log WHERE desc=? AND service=?",
		log.desc, log.service,
	).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotZero(t, id)

	// test with some errors but not fatal
	now := time.Now()
	log = &Log{
		date:    now,
		logType: 99, // incorrect
		desc:    "Test log message with incorrect logtype",
		service: "test_service",
	}
	errs, fatal = LogToDb(log)
	assert.False(t, fatal, "%v", errs)
	assert.NotEmpty(t, errs)

	// check that logtype was corrected
	var type_id LogType
	err = db.QueryRow(
		"SELECT type_id FROM log WHERE desc=? AND service=? AND timestamp=?",
		log.desc, log.service, now.Unix(),
	).Scan(&type_id)
	if err != nil {
		t.Fatal(err)
	}
	assert.Zero(t, type_id)

	// Logs with errors are also logged to the db
	var desc string
	err = db.QueryRow(
		"SELECT desc FROM log WHERE timestamp=? AND desc LIKE '%incorrect log type%' ORDER BY timestamp DESC",
		time.Now().Unix(),
	).Scan(&desc)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, desc, "Service test_service provided an incorrect log: [incorrect log type 99]")
}

func TestQueryLog(t *testing.T) {
	servs := []string{"service1", "service2", "service3"}
	for i := range 1000 {
		errs, fatal := LogToDb(&Log{
			// insert logs in the past
			date:    time.Now().Add(-time.Hour * time.Duration(i)),
			logType: LogType(rand.Intn(DEBUG)),
			service: servs[rand.Intn(2)],
			desc:    fmt.Sprintf("random description %% %v", rand.Intn(5)),
		})
		assert.Nil(t, errs)
		assert.False(t, fatal)
	}

	// check that we can get the first log

	query, err := QueryLog(time.Hour*5, nil, nil, "", 0)
	if err != nil {
		t.Fatal(err)
	} else {
		first_row, err := query.next()
		assert.NoError(t, err)
		fmt.Println(first_row)
		query.close()
	}
	// check for logtypes
	query, err = QueryLog(time.Hour*5, []LogType{WARNING}, nil, "", 0)
	if err != nil {
		t.Fatal(err)
	} else {
		first_row, err := query.next()
		assert.NoError(t, err)
		fmt.Println(first_row)
		query.close()
	}
	// check for services
	query, err = QueryLog(time.Hour*5, []LogType{ERROR}, []string{servs[0], servs[1]}, "", 0)
	if err != nil {
		t.Fatal(err)
	} else {
		first_row, err := query.next()
		assert.NoError(t, err)
		fmt.Println(first_row)
		query.close()
	}
	// check for description
	query, err = QueryLog(time.Hour*5, nil, nil, "description % 4", 0)
	if err != nil {
		t.Fatal(err)
	} else {
		first_row, err := query.next()
		assert.NoError(t, err)
		fmt.Println(first_row)
		query.close()
	}
	// check getting 1000
	query, err = QueryLog(time.Hour*1000, nil, nil, "", 1000)
	if err != nil {
		t.Fatal(err)
	} else {
		logs, err := query.getn(1000)
		assert.NoError(t, err)
		assert.Len(t, logs, 1000)
		query.close()
	}
}
