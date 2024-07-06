package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func lnd_widgets(torIsReady <-chan context.Context, lndIsReady chan<- context.Context, logs chan<- *Log) fyne.CanvasObject {
	card := widget.NewCard("🔴 lnd", "", nil)
	service := LndService{}

	var runlnd func()
	var ctx context.Context
	var cancel context.CancelFunc

	settings := widget.NewButtonWithIcon("config", theme.SettingsIcon(), func() {
		fmt.Println("Settings tor")
	})
	isRunning := false

	onLog := func(l *Log) {
		logs <- l
	}
	onReady := func() {
		mw_mutex.Lock()
		card.SetTitle("✅ lnd")
		card.SetContent(container.New(layout.NewGridLayoutWithColumns(2),
			widget.NewButtonWithIcon("stop", theme.MediaStopIcon(), cancel),
			settings))
		mw_mutex.Unlock()
		go func() {
			isRunning = true
			clock := time.Now()
			for isRunning {
				// TODO check real version
				mw_mutex.Lock()
				card.SetSubTitle("v0.18.1b    🕓 " + time.Since(clock).Round(time.Second).String())
				mw_mutex.Unlock()
				time.Sleep(time.Second)
			}
		}()
		// pass the context to the main window so the rest of the services can run
		logs <- service.fmtLog(INFO, "lnd is ready to accept connections")
		// call all dependencies once started:
		for range cap(lndIsReady) {
			lndIsReady <- ctx
		}
	}
	onStop := func(l *Log) {
		logs <- service.fmtLog(INFO, "lnd is stopped, it won't accept connections")
		mw_mutex.Lock()
		card.SetTitle("🔴 lnd")
		card.SetSubTitle("stopped")
		card.SetContent(container.New(layout.NewGridLayoutWithColumns(2),
			widget.NewButtonWithIcon("start", theme.MediaPlayIcon(), runlnd), settings))
		mw_mutex.Unlock()
		isRunning = false
		// this will make stop all child services
		cancel()
	}

	runlnd = func() {
		torctx := <-torIsReady
		ctx, cancel = context.WithCancel(torctx)
		mw_mutex.Lock()
		card.SetTitle("⏳ lnd")
		card.SetSubTitle("starting...")
		card.SetContent(container.New(layout.NewGridLayoutWithColumns(1),
			widget.NewButtonWithIcon("cancel", theme.CancelIcon(), cancel),
		))
		mw_mutex.Unlock()
		// card.SetContent(widget.NewProgressBarInfinite())
		go service.start(ctx, onReady, onStop, onLog)
	}
	go runlnd()

	// go func() {
	// 	torctx := <-torIsReady
	// 	ctx, cancel := context.WithCancel(torctx)
	// 	card.SetTitle("⏳ lnd")
	// 	// call all dependencies once started:
	// 	for range cap(lndIsReady) {
	// 		lndIsReady <- ctx
	// 	}
	// 	// TODO implement LND GUI
	// 	fmt.Println("lnd finished")
	// 	cancel()
	// }()
	return card
}
