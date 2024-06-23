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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)

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
	assert.True(t, gotReady, "Tor never got ready")
}
