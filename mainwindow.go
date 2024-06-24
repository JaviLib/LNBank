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

	left := []fyne.CanvasObject{
		widget.NewLabelWithStyle("SERVICES", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		tor_widgets(torIsReady),
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
	torversion := binding.NewString()
	_ = torversion.Set("Tor")
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
		// this will make stop all child services
		cancel()
	}

	go service.start(ctx, onReady, onStop, onLog)

	return torlabel
}
