# Move Reminder

A Go program that helps maintain healthy movement habits by showing fullscreen break reminders at customizable intervals.

## Features

- Shows fullscreen break reminders at configurable intervals (default: 25 minutes)
- Displays a movement encouragement message
- Shows a countdown timer during breaks (default: 5 minutes)
- Provides "Close Break" and "Skip Break" buttons for user control
- Automatically resumes work timer when break window is closed
- Customizable work and break durations via command-line flags

## How it works

1. The program starts and schedules the first break based on your work interval
2. When break time arrives:
   - A fullscreen window appears with a movement reminder
   - A countdown timer is displayed for your configured break duration
   - You can close the break early with the "Close Break" button
   - Or skip the break entirely with "Skip Break"
3. After the break ends (timer runs out or you close it), the next work cycle begins
4. This cycle repeats indefinitely until you quit the program

## Usage

### Default Settings (25min work / 5min break)

```bash
# Build and run with defaults
go build -o main main.go
./main

# Or run directly
go run main.go
```

### Custom Intervals

```bash
# Custom work and break intervals
go run main.go -work=30m -break=10m

# For testing with short intervals
go run main.go -work=10s -break=5s

# View available options
go run main.go -help
```

### Flag Options

- `-work`: Work interval duration (default: 25m)
  - Examples: `30m`, `1h`, `45m30s`
- `-break`: Break duration (default: 5m)  
  - Examples: `10m`, `2m30s`, `15s`

## Requirements

- macOS (uses system screen locking)
- Go 1.25.1 or later
- Fyne GUI framework (automatically installed via go mod)

## Dependencies

The program uses the Fyne v2 GUI framework for creating the fullscreen reminder window. All dependencies are managed through Go modules and will be automatically downloaded when you build the program.

## System Compatibility

This program works on any operating system that supports the Fyne GUI framework (Windows, macOS, Linux). No system-specific screen locking is used - it simply shows a fullscreen reminder window.