package main

import (
	"flag"
	"fmt"
	"log/slog"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	workInterval  time.Duration
	breakDuration time.Duration
)

type MoveReminder struct {
	app         fyne.App
	window      fyne.Window
	timerLabel  *widget.Label
	message     *widget.Label
	closeButton *widget.Button
	ticker      *time.Ticker
	breakEnd    time.Time
	workEnd     time.Time
	workTicker  *time.Ticker
}

func NewMoveReminder() *MoveReminder {
	myApp := app.New()
	myApp.SetIcon(nil)

	window := myApp.NewWindow("Move Reminder")
	window.CenterOnScreen()

	return &MoveReminder{
		app:    myApp,
		window: window,
	}
}

func (mr *MoveReminder) showBreakWindow() {
	mr.breakEnd = time.Now().Add(breakDuration)

	title := widget.NewLabel("Time to Move!")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	mr.message = widget.NewLabel("Stand up, stretch, and move around.\nTake a break from your computer!")
	mr.message.Alignment = fyne.TextAlignCenter
	mr.message.Wrapping = fyne.TextWrapWord

	mr.timerLabel = widget.NewLabel("")
	mr.timerLabel.Alignment = fyne.TextAlignCenter
	mr.timerLabel.TextStyle = fyne.TextStyle{Bold: true}

	mr.closeButton = widget.NewButton("Close Break", func() {
		mr.closeBreakWindow()
	})
	mr.closeButton.Importance = widget.HighImportance

	skipButton := widget.NewButton("Skip Break", func() {
		if mr.ticker != nil {
			mr.ticker.Stop()
		}
		mr.window.SetFullScreen(false)
		mr.window.Hide()
		mr.scheduleNext()
	})

	buttonContainer := container.NewHBox(skipButton, mr.closeButton)

	content := container.NewVBox(
		widget.NewSeparator(),
		title,
		widget.NewSeparator(),
		mr.message,
		widget.NewSeparator(),
		mr.timerLabel,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
	)

	mr.window.SetContent(container.NewCenter(content))
	mr.window.SetFullScreen(true)
	mr.window.Show()
	mr.window.RequestFocus()

	mr.ticker = time.NewTicker(time.Second)
	mr.startTimer()
}

func (mr *MoveReminder) startTimer() {
	mr.updateTimer()
}

func (mr *MoveReminder) updateTimer() {
	remaining := time.Until(mr.breakEnd)
	if remaining <= 0 {
		fyne.Do(func() {
			mr.message.SetText("Break time is complete!\nYou can now close this window and return to work.")
			mr.timerLabel.SetText("00:00")
			mr.closeButton.SetText("Return to Work")
		})
		return
	}

	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60
	fyne.Do(func() {
		mr.timerLabel.SetText(fmt.Sprintf("Time remaining: %02d:%02d", minutes, seconds))
	})

	time.AfterFunc(time.Second, mr.updateTimer)
}

func (mr *MoveReminder) closeBreakWindow() {
	if mr.ticker != nil {
		mr.ticker.Stop()
	}
	mr.window.SetFullScreen(false)
	mr.window.Hide()
	mr.scheduleNext()
}

func (mr *MoveReminder) startWorkTimer() {
	mr.workEnd = time.Now().Add(workInterval)
	mr.workTicker = time.NewTicker(30 * time.Second) // Log every 30 seconds

	go func() {
		for range mr.workTicker.C {
			remaining := time.Until(mr.workEnd)
			if remaining <= 0 {
				mr.workTicker.Stop()
				return
			}

			minutes := int(remaining.Minutes())
			seconds := int(remaining.Seconds()) % 60
			slog.Info("Work time remaining", "time", fmt.Sprintf("%02d:%02d", minutes, seconds))
		}
	}()
}

func (mr *MoveReminder) stopWorkTimer() {
	if mr.workTicker != nil {
		mr.workTicker.Stop()
	}
}

func (mr *MoveReminder) scheduleNext() {
	slog.Info("Starting work interval", "duration", workInterval)
	mr.startWorkTimer()

	time.AfterFunc(workInterval, func() {
		mr.stopWorkTimer()
		fyne.Do(func() {
			mr.showBreakWindow()
		})
	})
}

func (mr *MoveReminder) start() {
	slog.Info("Move reminder started", "work_interval", workInterval, "break_duration", breakDuration)
	mr.scheduleNext()
	mr.app.Run()
}

func main() {
	workFlag := flag.Duration("work", 25*time.Minute, "Work interval duration (e.g., 25m, 10s)")
	breakFlag := flag.Duration("break", 5*time.Minute, "Break duration (e.g., 5m, 10s)")
	flag.Parse()

	workInterval = *workFlag
	breakDuration = *breakFlag

	slog.Info("Move reminder configured", "work_interval", workInterval, "break_duration", breakDuration)

	reminder := NewMoveReminder()
	reminder.start()
}
