package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main_window(w fyne.Window) {
	// When Tor is ready, pass the context to the next service (LND)
	torIsReady := make(chan context.Context)
	// lndIsReady := make(chan context.Context)
	logs := make(chan *Log)

	left := container.New(
		layout.NewVBoxLayout(),
		tor_widgets(torIsReady, logs),
		// lnd_widgets(torIsReady, lndIsReady),
	)

	logwidget := widget.NewRichTextWithText("Session entries:\n")
	logwidget.Wrapping = fyne.TextWrapWord

	logscroll := container.NewScroll(logwidget)
	logscroll.SetMinSize(fyne.Size{Width: 640, Height: 480})

	content := container.NewBorder(nil, nil, left, nil, logscroll)

	go func() {
		for {
			select {
			case l := <-logs:
				bold := false
				if l.logType == ERROR || l.logType == FATAL || l.logType == WARNING {
					bold = true
				}
				segment := widget.TextSegment{
					Text: l.String(),
					Style: widget.RichTextStyle{
						TextStyle: fyne.TextStyle{
							Bold: bold,
						},
					},
				}
				logwidget.Segments = append(logwidget.Segments, &segment)
				logscroll.ScrollToBottom()
				logwidget.Refresh()
				errs, fatal := LogToDb(l)
				if errs != nil && fatal {
					dialog.ShowError(errors.Join(errs...), w)
				}
			// case <-lndIsReady:
			// logwidget.Segments = append(logwidget.Segments, &widget.TextSegment{Text: "lnd is ready"})
			case <-ServicesContext.Done():
				return
			}
		}
	}()

	w.SetContent(content)
}

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
		fmt.Println(l)
		logs <- l
	}
	onReady := func() {
		card.SetTitle("✅ Tor")
		card.SetContent(container.New(layout.NewGridLayoutWithColumns(2),
			widget.NewButtonWithIcon("stop", theme.MediaStopIcon(), cancel),
			settings))
		go func() {
			isRunning = true
			clock := time.Now()
			for isRunning {
				card.SetSubTitle("v0.4.8.12    🕓 " + time.Since(clock).Round(time.Second).String())
				time.Sleep(time.Second)
			}
		}()
		// pass the context to the main window so the rest of the services can run
		torIsReady <- ctx
	}
	onStop := func(l *Log) {
		fmt.Println(l)
		card.SetTitle("🔴 Tor")
		card.SetSubTitle("stopped")
		card.SetContent(container.New(layout.NewGridLayoutWithColumns(2),
			widget.NewButtonWithIcon("start", theme.MediaPlayIcon(), runtor), settings))
		isRunning = false
		// this will make stop all child services
		cancel()
	}

	runtor = func() {
		ctx, cancel = context.WithCancel(ServicesContext)
		card.SetTitle("⏳ Tor")
		card.SetSubTitle("starting...")
		card.SetContent(container.New(layout.NewGridLayoutWithColumns(1),
			widget.NewButtonWithIcon("cancel", theme.CancelIcon(), cancel),
		))
		// card.SetContent(widget.NewProgressBarInfinite())
		go service.start(ctx, onReady, onStop, onLog)
	}
	runtor()

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
