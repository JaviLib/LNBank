package main

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main_window(w fyne.Window) {
	// When Tor is ready, pass the context to the next service (LND)
	torIsReady := make(chan context.Context)
	lndIsReady := make(chan context.Context)

	left := []fyne.CanvasObject{
		widget.NewLabelWithStyle("SERVICES", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		tor_widgets(torIsReady),
		lnd_widgets(torIsReady, lndIsReady),
	}

	content := container.New(
		// General layout
		layout.NewGridLayout(2),
		// Left layout
		container.New(
			layout.NewGridLayoutWithRows(3),
			left...,
		),
		// Right layout
		widget.NewLabel("Right"),
	)

	w.SetContent(content)
}

func tor_widgets(torIsReady chan<- context.Context) fyne.CanvasObject {
	torversion := binding.NewString()
	_ = torversion.Set("Tor STOPPED")
	torlabel := widget.NewLabelWithData(torversion)

	service := TorService{}

	ctx, cancel := context.WithCancel(ServicesContext)

	onLog := func(l *Log) {
		fmt.Println(l)
	}
	onReady := func() {
		_ = torversion.Set("Tor READY")
		// pass the context to the main window so the rest of the services can run
		torIsReady <- ctx
	}
	onStop := func(l *Log) {
		fmt.Println(l)
		_ = torversion.Set("Tor STOPPED")
		// this will make stop all child services
		cancel()
	}

	_ = torversion.Set("Tor STARTING")
	go service.start(ctx, onReady, onStop, onLog)

	return torlabel
}

func lnd_widgets(torIsReady <-chan context.Context, lndIsReady chan<- context.Context) fyne.CanvasObject {
	lndversion := binding.NewString()
	_ = lndversion.Set("LND STOPPED")
	lndlabel := widget.NewLabelWithData(lndversion)
	go func() {
		torctx := <-torIsReady
		ctx, cancel := context.WithCancel(torctx)
		_ = lndversion.Set("LND STARTING")
		lndIsReady <- ctx
		// TODO implement LND GUI
		cancel()
	}()
	return lndlabel
}
