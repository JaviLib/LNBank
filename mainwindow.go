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
	Services["Tor"] = TorService{}

	left := []fyne.CanvasObject{
		widget.NewLabel("Services:"),
		tor_widgets(),
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

func tor_widgets() fyne.CanvasObject {
	torversion := binding.NewString()
	torlabel := widget.NewLabel("Tor")
	torversion.Set("Tor")
	torlabel.Bind(torversion)

	service := TorService{}

	ctx, cancel := context.WithCancel(ServicesContext)

	onLog := func(l *Log) {
		fmt.Println(l)
	}
	onStop := func(l *Log) {
		fmt.Println(l)
		cancel()
	}
	onReady := func() {
		torversion.Set("Tor READY")
		fmt.Println("Ready")
	}

	go service.start(ctx, onReady, onStop, onLog)

	return torlabel
}
