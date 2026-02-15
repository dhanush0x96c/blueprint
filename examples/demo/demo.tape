# VHS demo tape for Blueprint CLI
# Run: vhs examples/demo/demo.tape

Output examples/demo/demo.gif

Set Shell bash
Set FontSize 16
Set Width 900
Set Height 550
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set PlaybackSpeed 1

# Ensure clean output directory
Hide
Type "cd $(mktemp -d) && clear"
Enter
Sleep 1s
Show

# Run blueprint init
Sleep 500ms
Type "blueprint init go-cli"
Sleep 500ms
Enter
Sleep 2s

# Feature selection prompt: select features/go/testing with Space, then confirm with Enter
Space
Sleep 1.5s
Enter
Sleep 1.5s

# Variable form - all 3 fields shown at once
# Field 1: "What is your application name?"
Type "wild-idea"
Sleep 500ms

# Move to next field
Tab
Sleep 500ms

# Field 2: "What is your module path?"
Type "github.com/example/wild-idea"
Sleep 500ms

# Move to next field (description already has default)
Tab
Sleep 500ms

# Field 3: "Brief description of your application" - clear default and type custom
Ctrl+U
Backspace
Type "A wild idea brought to life"
Sleep 1s
Enter
Sleep 1s

# Testing framework prompt: select Yes with Left arrow, then confirm
Left
Sleep 1s
Enter
Sleep 3s

Type "clear"
Sleep 500ms
Enter
Sleep 500ms

# Show the generated project structure
Type "eza --tree wild-idea"
Sleep 500ms
Enter
Sleep 3s

# Show a generated file
Type "bat wild-idea/main.go"
Sleep 500ms
Enter
Sleep 5s

