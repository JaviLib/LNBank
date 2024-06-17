package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main_window(w fyne.Window) {
	Services["Tor"] = TorService{}

	content := container.New(
		// General layout
		layout.NewGridLayout(2),
		// Left layout
		container.New(
			layout.NewGridLayoutWithRows(2),
			widget.NewLabel("Services:"),
			widget.NewLabel("Left2"),
		),
		// Right layout
		widget.NewLabel("Right"),
	)

	w.SetContent(content)
}
