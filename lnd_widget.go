package main

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func lnd_widgets(torIsReady <-chan context.Context, lndIsReady chan<- context.Context) fyne.CanvasObject {
	card := widget.NewCard("🔴 lnd", "", nil)
	go func() {
		torctx := <-torIsReady
		ctx, cancel := context.WithCancel(torctx)
		card.SetTitle("⏳ lnd")
		// call all dependencies once started:
		for range cap(lndIsReady) {
			lndIsReady <- ctx
		}
		// TODO implement LND GUI
		fmt.Println("lnd finished")
		cancel()
	}()
	return card
}
