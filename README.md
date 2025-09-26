# Move Reminder

A Go program that helps maintain healthy movement habits by showing **mandatory fullscreen break reminders** at customizable intervals with true modal behavior to ensure you actually take breaks.

## Features

### 🎯 Core Functionality

- **Mandatory Break System**: Break windows cannot be closed during the timer countdown
- **Modal Break Windows**: Continuously brings break window to front, preventing work on other applications
- **Work Time Logging**: Logs remaining work time every 30 seconds (10 seconds with `-verbose`)
- **System Notifications**: Shows native macOS notifications when breaks start
- **Fullscreen Break Display**: Shows encouraging movement messages with countdown timer

### 🎛️ User Controls

- **Multiple Exit Options**:
  - Wait for break timer to complete → "Return to Work" button becomes available
  - Press 'S' key → Skip break immediately (hidden shortcut)
  - Press Cmd+Q → Quit entire application
- **Smart Button Management**: Prevents double-clicks and UI confusion
- **Clear User Feedback**: Shows available options when break window close is attempted

### ⚙️ Customization

- **Configurable Intervals**: Custom work and break durations via command-line flags
- **Verbose Logging**: Optional detailed work time tracking
- **Flexible Duration Format**: Supports minutes, seconds, hours (e.g., `25m`, `30s`, `1h15m`)

## How it works

1. **Work Phase**: The program logs remaining work time and schedules the next break
2. **Break Trigger**: When break time arrives:
   - System notification appears
   - Fullscreen modal window takes over the screen
   - Window continuously maintains focus to prevent working
   - Countdown timer shows remaining break time
3. **Break Enforcement**:
   - "Return to Work" button is disabled until timer completes
   - Window close (X button) is blocked with helpful message
   - User can skip with 'S' key or quit app with Cmd+Q
4. **Break Completion**: Window closes and next work cycle begins automatically
5. **Cycle Repeats**: Process continues indefinitely until user quits

## Break Window Behavior

The break window is designed to be **truly interrupting** to ensure effective breaks:

- ✅ **Fullscreen and modal** - takes over entire screen
- ✅ **Continuous focus stealing** - brings window to front every second
- ✅ **System-level prominence** - uses AppleScript to maintain visibility
- ✅ **Close protection** - cannot be dismissed early during timer
- ✅ **Clear messaging** - shows available options when close is attempted

## Usage

### Quick Start

```bash
# Build and run with defaults (25min work / 5min break)
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

# Enable verbose work time logging (every 1 second)
go run main.go -verbose

# Combine flags
go run main.go -work=20m -break=3m -verbose
```

### Command-Line Options

- `-work=<duration>`: Work interval duration (default: `25m`)
  - Examples: `30m`, `1h`, `45m30s`, `90m`
- `-break=<duration>`: Break duration (default: `5m`)  
  - Examples: `10m`, `2m30s`, `30s`, `15m`
- `-verbose`: Enable detailed work time logging every 1 second (default: every 30 seconds)
- `-help`: Show all available options

### During Breaks

When a break window appears, you have these options:

- **Wait it out and Move**: Let the timer count down, then click "Return to Work"
- **Skip break**: Press the 'S' key to skip immediately
- **Quit app**: Press Cmd+Q to quit the entire application
- **❌ Cannot close**: The X button is disabled during mandatory break timer

## Requirements

- **macOS** (enhanced modal behavior uses AppleScript for window management)
- **Go 1.25.1** or later
- **Fyne GUI framework** (automatically installed via go mod)

## Installation & Building

```bash
# Clone or download the project
cd move-reminder

# Build the application
go build -o main main.go

# Or use the build script
chmod +x build.sh
./build.sh

# Run with default settings
./main
```

## Dependencies

The program uses:

- **Fyne v2** GUI framework for fullscreen windows and UI components
- **Standard Go libraries** for timing, logging, and system commands
- **macOS system commands** (AppleScript) for enhanced window focus management

All dependencies are managed through Go modules and automatically downloaded during build.

## System Features

### macOS Enhancements

- **System notifications** when breaks start
- **AppleScript integration** for window focus management
- **Native window behavior** with proper modal enforcement

### Cross-Platform Base

The core functionality works on any OS supporting Fyne (Windows, macOS, Linux), but the enhanced modal behavior and notifications are optimized for macOS.

## Development

### Testing with Short Intervals

```bash
# Quick testing (3s work, 2s break, verbose logging)
./main -work=3s -break=2s -verbose
```

### Build Options

```bash
# Standard build
go build -o main main.go

# Build for testing with shorter logging intervals  
go build -o test-main main.go
```
