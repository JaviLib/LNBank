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
	defer close(torIsReady)
	lndIsReady := make(chan context.Context)
	defer close(lndIsReady)
	logs := make(chan *Log, 1000)
	defer close(logs)

	left := container.New(
		layout.NewVBoxLayout(),
		tor_widgets(torIsReady, logs),
		lnd_widgets(torIsReady, lndIsReady),
	)

	logwidget := widget.NewRichTextWithText("Session entries:\n")
	logwidget.Wrapping = fyne.TextWrapWord

	logscroll := container.NewScroll(logwidget)
	logscroll.SetMinSize(fyne.Size{Width: 640, Height: 480})

	filterchecks := widget.NewCheckGroup(
		[]string{"Tor", "lnd", "Neutrino", "HTCL", "gossip", "wallet", "LNBits", "lndG", "clearnet"},
		func([]string) {},
	)
	filterchecks.Horizontal = true
	filterchecks.SetSelected([]string{"Tor", "lnd", "Neutrino", "HTCL", "gossip", "wallet", "LNBits", "lndG", "clearnet"})
	filterentry := widget.NewEntry()
	filterentry.PlaceHolder = "description filter      "
	filterentry.Scroll = container.ScrollNone
	filterentry.Wrapping = fyne.TextWrapOff

	filtererrors := widget.NewCheckGroup(
		[]string{LogType(FATAL).String(), LogType(ERROR).String(), LogType(WARNING).String(), LogType(INFO).String() + " INFO", LogType(DEBUG).String() + " DEBUG"},
		func([]string) {},
	)
	filtererrors.Horizontal = true
	filtererrors.SetSelected([]string{LogType(FATAL).String(), LogType(ERROR).String(), LogType(WARNING).String(), LogType(INFO).String() + " INFO", LogType(DEBUG).String() + " DEBUG"})

	filterchoices := widget.NewRadioGroup([]string{"Session", "All", "Month", "Week", "Day", "Hour"}, func(string) {})
	filterchoices.Horizontal = true
	filterchoices.SetSelected("Session")
	toggleonall := widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() {})
	toggleoffall := widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {})
	copylogs := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {})
	filtercheckboxes := container.NewVBox(
		container.NewHBox(toggleoffall, toggleonall, copylogs, filterchecks, filterentry),
		container.NewHBox(filterchoices, filtererrors),
	)

	right := container.NewBorder(nil, filtercheckboxes, nil, nil, logscroll)

	content := container.NewBorder(nil, nil, left, nil, right)

	sessionlogs := make([]*Log, 0)
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
				logwidget.Refresh()
				logscroll.ScrollToBottom()
				sessionlogs = append(sessionlogs, l)
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
	w.SetCloseIntercept(func() {
		// TODO in fact, hide it and minimize to systray
		logs <- &Log{
			date:    time.Now(),
			logType: WARNING,
			service: "LNBank",
			desc:    "closing all services and exiting in 2 seconds",
		}
		ServicesCancelFunc()
		time.Sleep(time.Second * 2)
		w.Close()
	})
	w.ShowAndRun()
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
