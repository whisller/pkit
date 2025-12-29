# VHS tape file for creating pkit demo
# Install VHS: go install github.com/charmbracelet/vhs@latest
# Run: vhs docs/demo.tape

Output docs/demo.gif

Require pkit

Set Shell "bash"
Set FontSize 14
Set Width 1400
Set Height 800
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 100ms
Set PlaybackSpeed 0.8

# Title
Type "# pkit - Multi-source AI Prompt Manager"
Sleep 1s
Enter
Sleep 500ms

# Subscribe to a source
Type "pkit subscribe fabric/patterns"
Sleep 500ms
Enter
Sleep 3s

# Interactive finder
Type "pkit find"
Sleep 500ms
Enter
Sleep 2s

# Type search query
Type "code review"
Sleep 2s

# Navigate through results
Down 2
Sleep 500ms
Down 1
Sleep 500ms

# Select a prompt
Enter
Sleep 2s

# Exit TUI
Escape
Sleep 1s

# Show bookmark usage
Type "pkit bookmark add fabric:code-review --alias review"
Sleep 500ms
Enter
Sleep 2s

# Show piping example
Type "echo 'func main() { }' | pkit review | head -n 5"
Sleep 500ms
Enter
Sleep 2s

# Final message
Sleep 1s
Type "# Try it yourself: github.com/whisller/pkit"
Sleep 2s
