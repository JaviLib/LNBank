package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main_window(w fyne.Window) {
	// When Tor is ready, pass the context to the next service (LND)
	torIsReady := make(chan context.Context)
	lndIsReady := make(chan context.Context)

	left := []fyne.CanvasObject{
		// widget.NewLabelWithStyle("SERVICES", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		tor_widgets(torIsReady),
		lnd_widgets(torIsReady, lndIsReady),
	}

	content := container.New(
		// General layout
		layout.NewGridLayout(2),
		// Left layout
		container.New(
			layout.NewGridLayoutWithRows(2),
			left...,
		),
		// Right layout
		widget.NewLabel("Right"),
	)

	w.SetContent(content)
}

func tor_widgets(torIsReady chan<- context.Context) fyne.CanvasObject {
	card := widget.NewCard("🔴 Tor", "", nil)
	service := TorService{}

	ctx, cancel := context.WithCancel(ServicesContext)

	onLog := func(l *Log) {
		fmt.Println(l)
	}
	isRunning := false
	onReady := func() {
		card.SetTitle("✅ Tor")
		card.SetContent(nil)
		go func() {
			isRunning = true
			clock := time.Now()
			for isRunning {
				card.SetSubTitle("v0.4.8.12    🕓 " + time.Since(clock).Round(time.Second).String())
				<-time.After(time.Second)
			}
		}()
		// pass the context to the main window so the rest of the services can run
		torIsReady <- ctx
	}
	onStop := func(l *Log) {
		fmt.Println(l)
		card.SetTitle("🔴 Tor")
		isRunning = false
		// this will make stop all child services
		cancel()
	}

	card.SetTitle("⏳ Tor")
	card.SetSubTitle("starting...")
	card.SetContent(widget.NewProgressBarInfinite())
	go service.start(ctx, onReady, onStop, onLog)

	return card
}

func lnd_widgets(torIsReady <-chan context.Context, lndIsReady chan<- context.Context) fyne.CanvasObject {
	card := widget.NewCard("🔴 lnd", "", nil)
	go func() {
		torctx := <-torIsReady
		ctx, cancel := context.WithCancel(torctx)
		card.SetTitle("⏳ lnd")
		lndIsReady <- ctx
		// TODO implement LND GUI
		cancel()
	}()
	return card
}
