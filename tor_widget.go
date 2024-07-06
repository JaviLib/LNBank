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

func tor_widgets(torIsReady chan<- context.Context, logs chan<- *Log) fyne.CanvasObject {
	card := widget.NewCard("🔴 Tor", "", nil)
	service := TorService{}

	var runtor func()
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
		card.SetTitle("✅ Tor")
		card.SetContent(container.New(layout.NewGridLayoutWithColumns(2),
			widget.NewButtonWithIcon("stop", theme.MediaStopIcon(), cancel),
			settings))
		mw_mutex.Unlock()
		go func() {
			isRunning = true
			clock := time.Now()
			for isRunning {
				// TODO check real tor version
				mw_mutex.Lock()
				card.SetSubTitle("v0.4.8.12    🕓 " + time.Since(clock).Round(time.Second).String())
				mw_mutex.Unlock()
				time.Sleep(time.Second)
			}
		}()
		// pass the context to the main window so the rest of the services can run
		logs <- service.fmtLog(INFO, "Tor is ready to accept connections")
		torIsReady <- ctx
	}
	onStop := func(l *Log) {
		logs <- service.fmtLog(INFO, "Tor is stopped, it won't accept connections")
		mw_mutex.Lock()
		card.SetTitle("🔴 Tor")
		card.SetSubTitle("stopped")
		card.SetContent(container.New(layout.NewGridLayoutWithColumns(2),
			widget.NewButtonWithIcon("start", theme.MediaPlayIcon(), runtor), settings))
		mw_mutex.Unlock()
		isRunning = false
		// this will make stop all child services
		cancel()
	}

	runtor = func() {
		ctx, cancel = context.WithCancel(ServicesContext)
		mw_mutex.Lock()
		card.SetTitle("⏳ Tor")
		card.SetSubTitle("starting...")
		card.SetContent(container.New(layout.NewGridLayoutWithColumns(1),
			widget.NewButtonWithIcon("cancel", theme.CancelIcon(), cancel),
		))
		mw_mutex.Unlock()
		// card.SetContent(widget.NewProgressBarInfinite())
		go service.start(ctx, onReady, onStop, onLog)
	}
	runtor()

	return card
}
