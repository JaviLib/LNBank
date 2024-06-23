package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTorStart(t *testing.T) {
	ts := TorService{}
	gotReady := false
	onReady := func() {
		gotReady = true
	}
	onStop := func(log *Log) {
		assert.NotNil(t, log, "stoppped with nil log")
		if log.logType == FATAL {
			t.Fatal(log.desc)
		}
	}
	onLog := func(log *Log) {
		assert.NotNil(t, log)
		assert.Equal(t, LogType(INFO), log.logType, "Log type is not INFO but %v", log.logType)
		if testing.Verbose() {
			fmt.Println(log)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	go ts.start(ctx, onReady, onStop, onLog)
	time.Sleep(time.Second * 5)
	cancel()
	assert.True(t, gotReady, "Tor never got ready")
}
