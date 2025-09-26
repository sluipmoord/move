package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
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
	workEnd     time.Time
	workTicker  *time.Ticker
	app         fyne.App
	breakWindow fyne.Window
	isRunning   bool
	appInit     bool
	breakDone   chan bool
}

func NewMoveReminder() *MoveReminder {
	mr := &MoveReminder{
		isRunning: true,
		appInit:   false,
	}

	return mr
}

func (mr *MoveReminder) initApp() {
	// Create the app and mark as initialized
	mr.app = app.New()
	mr.app.SetIcon(nil)
	mr.appInit = true

	// Create a persistent window to keep the app alive - must be visible for GLFW stability
	persistentWindow := mr.app.NewWindow("")
	persistentWindow.Resize(fyne.NewSize(1, 1)) // Make it as small as possible
	persistentWindow.SetCloseIntercept(func() {
		// Prevent this window from being closed - it keeps the app alive
	})
	persistentWindow.Show() // Must show for GLFW context to remain valid
}

func (mr *MoveReminder) showBreakWindow() {
	// Show system notification
	mr.showNotification()

	// Create break window on Fyne thread (app is already running)
	mr.breakDone = make(chan bool)
	mr.createBreakWindow()

	// Wait for break to complete
	<-mr.breakDone

	slog.Info("Break completed, resuming work")
	go mr.scheduleNext()
}

func (mr *MoveReminder) showNotification() {
	// Show system notification on macOS
	title := "Move Break Time!"
	message := "Stand up, stretch, and move around. Take a break from your computer!"
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`display notification "%s" with title "%s" sound name "Glass"`, message, title))
	err := cmd.Run()
	if err != nil {
		slog.Error("Failed to show notification", "error", err)
	}
}

func (mr *MoveReminder) createBreakWindow() {
	fyne.Do(func() {
		if mr.breakWindow != nil {
			mr.breakWindow.Close()
		}

		mr.breakWindow = mr.app.NewWindow("Move Break")

		// Make it fullscreen and modal
		mr.breakWindow.SetFullScreen(true)
		mr.breakWindow.CenterOnScreen()

		// Prevent closing with close button
		mr.breakWindow.SetCloseIntercept(func() {
			// Do nothing - force user to use buttons
		})

		// Create break window content
		mr.setupBreakWindowContent()

		// Show the window
		mr.breakWindow.Show()
		mr.breakWindow.RequestFocus()

		// Keep window focused
		go mr.maintainFocus()
	})
}

func (mr *MoveReminder) setupBreakWindowContent() {
	title := widget.NewLabel("🚶 Time to Move! 🚶")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	message := widget.NewLabel("Stand up, stretch, and move around.\nTake a break from your computer!")
	message.Alignment = fyne.TextAlignCenter
	message.Wrapping = fyne.TextWrapWord

	timerLabel := widget.NewLabel("")
	timerLabel.Alignment = fyne.TextAlignCenter
	timerLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Skip button
	skipButton := widget.NewButton("Skip Break (S)", func() {
		slog.Info("Break skipped by user")
		mr.closeBreakWindow()
	})
	skipButton.Importance = widget.MediumImportance

	// Return to work button (initially disabled)
	returnButton := widget.NewButton("Return to Work", func() {
		slog.Info("User clicked return to work")
		mr.closeBreakWindow()
	})
	returnButton.Importance = widget.HighImportance
	returnButton.Disable()

	// Keyboard shortcut handler
	mr.breakWindow.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyS {
			slog.Info("Break skipped via keyboard shortcut (S key)")
			mr.closeBreakWindow()
		}
	})

	content := container.NewVBox(
		widget.NewSeparator(),
		title,
		widget.NewSeparator(),
		message,
		widget.NewSeparator(),
		timerLabel,
		widget.NewSeparator(),
		container.NewHBox(skipButton, returnButton),
		widget.NewSeparator(),
		widget.NewLabel("Press 'S' to skip • Cmd+Q to quit app"),
	)

	mr.breakWindow.SetContent(container.NewCenter(content))

	// Start the break timer
	mr.startBreakTimer(timerLabel, returnButton)
}

func (mr *MoveReminder) startBreakTimer(timerLabel *widget.Label, returnButton *widget.Button) {
	endTime := time.Now().Add(breakDuration)

	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()

		for range ticker.C {
			if mr.breakWindow == nil {
				return
			}

			remaining := time.Until(endTime)
			if remaining <= 0 {
				// Break time complete
				fyne.Do(func() {
					timerLabel.SetText("Break Complete!")
					returnButton.SetText("Return to Work")
					returnButton.Enable()
				})
				return
			}

			minutes := int(remaining.Minutes())
			seconds := int(remaining.Seconds()) % 60
			fyne.Do(func() {
				timerLabel.SetText(fmt.Sprintf("Time remaining: %02d:%02d", minutes, seconds))
			})
		}
	}()
}

func (mr *MoveReminder) maintainFocus() {
	focusTicker := time.NewTicker(500 * time.Millisecond)
	defer focusTicker.Stop()

	for range focusTicker.C {
		if mr.breakWindow == nil {
			return
		}

		fyne.Do(func() {
			mr.breakWindow.RequestFocus()
		})

		// Use AppleScript to bring to front
		cmd := exec.Command("osascript", "-e", `tell application "System Events" to set frontmost of first process whose name contains "main" to true`)
		cmd.Run()
	}
}

func (mr *MoveReminder) closeBreakWindow() {
	fyne.Do(func() {
		if mr.breakWindow != nil {
			mr.breakWindow.Close()
			mr.breakWindow = nil
		}
	})

	// Signal that break is done
	if mr.breakDone != nil {
		mr.breakDone <- true
	}
}

func (mr *MoveReminder) startWorkTimer() {
	if !mr.isRunning {
		return
	}

	mr.workEnd = time.Now().Add(workInterval)

	// Use different intervals based on verbose flag
	interval := 10 * time.Second
	if verbose {
		interval = 1 * time.Second
	}

	mr.workTicker = time.NewTicker(interval)

	go func() {
		for range mr.workTicker.C {
			if !mr.isRunning {
				mr.workTicker.Stop()
				return
			}

			remaining := time.Until(mr.workEnd)
			if remaining <= 0 {
				mr.workTicker.Stop()
				slog.Info("Work interval completed - break time!")
				os.Stdout.Sync()
				mr.showBreakWindow()
				return
			}

			minutes := int(remaining.Minutes())
			seconds := int(remaining.Seconds()) % 60
			slog.Info("Work time remaining", "time", fmt.Sprintf("%02d:%02d", minutes, seconds))
			os.Stdout.Sync()
		}
	}()
}

func (mr *MoveReminder) stopWorkTimer() {
	if mr.workTicker != nil {
		mr.workTicker.Stop()
	}
	mr.isRunning = false
}

func (mr *MoveReminder) scheduleNext() {
	if !mr.isRunning {
		return
	}
	slog.Info("Starting work interval", "duration", workInterval)
	os.Stdout.Sync()
	mr.startWorkTimer()
}

func (mr *MoveReminder) start() {
	slog.Info("Move reminder started", "work_interval", workInterval, "break_duration", breakDuration)

	// Initialize Fyne app from main thread
	mr.initApp()

	// Start work timer in background
	go mr.scheduleNext()

	// Run Fyne main loop - this will handle break windows when needed
	mr.app.Run()
}

func (mr *MoveReminder) quit() {
	slog.Info("Quitting application")
	mr.stopWorkTimer()
	mr.closeBreakWindow()
	if mr.appInit && mr.app != nil {
		fyne.Do(func() {
			mr.app.Quit()
		})
	}
	os.Exit(0)
}

func main() {
	workFlag := flag.Duration("work", 25*time.Minute, "Work interval duration (e.g., 25m, 10s)")
	breakFlag := flag.Duration("break", 5*time.Minute, "Break duration (e.g., 5m, 10s)")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging every 1 seconds")
	flag.Parse()

	workInterval = *workFlag
	breakDuration = *breakFlag
	verbose = *verboseFlag

	slog.Info("Move reminder configured", "work_interval", workInterval, "break_duration", breakDuration)
	os.Stdout.Sync()

	if verbose {
		slog.Info("Verbose logging enabled - will log every 1 seconds")
		os.Stdout.Sync()
	}

	reminder := NewMoveReminder()

	// Handle system signals (Cmd+Q) gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Info("Received quit signal - exiting")
		reminder.quit()
		os.Exit(0)
	}()

	reminder.start()
}
