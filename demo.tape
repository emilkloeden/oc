Output docs/demo.gif

Set FontSize 14
Set Width 900
Set Height 450
Set Theme "Catppuccin Mocha"
Set Shell "bash"
Set Padding 20
Set TypingSpeed 80ms

Hide
# Clean prompt and temp working directory
Type "export PS1='$ '"
Enter
Sleep 200ms
Type "cd $(mktemp -d)"
Enter
Sleep 300ms
Type "clear"
Enter
Sleep 200ms
Show

Type "oc new my_app"
Enter
Sleep 3s

Type "cd my_app"
Enter
Sleep 500ms

Type "oc add yojson"
Enter
Sleep 6s

Type "oc run"
Enter
Sleep 6s

Sleep 4s
