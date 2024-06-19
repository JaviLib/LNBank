package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTorStart(t *testing.T) {
	ts := TorService{}
	onReady := func() {
	}
	onFatal := func(log *Log) {
		t.Fatal(log.desc)
	}
	onLog := func(log *Log) {
		assert.NotNil(t, log)
		assert.Equal(t, LogType(INFO), log.logType)
		if testing.Verbose() {
			fmt.Println(log)
		}
	}
	ts.start(onReady, onFatal, onLog)
	// wait for goroutines to finish
	time.Sleep(time.Second)
}
