package main

import (
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
	}
	ts.start(onReady, onFatal, onLog)
	// wait for goroutines to finish
	time.Sleep(time.Second)
}
