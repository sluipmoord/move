package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	workInterval  time.Duration
	breakDuration time.Duration
	verbose       bool
)

type MoveReminder struct {
	app         fyne.App
	window      fyne.Window
	breakWindow fyne.Window
	timerLabel  *widget.Label
	message     *widget.Label
	closeButton *widget.Button
	ticker      *time.Ticker
	breakEnd    time.Time
	workEnd     time.Time
	workTicker  *time.Ticker
	timerActive bool
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

	// Create a new window for each break
	mr.breakWindow = mr.app.NewWindow("Move Break")

	// Prevent window from being closed by user during break
	mr.breakWindow.SetCloseIntercept(func() {
		remaining := time.Until(mr.breakEnd)
		if remaining > 0 {
			// Block window closing during break timer
			slog.Info("Break window close blocked - break is mandatory (use Cmd+Q to quit app)")
			// Update message to show options
			fyne.Do(func() {
				mr.message.SetText("Break is mandatory and cannot be closed early.\n\nOptions:\n• Press 'S' to skip break\n• Press Cmd+Q to quit application")
			})
			return
		}
		// Allow closing when break is complete
		mr.closeBreakWindow()
	})

	title := widget.NewLabel("Time to Move!")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	mr.message = widget.NewLabel("Stand up, stretch, and move around.\nTake a break from your computer!")
	mr.message.Alignment = fyne.TextAlignCenter
	mr.message.Wrapping = fyne.TextWrapWord

	mr.timerLabel = widget.NewLabel("")
	mr.timerLabel.Alignment = fyne.TextAlignCenter
	mr.timerLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Only show close button, initially disabled
	mr.closeButton = widget.NewButton("Close Break", func() {
		mr.closeBreakWindow()
	})
	mr.closeButton.Importance = widget.HighImportance
	mr.closeButton.Disable() // Disabled until break is complete

	// Add keyboard shortcut handler for skip (S key)
	mr.breakWindow.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyS {
			slog.Info("Break skipped via keyboard shortcut (S key)")
			mr.skipBreak()
		}
	})

	content := container.NewVBox(
		widget.NewSeparator(),
		title,
		widget.NewSeparator(),
		mr.message,
		widget.NewSeparator(),
		mr.timerLabel,
		widget.NewSeparator(),
		mr.closeButton,
		widget.NewSeparator(),
		widget.NewLabel("Press 'S' to skip • Cmd+Q to quit app"),
	)

	mr.breakWindow.SetContent(container.NewCenter(content))
	mr.breakWindow.SetFullScreen(true)
	mr.breakWindow.Show()
	mr.breakWindow.RequestFocus()

	mr.ticker = time.NewTicker(time.Second)
	mr.startTimer()
}

func (mr *MoveReminder) startTimer() {
	mr.timerActive = true
	mr.updateTimer()
}

func (mr *MoveReminder) updateTimer() {
	if !mr.timerActive {
		return // Timer has been stopped
	}

	remaining := time.Until(mr.breakEnd)
	if remaining <= 0 {
		fyne.Do(func() {
			mr.message.SetText("Break time is complete!\nYou can now close this window and return to work.")
			mr.timerLabel.SetText("00:00")
			mr.closeButton.SetText("Return to Work")
			mr.closeButton.Enable() // Enable the button only when break is complete
		})
		return
	}

	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60
	fyne.Do(func() {
		mr.timerLabel.SetText(fmt.Sprintf("Time remaining: %02d:%02d", minutes, seconds))
	})

	if mr.timerActive {
		time.AfterFunc(time.Second, mr.updateTimer)
	}
}

func (mr *MoveReminder) skipBreak() {
	slog.Info("Skipping break via keyboard shortcut")
	mr.timerActive = false
	if mr.ticker != nil {
		mr.ticker.Stop()
	}

	// Close the break window completely
	if mr.breakWindow != nil {
		mr.breakWindow.Close()
	}

	slog.Info("Break skipped, scheduling next work interval")
	os.Stdout.Sync()
	mr.scheduleNext()
}

func (mr *MoveReminder) closeBreakWindow() {
	slog.Info("Closing break window")

	// Stop all timers immediately
	mr.timerActive = false
	if mr.ticker != nil {
		mr.ticker.Stop()
	}

	// Disable the button to prevent double-clicks
	mr.closeButton.Disable()

	// Close the break window completely instead of hiding
	if mr.breakWindow != nil {
		mr.breakWindow.Close()
	}

	slog.Info("Break window closed, scheduling next work interval")
	os.Stdout.Sync()
	mr.scheduleNext()
}

func (mr *MoveReminder) startWorkTimer() {
	mr.workEnd = time.Now().Add(workInterval)

	// Use different intervals based on verbose flag
	interval := 10 * time.Second
	if verbose {
		interval = 1 * time.Second
	}

	mr.workTicker = time.NewTicker(interval)

	go func() {
		for range mr.workTicker.C {
			remaining := time.Until(mr.workEnd)
			if remaining <= 0 {
				mr.workTicker.Stop()
				slog.Info("Work interval completed - break time!")
				os.Stdout.Sync() // Force flush
				return
			}

			minutes := int(remaining.Minutes())
			seconds := int(remaining.Seconds()) % 60
			slog.Info("Work time remaining", "time", fmt.Sprintf("%02d:%02d", minutes, seconds))
			os.Stdout.Sync() // Force flush after each log
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
	os.Stdout.Sync()
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
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging every 1 seconds")
	flag.Parse()

	workInterval = *workFlag
	breakDuration = *breakFlag
	verbose = *verboseFlag

	// Force logs to be visible by flushing stdout
	slog.Info("Move reminder configured", "work_interval", workInterval, "break_duration", breakDuration)
	os.Stdout.Sync()

	if verbose {
		slog.Info("Verbose logging enabled - will log every 1 seconds")
		os.Stdout.Sync()
	}

	reminder := NewMoveReminder()
	reminder.start()
}
