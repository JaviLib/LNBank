package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLndStart(t *testing.T) {
	ts := LndService{}
	ctx, cancel := context.WithTimeout(ServicesContext, time.Second*30)

	gotReady := false
	onReady := func() {
		gotReady = true
		cancel()
	}
	onStop := func(log *Log) {
		assert.NotNil(t, log, "stoppped with nil log")
		if log.logType == FATAL {
			t.Fatal(log.desc)
		}
	}
	onLog := func(log *Log) {
		assert.NotNil(t, log)
		assert.Equal(t, LogType(INFO), log.logType, "Log type is not INFO but %v: %v", log.logType, log.desc)
		if testing.Verbose() {
			fmt.Println(log)
		}
	}
	ts.start(ctx, onReady, onStop, onLog)
	time.Sleep(time.Second * 2)
	assert.True(t, gotReady, "Lnd never got ready")
}
