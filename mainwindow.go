package main

import (
	"context"
	"errors"
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
	// this channel should have a length equal to the number of services that
	// depend on lnd, and lnd should call all of them checking its capacity:
	lndIsReady := make(chan context.Context, 2)
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
	done := make(chan bool)
	go func() {
		for {
			select {
			case l := <-logs:
				bold := false
				if l.logType == ERROR || l.logType == FATAL {
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
				// TODO remove once the dependencies are installed
			case <-lndIsReady:
				logwidget.Segments = append(logwidget.Segments, &widget.TextSegment{Text: "lnd is ready"})
				logwidget.Refresh()
				logscroll.ScrollToBottom()
			case <-done:
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
			desc:    "closing all services",
		}
		// wait for the last logs to happen
		ServicesCancelFunc()
		done <- true
		w.Close()
	})
	w.ShowAndRun()
}
