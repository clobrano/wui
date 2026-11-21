# VHS tape for the README demo (https://github.com/charmbracelet/vhs).
#
# Regenerate with `make demo` from the repository root. It runs wui on the
# throwaway tasks from docs/demo/seed.sh, never on your own data.

Output docs/demo/wui.gif

Require task

Set Shell bash
Set FontSize 16
Set Width 1200
Set Height 640
Set Padding 20
Set Framerate 24
Set TypingSpeed 60ms
Set Theme "Catppuccin Mocha"

# Fake $HOME with sample tasks, and the freshly built ./wui on PATH.
Hide
Type "export HOME=/tmp/wui-demo PATH=$PWD:$PATH PS1='$ ' && docs/demo/seed.sh $HOME && clear"
Enter
Sleep 1s
Show

Type "wui"
Sleep 300ms
Enter
Sleep 2s

# Move around the Next tab
Down@400ms 3
Sleep 600ms
Type "g"
Sleep 800ms

# Task details
Enter
Sleep 3s
Escape
Sleep 800ms

# Browse by project and drill into one
Tab@700ms 2
Sleep 1.5s
Down
Sleep 500ms
Enter
Sleep 2.5s

# Back to Next and narrow it down with Taskwarrior syntax
Type "2"
Sleep 800ms
Type "/"
Sleep 600ms
Type "project:home"
Sleep 600ms
Enter
Sleep 2.5s

# Switching tabs restores the tab's own filter
Tab
Sleep 600ms
Shift+Tab
Sleep 1.5s

# Add a task
Type "n"
Sleep 600ms
Type "Prepare slides for the meetup project:wui due:thursday +talk"
Sleep 600ms
Enter
Sleep 2.5s

# Quick modify the new task: it moves up the urgency-sorted list
Down@300ms 4
Sleep 800ms
Type "m"
Sleep 500ms
Type "priority:H due:today"
Sleep 500ms
Enter
Sleep 2.5s

# Mark the top task done, then undo
Type "g"
Sleep 600ms
Type "d"
Sleep 2s
Type "u"
Sleep 2.5s

Type "q"
Sleep 500ms
