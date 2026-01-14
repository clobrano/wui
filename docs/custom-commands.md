# Custom Commands

Custom commands allow you to execute system commands with task data using a flexible templating system.

## Configuration

Add custom commands to your `~/.config/wui/config.yaml`:

```yaml
tui:
  custom_commands:
    o:  # Key to press
      name: "Open URL"
      command: "xdg-open {{.url}}"
      description: "Opens the task's URL field in the default browser"

    "1":
      name: "Copy Description"
      command: "echo {{.description}} | xclip -selection clipboard"
      description: "Copies task description to clipboard"

    b:
      name: "Open in Browser"
      command: "firefox {{.url}}"
      description: "Opens URL in Firefox"
```

## Template Syntax

Use `{{.fieldname}}` to insert task field values into commands. All fields accessible via taskwarrior are supported:

### Standard Fields
- `{{.id}}` - Task ID
- `{{.uuid}}` - Task UUID (full)
- `{{.short_uuid}}` - Task UUID (first 8 characters)
- `{{.description}}` - Task description
- `{{.project}}` - Project name
- `{{.priority}}` - Priority (H/M/L)
- `{{.status}}` - Status (pending/completed/etc)
- `{{.tags}}` - Tags (comma-separated)
- `{{.due}}` - Due date
- `{{.scheduled}}` - Scheduled date
- `{{.urgency}}` - Urgency score

### User-Defined Attributes (UDAs)
Any custom field you've defined in taskwarrior:
- `{{.url}}` - URL field (if defined as UDA)
- `{{.github}}` - GitHub field (if defined)
- `{{.contact}}` - Contact field (if defined)

## Platform-Specific Examples

### Linux

```yaml
tui:
  custom_commands:
    o:
      name: "Open URL"
      command: "xdg-open {{.url}}"
      description: "Open URL in default browser"

    c:
      name: "Copy to Clipboard"
      command: "echo '{{.description}}' | xclip -selection clipboard"
      description: "Copy description to clipboard"
```

### Linux (Termux)

```yaml
tui:
  custom_commands:
    o:
      name: "Open URL"
      command: "/data/data/com.termux/files/usr/bin/termux-open-url {{.url}}"
      description: "Open URL in browser"

    s:
      name: "Share Task"
      command: "/data/data/com.termux/files/usr/bin/termux-share -a send '{{.description}}'"
      description: "Share task via Android"
```

**Termux Note:** If you get "cannot open" errors, use full paths to Termux commands. Find the path with `which termux-open-url`.

### macOS

```yaml
tui:
  custom_commands:
    o:
      name: "Open URL"
      command: "open {{.url}}"
      description: "Open URL in default browser"

    c:
      name: "Copy to Clipboard"
      command: "echo '{{.description}}' | pbcopy"
      description: "Copy description to clipboard"
```

### Windows

```yaml
tui:
  custom_commands:
    o:
      name: "Open URL"
      command: "cmd /c start {{.url}}"
      description: "Open URL in default browser"
```

## Advanced Examples

### Multiple Fields

```yaml
tui:
  custom_commands:
    g:
      name: "Git Clone"
      command: "git clone {{.url}} ~/projects/{{.project}}"
      description: "Clone repository from URL to project folder"
```

### Complex Commands

```yaml
tui:
  custom_commands:
    n:
      name: "Create Note"
      command: "echo '# {{.description}}' > ~/notes/{{.id}}.md && vim ~/notes/{{.id}}.md"
      description: "Create markdown note from task"

    f:
      name: "Format for Notes"
      command: "sh -c \"echo '{{.description}} ({{.short_uuid}})' | xclip -selection clipboard\""
      description: "Copy task description with short UUID to clipboard"
```

### Quoted Arguments

When task fields contain spaces, use quotes in your command:

```yaml
tui:
  custom_commands:
    e:
      name: "Email Task"
      command: "mutt -s '{{.description}}' -- recipient@example.com"
      description: "Email task details"
```

## Usage

1. Navigate to a task
2. Press the configured key (e.g., `o`)
3. The command executes with task data substituted
4. Status message shows success/failure

Press `?` in wui to see your configured custom commands in the help screen.

## Error Handling

wui provides comprehensive error reporting when custom commands fail:

### Error Types

- **"Field not found"** - The field doesn't exist in the task (e.g., referencing `{{.url}}` when task has no URL field)
- **"No task selected"** - You're not on a task (e.g., in group view)
- **"Command parsing failed"** - Invalid command syntax (e.g., unterminated quotes)
- **"Command failed (exit code N)"** - Command ran but exited with non-zero status, includes:
  - Exit code number
  - Error output (stderr) from the command

### Examples

```
✗ Command 'Open URL' failed (exit code 127): command not found: xdg-open
✗ Command expansion failed: field 'url' not found in task
✗ Command parsing failed: unterminated quote in command
✗ Command 'Git Clone' failed (exit code 128): fatal: repository not found
```

### What Gets Reported

When a command fails:
1. **Exit code**: Shows the numeric exit status (if the command ran)
2. **Error message**: Shows stderr output from the command
3. **Context**: Shows which custom command failed by name

## Shell Features (Pipes, Redirects, etc.)

Custom commands are executed directly without invoking a shell. This means shell features like pipes (`|`), redirections (`>`), and command substitution (`$()`) won't work by default.

### Using Shell Features

To use shell features, wrap your command with `sh -c "..."`:

```yaml
tui:
  custom_commands:
    c:
      name: "Copy to Clipboard"
      command: "sh -c \"echo '{{.description}}' | xclip -selection clipboard\""
      description: "Pipe output to clipboard"

    l:
      name: "Log Task"
      command: "sh -c \"echo '{{.description}}' >> ~/tasks.log\""
      description: "Redirect output to log file"
```

**Note**: Without `sh -c`, the pipe or redirect operators will be treated as literal arguments, not as shell operators.

## Tips

1. **Test commands first**: Run commands manually in your terminal before adding to config
2. **Check field names**: Use `task <id> export` to see available fields
3. **Platform detection**: Create different configs per platform or use a script wrapper
4. **Security**: Be careful with commands that modify or delete data
5. **Quotes**: Use single quotes `'` to protect special characters in shell commands
6. **Shell features**: Wrap commands needing pipes or redirects with `sh -c "..."`
