# Agent Guidelines for Move Reminder

## Build Commands
- `go build` - Build the main application  
- `go build -o main main.go` - Build main application with specific output
- `go build -o test-main test-main.go` - Build test version with shortened intervals
- `go run main.go` - Run main application directly
- `go run test-main.go` - Run test version directly

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
- macOS specific screen locking via CGSession command