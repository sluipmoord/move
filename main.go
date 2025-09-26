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
	workEnd    time.Time
	workTicker *time.Ticker
}

func NewMoveReminder() *MoveReminder {
	return &MoveReminder{}
}

func (mr *MoveReminder) showBreakWindow() {
	// Show system notification
	mr.showNotification()

	// Launch separate GUI process for break window
	cmd := exec.Command(os.Args[0], "-break-mode", fmt.Sprintf("-break-duration=%s", breakDuration))
	err := cmd.Start()
	if err != nil {
		slog.Error("Failed to start break window process", "error", err)
		return
	}

	// Wait for break process to complete
	err = cmd.Wait()
	if err != nil {
		slog.Info("Break process ended", "error", err)
	}

	slog.Info("Break completed, resuming work")
	mr.scheduleNext()
}

func (mr *MoveReminder) showNotification() {
	// Show system notification on macOS
	title := "Move Break Time!"
	message := "Stand up, stretch, and move around. Take a break from your computer!"
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`display notification "%s" with title "%s"`, message, title))
	err := cmd.Run()
	if err != nil {
		slog.Error("Failed to show notification", "error", err)
	}
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
				mr.showBreakWindow()
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
}

func (mr *MoveReminder) start() {
	slog.Info("Move reminder started", "work_interval", workInterval, "break_duration", breakDuration)
	mr.scheduleNext()

	// Keep the main thread alive
	select {}
}

func main() {
	workFlag := flag.Duration("work", 25*time.Minute, "Work interval duration (e.g., 25m, 10s)")
	breakFlag := flag.Duration("break", 5*time.Minute, "Break duration (e.g., 5m, 10s)")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging every 1 seconds")
	breakMode := flag.Bool("break-mode", false, "Run in break window mode (internal use)")
	breakDurationFlag := flag.Duration("break-duration", 5*time.Minute, "Break duration for break mode")
	flag.Parse()

	workInterval = *workFlag
	breakDuration = *breakFlag
	verbose = *verboseFlag

	// If in break mode, run the GUI break window
	if *breakMode {
		runBreakWindow(*breakDurationFlag)
		return
	}

	// Force logs to be visible by flushing stdout
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
		os.Exit(0)
	}()

	reminder.start()
}

type BreakWindow struct {
	app         fyne.App
	window      fyne.Window
	timerLabel  *widget.Label
	message     *widget.Label
	closeButton *widget.Button
	ticker      *time.Ticker
	breakEnd    time.Time
	timerActive bool
	duration    time.Duration
}

func NewBreakWindow(duration time.Duration) *BreakWindow {
	myApp := app.New()
	myApp.SetIcon(nil)

	bw := &BreakWindow{
		app:      myApp,
		duration: duration,
	}

	return bw
}

func (bw *BreakWindow) showBreakWindow() {
	bw.breakEnd = time.Now().Add(bw.duration)

	// Create break window
	bw.window = bw.app.NewWindow("Move Break")

	// Intercept close attempts - Cmd+Q should always quit
	bw.window.SetCloseIntercept(func() {
		slog.Info("Break window close intercepted - quitting")
		bw.timerActive = false
		if bw.ticker != nil {
			bw.ticker.Stop()
		}
		bw.window.Hide()
		bw.app.Quit()
	})

	title := widget.NewLabel("Time to Move!")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	bw.message = widget.NewLabel("Stand up, stretch, and move around.\nTake a break from your computer!")
	bw.message.Alignment = fyne.TextAlignCenter
	bw.message.Wrapping = fyne.TextWrapWord

	bw.timerLabel = widget.NewLabel("")
	bw.timerLabel.Alignment = fyne.TextAlignCenter
	bw.timerLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Only show close button, initially disabled
	bw.closeButton = widget.NewButton("Close Break", func() {
		bw.closeBreakWindow()
	})
	bw.closeButton.Importance = widget.HighImportance
	bw.closeButton.Disable() // Disabled until break is complete

	// Add keyboard shortcut handler for skip (S key)
	bw.window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyS {
			slog.Info("Break skipped via keyboard shortcut (S key)")
			bw.skipBreak()
		}
	})

	content := container.NewVBox(
		widget.NewSeparator(),
		title,
		widget.NewSeparator(),
		bw.message,
		widget.NewSeparator(),
		bw.timerLabel,
		widget.NewSeparator(),
		bw.closeButton,
		widget.NewSeparator(),
		widget.NewLabel("Press 'S' to skip • Cmd+Q to quit app"),
	)

	bw.window.SetContent(container.NewCenter(content))
	bw.window.SetFullScreen(true)
	bw.window.Show()
	bw.window.RequestFocus()

	// Continuously ensure the window stays focused during break
	go bw.maintainFocus()

	bw.ticker = time.NewTicker(time.Second)
	bw.startTimer()
}

func (bw *BreakWindow) startTimer() {
	bw.timerActive = true
	bw.updateTimer()
}

func (bw *BreakWindow) updateTimer() {
	if !bw.timerActive {
		return // Timer has been stopped
	}

	remaining := time.Until(bw.breakEnd)
	if remaining <= 0 {
		fyne.Do(func() {
			bw.message.SetText("Break time is complete!\nYou can now close this window and return to work.")
			bw.timerLabel.SetText("00:00")
			bw.closeButton.SetText("Return to Work")
			bw.closeButton.Enable() // Enable the button only when break is complete
		})
		return
	}

	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60
	fyne.Do(func() {
		bw.timerLabel.SetText(fmt.Sprintf("Time remaining: %02d:%02d", minutes, seconds))
	})

	if bw.timerActive {
		time.AfterFunc(time.Second, bw.updateTimer)
	}
}

func (bw *BreakWindow) skipBreak() {
	slog.Info("Skipping break via keyboard shortcut")
	bw.timerActive = false
	if bw.ticker != nil {
		bw.ticker.Stop()
	}

	// Hide the break window
	if bw.window != nil {
		bw.window.Hide()
	}

	slog.Info("Break skipped, exiting break window")
	bw.app.Quit()
}

func (bw *BreakWindow) bringToFront() {
	// Use AppleScript to bring our app to the front on macOS
	cmd := exec.Command("osascript", "-e", `tell application "System Events" to set frontmost of first process whose name contains "main" to true`)
	err := cmd.Run()
	if err != nil {
		slog.Error("Failed to bring window to front", "error", err)
	}
}

func (bw *BreakWindow) maintainFocus() {
	// Keep bringing the break window to front during mandatory break
	focusTicker := time.NewTicker(1 * time.Second)
	defer focusTicker.Stop()

	for range focusTicker.C {
		if !bw.timerActive {
			break
		}

		remaining := time.Until(bw.breakEnd)
		if remaining <= 0 {
			break
		}

		// Continuously request focus to keep window active
		if bw.window != nil {
			fyne.Do(func() {
				bw.window.RequestFocus()
			})
			// Also try to bring to front via system command
			bw.bringToFront()
		}
	}
}

func (bw *BreakWindow) closeBreakWindow() {
	slog.Info("Closing break window")

	// Stop all timers immediately
	bw.timerActive = false
	if bw.ticker != nil {
		bw.ticker.Stop()
	}

	// Disable the button to prevent double-clicks
	bw.closeButton.Disable()

	// Hide the break window
	if bw.window != nil {
		bw.window.Hide()
	}

	slog.Info("Break window closed, exiting")
	bw.app.Quit()
}

func (bw *BreakWindow) start() {
	bw.showBreakWindow()
	bw.app.Run()
}

func runBreakWindow(duration time.Duration) {
	breakWindow := NewBreakWindow(duration)
	breakWindow.start()
}
