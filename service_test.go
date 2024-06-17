package main

import (
	"errors"
	"strings"
	"testing"
	"time"
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
