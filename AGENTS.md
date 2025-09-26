# Agent Guidelines for Move Reminder

Don't use pkill -f main to terminate the app; instead, use the Quit option in the system tray menu to ensure proper cleanup and resource management.

## Build Commands

- `go build` - Build the main application  
- `go build -o main main.go` - Build main application with specific output
- `go run main.go` - Run main application directly

## Testing

- `go test` - Run tests (currently no *_test.go files exist)
- `go test -v` - Run tests with verbose output
- `go test -run TestName` - Run specific test

## Code Style Guidelines

- **Package**: Single main package for this simple GUI application
- **Imports**: Group standard library first, then third-party (fyne.io), use tabs for alignment
- **Naming**: CamelCase for exported types (MoveReminder), camelCase for unexposed (lockScreen)
- **Constants**: ALL_CAPS with underscores (workInterval, breakDuration)
- **Struct fields**: Exported fields use CamelCase, unexported use camelCase
- **Logger**: Use slog package for logging
- **Error handling**: Check errors with `if err != nil` pattern, log with `slog.Error`
- **Receivers**: Use short names (mr for MoveReminder), pointer receivers for methods that modify state
- **Comments**: Minimal, only where necessary for clarity

## Dependencies

- Uses Fyne v2 GUI framework (fyne.io/fyne/v2)
- macOS specific enhancements via AppleScript for modal behavior and notifications

## Documentation Guidelines

### README.md Updates

When significant features are added or changed, update README.md to reflect:

- **Feature descriptions**: Update the Features section with new capabilities
- **Usage examples**: Add command-line examples for new flags or options  
- **Behavior changes**: Document any changes to how the application works
- **System requirements**: Update if new system dependencies are added
- **User instructions**: Include new keyboard shortcuts, controls, or workflows

### Key Sections to Maintain

- **Features**: Keep comprehensive list of current capabilities
- **How it works**: Update workflow descriptions when behavior changes
- **Usage**: Include all command-line options and examples
- **During Breaks**: Document user options and controls available during breaks
- **Requirements**: Keep system requirements and dependencies current

### Documentation Style

- Use clear, concise descriptions
- Include code examples for command-line usage
- Use emojis sparingly for section headers (🎯, ⚙️, ✅, ❌)
- Maintain consistent formatting with existing style
- Focus on user benefits and practical usage
